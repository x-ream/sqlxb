# Qdrant nil/0 过滤和 JOIN 查询

## 📋 用户问题

1. **在转换 JSON 之前，有没有完成 nil/0 过滤？**
2. **Qdrant 有类似 JOIN 查询吗？**

---

## ✅ 问题 1: nil/0 过滤 - 已完成

### 答案：是的，已完成！

**nil/0 过滤在构建 Bb 阶段就完成了**，不是在转换 JSON 时。

---

### 过滤实现位置

```go
// cond_builder.go 第 95-123 行
func (cb *CondBuilder) doGLE(p string, k string, v interface{}) *CondBuilder {

	switch v.(type) {
	case string:
		if v.(string) == "" {
			return cb  // ⭐ 空字符串被过滤，不添加到 bbs
		}
	
	case uint64, uint, int64, int, int32, int16, int8, bool, byte, float64, float32:
		if v == 0 {
			return cb  // ⭐ 0 被过滤
		}
	
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return cb  // ⭐ nil 被过滤
		}
		return cb.addBb(p, k, n)
	
	default:
		if v == nil {
			return cb  // ⭐ nil 被过滤
		}
	}
	
	return cb.addBb(p, k, v)  // ⭐ 只有非 nil/0 的值才会添加
}
```

---

### 数据流

```
用户代码
  ↓
xb.Of(&CodeVector{}).
    Eq("language", "golang").      // ✅ 有效值
    Eq("category", "").            // ⭐ 空字符串
    Gt("score", 0.8).              // ✅ 有效值
    Gt("rank", 0)                  // ⭐ 值为 0
  ↓
doGLE() 处理每个条件
  ↓
  Eq("language", "golang")  → addBb() ✅
  Eq("category", "")        → return cb (不添加) ❌
  Gt("score", 0.8)          → addBb() ✅
  Gt("rank", 0)             → return cb (不添加) ❌
  ↓
cb.bbs = [
    Bb{op: EQ, key: "language", value: "golang"},
    Bb{op: GT, key: "score", value: 0.8}
]
  ↓
Build()
  ↓
JsonOfSelect()
  ↓
只转换有效的 2 个条件 ✅
```

---

### 验证测试

```go
// 测试：包含 nil/0 的查询
built := xb.Of(&CodeVector{}).
    Eq("language", "golang").      // ✅
    Eq("category", "").            // ⭐ 被过滤
    Gt("score", 0.8).              // ✅
    Gt("rank", 0).                 // ⭐ 被过滤
    Lt("complexity", 100).         // ✅
    VectorSearch("embedding", vec, 20).
    Build()

json, _ := built.JsonOfSelect()
```

**输出**：

```json
{
  "vector": [0.1, 0.2, 0.3],
  "limit": 20,
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}},
      {"key": "score", "range": {"gt": 0.8}},
      {"key": "complexity", "range": {"lt": 100}}
    ]
  }
}
```

**验证结果**：✅ 只有 3 个有效条件，`category=""` 和 `rank=0` 被自动过滤

---

### 测试结果

```bash
$ go test -v -run TestQdrant_NilZeroFilter

=== RUN   TestQdrant_NilZeroFilter
    ✅ nil/0 过滤成功：只有 3 个有效条件
--- PASS: TestQdrant_NilZeroFilter (0.00s)

$ go test -v -run TestPostgreSQL_NilZeroFilter

=== RUN   TestPostgreSQL_NilZeroFilter
    SQL: SELECT ... WHERE language = ? AND score > ? ...
    Args: [queryVector, "golang", 0.8]
    ✅ PostgreSQL nil/0 过滤成功
--- PASS: TestPostgreSQL_NilZeroFilter (0.00s)
```

---

### 支持的过滤规则

| 类型 | 被过滤的值 | 示例 |
|------|-----------|------|
| `string` | `""` (空字符串) | `Eq("name", "")` → 被忽略 |
| `int`, `int64`, `int32`, ... | `0` | `Gt("count", 0)` → 被忽略 |
| `float32`, `float64` | `0` | `Lt("score", 0)` → 被忽略 |
| `*int`, `*string`, ... (指针) | `nil` | `Eq("name", nil)` → 被忽略 |
| 任意类型 | `nil` | `Eq("obj", nil)` → 被忽略 |

---

## ❌ 问题 2: Qdrant 的 JOIN 查询

### 答案：不支持传统 JOIN

**Qdrant 不支持传统 SQL 的 JOIN 操作**，因为它是专为向量相似度搜索设计的，不是关系型数据库。

---

### Qdrant 的设计理念

```
关系型数据库 (MySQL, PostgreSQL):
  ├── 强项: JOIN, 事务, 关系约束
  └── 弱项: 向量相似度搜索

向量数据库 (Qdrant, Milvus):
  ├── 强项: 高效向量检索, 大规模相似度搜索
  └── 弱项: JOIN, 复杂关系查询
```

---

### 替代方案

#### 方案 1: Payload 嵌入（推荐）⭐

**适用场景**：一对多关系，数据量不大

```json
// Qdrant 数据模型
{
  "id": 123,
  "vector": [0.1, 0.2, 0.3],
  "payload": {
    "content": "func login() { ... }",
    "language": "golang",
    
    // ⭐ 嵌入关联数据
    "author": {
      "id": 456,
      "name": "张三",
      "department": "后端组"
    },
    
    "tags": ["authentication", "security", "jwt"]
  }
}
```

**xb 查询**：

```go
// 一次查询获取所有数据
built := xb.Of(&CodeVector{}).
    Eq("language", "golang").
    Eq("author.department", "后端组").  // ⭐ 嵌套过滤
    VectorSearch("embedding", queryVector, 20).
    Build()

json, _ := built.JsonOfSelect()
```

**Qdrant JSON**：

```json
{
  "vector": [0.1, 0.2, 0.3],
  "limit": 20,
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}},
      {"key": "author.department", "match": {"value": "后端组"}}
    ]
  }
}
```

**优点**：
- ✅ 一次查询获取所有数据
- ✅ 查询速度快
- ✅ 无需二次查询

**缺点**：
- ❌ 数据冗余（author 信息重复存储）
- ❌ 更新复杂（需要更新所有相关记录）

---

#### 方案 2: 两阶段查询（应用层 JOIN）

**适用场景**：多对多关系，数据量大

```go
// 第 1 阶段：查询代码向量
codeResults := xb.Of(&CodeVector{}).
    VectorSearch("embedding", queryVector, 20).
    Build().
    QueryQdrant()  // 假设的方法

// 提取作者 ID
authorIDs := []int64{}
for _, code := range codeResults {
    authorIDs = append(authorIDs, code.AuthorID)
}

// 第 2 阶段：批量查询作者信息（从 MySQL 或另一个集合）
authors := xb.Of(&Author{}).
    In("id", authorIDs...).
    Build().
    Query()

// 第 3 阶段：应用层关联
codeWithAuthors := []CodeWithAuthor{}
for _, code := range codeResults {
    for _, author := range authors {
        if code.AuthorID == author.ID {
            codeWithAuthors = append(codeWithAuthors, CodeWithAuthor{
                Code:   code,
                Author: author,
            })
            break
        }
    }
}
```

**优点**：
- ✅ 无数据冗余
- ✅ 更新简单
- ✅ 灵活性高

**缺点**：
- ❌ 需要多次查询
- ❌ 应用层需要手动关联
- ❌ 性能开销较大

---

#### 方案 3: 混合架构（推荐）⭐⭐⭐

**适用场景**：生产环境

```
向量检索（Qdrant）+ 关系查询（PostgreSQL/MySQL）

┌─────────────────────────────────────────┐
│         用户查询                         │
│  "找相似的 Golang 后端组的代码"          │
└──────────────┬──────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│  Step 1: Qdrant 向量检索                 │
│  - 找到相似代码的 ID                     │
│  - 返回: [id: 1, 2, 3, ..., 20]         │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│  Step 2: PostgreSQL 关系查询             │
│  SELECT c.*, a.name, a.department        │
│  FROM codes c                            │
│  JOIN authors a ON c.author_id = a.id    │
│  WHERE c.id IN (1,2,3,...,20)            │
│    AND a.department = '后端组'           │
└──────────────┬───────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│  返回结果                                │
│  - 相似代码 + 作者信息                   │
└──────────────────────────────────────────┘
```

**实现代码**：

```go
// Step 1: Qdrant 向量检索
built := xb.Of(&CodeVector{}).
    VectorSearch("embedding", queryVector, 100).
    Build()

qdrantResults := qdrantClient.Search(built.JsonOfSelect())

// 提取 ID
codeIDs := []int64{}
for _, result := range qdrantResults {
    codeIDs = append(codeIDs, result.ID)
}

// Step 2: PostgreSQL 关系查询（带 JOIN）
// 使用原生 SQL 或 sqlx
query := `
    SELECT c.*, a.name, a.department
    FROM codes c
    JOIN authors a ON c.author_id = a.id
    WHERE c.id IN (?) AND a.department = ?
`
results := db.Query(query, codeIDs, "后端组")
```

**优点**：
- ✅ 充分利用 Qdrant 的向量检索能力
- ✅ 充分利用 PostgreSQL 的关系查询能力
- ✅ 无数据冗余
- ✅ 查询灵活

**缺点**：
- ⚠️ 需要维护两个数据库
- ⚠️ 数据同步复杂
- ⚠️ 架构复杂度增加

---

### Qdrant 的最佳实践

```
场景 1: 简单查询（无关联）
  → 纯 Qdrant ✅

场景 2: 一对多关系（小数据量）
  → Qdrant Payload 嵌入 ✅

场景 3: 多对多关系（大数据量）
  → 混合架构（Qdrant + PostgreSQL）✅⭐

场景 4: 复杂关系查询
  → PostgreSQL pgvector（牺牲向量检索性能）
  → 或混合架构 ✅⭐
```

---

## 🎯 xb 的优势

### 优势 1: 统一 API

```go
// 相同的代码

builder := xb.Of(&CodeVector{}).
    Eq("language", "golang").
    Eq("category", "").            // ⭐ 自动过滤
    VectorSearch("embedding", vec, 20)

// PostgreSQL
sql, args := builder.Build().SqlOfVectorSearch()
// nil/0 已过滤 ✅

// Qdrant
json, _ := builder.Build().JsonOfSelect()
// nil/0 已过滤 ✅
```

---

### 优势 2: 自动过滤

```
用户不需要关心 nil/0 过滤：

❌ 手动过滤（繁琐）:
if language != "" {
    builder.Eq("language", language)
}
if score > 0 {
    builder.Gt("score", score)
}

✅ xb 自动过滤（简洁）:
builder.
    Eq("language", language).  // 空字符串自动忽略
    Gt("score", score)         // 0 自动忽略
```

---

### 优势 3: 多后端兼容

```
一份代码，多种后端：

// 开发环境: PostgreSQL + pgvector
sql, args := built.SqlOfVectorSearch()

// 生产环境: Qdrant
json, _ := built.JsonOfSelect()

// 未来: Milvus, Weaviate, ...
// 只需添加 ToMilvusJSON(), ToWeaviateJSON()
```

---

## 📝 总结

### 问题 1: nil/0 过滤

✅ **已完成**  
✅ **过滤在构建 Bb 阶段**  
✅ **PostgreSQL 和 Qdrant 都受益**  
✅ **用户无需手动处理**

---

### 问题 2: JOIN 查询

❌ **Qdrant 不支持 JOIN**  
✅ **替代方案 1: Payload 嵌入（简单场景）**  
✅ **替代方案 2: 两阶段查询（复杂场景）**  
✅ **替代方案 3: 混合架构（推荐生产环境）**

**推荐架构**：

```
向量检索 → Qdrant
关系查询 → PostgreSQL
统一 API → xb
```

---

## 🚀 下一步

### 可能的扩展

1. **支持更多向量数据库**：
   - `ToMilvusJSON()`
   - `ToWeaviateJSON()`
   - `ToPineconeJSON()`

2. **混合架构助手**：
   - `WithPostgreSQLJoin()` - 先 Qdrant 再 PostgreSQL
   - 自动数据同步

3. **更强大的 Payload 支持**：
   - 嵌套字段查询：`Eq("author.department", "后端组")`
   - 数组字段查询：`In("tags", "security")`

---

**xb：一个 API，多种后端，AI-First！** 🎉

