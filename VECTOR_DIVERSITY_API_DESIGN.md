# sqlxb 向量多样性 API 设计

## 📋 设计目标

根据用户需求：

1. ✅ **设计 API，支持 Qdrant**
2. ✅ **如果是其他数据库且不支持，不能报错，要忽略**（优雅降级）
3. ✅ **转 Qdrant 需要的 JSON**

---

## 🎯 核心设计原则

### 1. 优雅降级（Graceful Degradation）

```go
// 相同代码，不同后端

builder := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 20).
    WithHashDiversity("semantic_hash")  // ⭐ 多样性参数

built := builder.Build()

// PostgreSQL: 自动忽略多样性 ✅
sql, args := built.SqlOfVectorSearch()
// SQL: ... LIMIT 20  (不是 100)

// Qdrant: 应用多样性 ✅
json, _ := built.ToQdrantJSON()
// limit: 100 (20 * 5 倍)
```

**关键实现**：

```go
// to_vector_sql.go (PostgreSQL)
func (built *Built) SqlOfVectorSearch() (string, []interface{}) {
    vectorBb := findVectorSearchBb(built.Conds)
    params := vectorBb.value.(VectorSearchParams)
    
    // ⭐ 忽略 Diversity，使用原始 TopK
    limit := params.TopK  
    // 而不是 params.TopK * params.Diversity.OverFetchFactor
    
    sql := fmt.Sprintf("... LIMIT %d", limit)
    return sql, args
}

// to_qdrant_json.go (Qdrant)
func (built *Built) ToQdrantRequest() (*QdrantSearchRequest, error) {
    params := vectorBb.value.(VectorSearchParams)
    
    req := &QdrantSearchRequest{
        Limit: params.TopK,
    }
    
    // ⭐ 应用 Diversity，扩大 limit
    if params.Diversity != nil && params.Diversity.Enabled {
        factor := params.Diversity.OverFetchFactor
        if factor <= 0 {
            factor = 5
        }
        req.Limit = params.TopK * factor  // 20 * 5 = 100
    }
    
    return req, nil
}
```

---

### 2. 类型安全

```go
// 多样性策略（类型安全的枚举）
type DiversityStrategy string

const (
    DiversityByHash     DiversityStrategy = "hash"
    DiversityByDistance DiversityStrategy = "distance"
    DiversityByMMR      DiversityStrategy = "mmr"
)

// 多样性参数（结构化配置）
type DiversityParams struct {
    Enabled         bool
    Strategy        DiversityStrategy
    HashField       string   // for DiversityByHash
    MinDistance     float32  // for DiversityByDistance
    Lambda          float32  // for DiversityByMMR
    OverFetchFactor int      // 过度获取因子
}
```

---

### 3. 链式 API（Fluent API）

```go
// 通用方法
WithDiversity(strategy DiversityStrategy, params ...interface{}) *BuilderX

// 快捷方法（语法糖）
WithHashDiversity(hashField string) *BuilderX
WithMinDistance(minDistance float32) *BuilderX
WithMMR(lambda float32) *BuilderX

// 示例
sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    WithHashDiversity("semantic_hash").  // ⭐ 链式调用
    Build()
```

---

## 🏗️ 架构设计

### 文件结构

```
sqlxb/
├── vector_types.go           # 向量类型定义
│   ├── Vector
│   ├── VectorDistance
│   ├── DiversityStrategy      ⭐ 新增
│   └── DiversityParams        ⭐ 新增
│
├── cond_builder_vector.go    # 向量查询构建器
│   ├── VectorSearch()
│   ├── VectorDistance()
│   ├── WithDiversity()        ⭐ 新增
│   ├── WithHashDiversity()    ⭐ 新增
│   ├── WithMinDistance()      ⭐ 新增
│   └── WithMMR()              ⭐ 新增
│
├── builder_vector.go         # BuilderX 扩展
│   ├── VectorSearch()
│   ├── WithDiversity()        ⭐ 新增
│   ├── WithHashDiversity()    ⭐ 新增
│   ├── WithMinDistance()      ⭐ 新增
│   └── WithMMR()              ⭐ 新增
│
├── to_vector_sql.go          # PostgreSQL SQL 生成
│   └── SqlOfVectorSearch()   (忽略多样性)
│
├── to_qdrant_json.go         # Qdrant JSON 生成 ⭐ 新增
│   ├── ToQdrantJSON()
│   ├── ToQdrantRequest()
│   ├── QdrantSearchRequest
│   ├── QdrantFilter
│   └── QdrantCondition
│
├── qdrant_test.go            # Qdrant 测试 ⭐ 新增
└── VECTOR_DIVERSITY_QDRANT.md  # 文档 ⭐ 新增
```

---

## 📊 数据流

### PostgreSQL 流程

```
用户代码
  ↓
sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    WithHashDiversity("semantic_hash")
  ↓
Build()
  ↓
VectorSearchParams {
    QueryVector: vec,
    TopK: 20,
    Diversity: &DiversityParams{  // ⭐ 存在但被忽略
        Enabled: true,
        Strategy: "hash",
        HashField: "semantic_hash",
        OverFetchFactor: 5,
    }
}
  ↓
SqlOfVectorSearch()
  ↓
⭐ 关键：只使用 TopK，忽略 Diversity
  ↓
SQL: SELECT ... LIMIT 20
Args: [vec, "golang"]
```

---

### Qdrant 流程

```
用户代码
  ↓
sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    WithHashDiversity("semantic_hash")
  ↓
Build()
  ↓
VectorSearchParams {
    QueryVector: vec,
    TopK: 20,
    Diversity: &DiversityParams{
        Enabled: true,
        Strategy: "hash",
        HashField: "semantic_hash",
        OverFetchFactor: 5,
    }
}
  ↓
ToQdrantJSON()
  ↓
⭐ 关键：应用 Diversity，扩大 limit
  ↓
{
  "vector": [0.1, 0.2, 0.3],
  "limit": 100,  // 20 * 5 ⭐
  "filter": {...},
  "with_payload": true
}
  ↓
应用层去重（由用户实现）
  ↓
返回 Top-20 多样化结果
```

---

## 🔧 Qdrant JSON 映射

### sqlxb → Qdrant 映射表

| sqlxb 操作 | Qdrant JSON | 说明 |
|-----------|-------------|------|
| `Eq("k", v)` | `{"key": "k", "match": {"value": v}}` | 精确匹配 |
| `In("k", v1, v2)` | `{"key": "k", "match": {"any": [v1, v2]}}` | 多值匹配 |
| `Gt("k", v)` | `{"key": "k", "range": {"gt": v}}` | 大于 |
| `Gte("k", v)` | `{"key": "k", "range": {"gte": v}}` | 大于等于 |
| `Lt("k", v)` | `{"key": "k", "range": {"lt": v}}` | 小于 |
| `Lte("k", v)` | `{"key": "k", "range": {"lte": v}}` | 小于等于 |
| `Ne("k", v)` | ❌ 忽略 | Qdrant 需用 must_not |
| `Like("k", v)` | ❌ 忽略 | Qdrant 不支持 |
| `Between("k", v1, v2)` | ❌ 暂不支持 | 未来可能支持 |

---

### 完整示例

**sqlxb 代码**：

```go
sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    Gt("quality_score", 0.8).
    In("layer", "service", "repository").
    VectorSearch("embedding", queryVector, 20).
    WithHashDiversity("semantic_hash").
    Build().
    ToQdrantJSON()
```

**生成的 JSON**：

```json
{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "limit": 100,
  "filter": {
    "must": [
      {
        "key": "language",
        "match": {"value": "golang"}
      },
      {
        "key": "quality_score",
        "range": {"gt": 0.8}
      },
      {
        "key": "layer",
        "match": {"any": ["service", "repository"]}
      }
    ]
  },
  "with_payload": true,
  "params": {
    "hnsw_ef": 128
  }
}
```

---

## ✅ 测试验证

### 测试用例

```go
// 1. 基础 Qdrant JSON 生成
TestToQdrantJSON_Basic                ✅ PASS

// 2. 带过滤器
TestToQdrantJSON_WithFilter           ✅ PASS

// 3. 哈希多样性
TestToQdrantJSON_WithHashDiversity    ✅ PASS
// 验证：limit 从 20 扩大到 100

// 4. 最小距离多样性
TestToQdrantJSON_WithMinDistance      ✅ PASS
// 验证：limit 从 20 扩大到 100

// 5. MMR 多样性
TestToQdrantJSON_WithMMR              ✅ PASS
// 验证：limit 从 20 扩大到 100

// 6. 范围查询
TestToQdrantJSON_WithRange            ✅ PASS
// 验证：Gt, Lt 正确转换

// 7. IN 查询
TestToQdrantJSON_WithIn               ✅ PASS
// 验证：IN 转换为 match.any

// 8. PostgreSQL 忽略多样性 ⭐ 关键
TestSqlOfVectorSearch_IgnoresDiversity  ✅ PASS
// 验证：SQL LIMIT 保持为 20，不是 100

// 9. 完整工作流
TestQdrant_FullWorkflow               ✅ PASS
// 验证：一份代码，两种后端
```

**测试结果**：

```
=== RUN   TestToQdrantJSON_Basic
--- PASS: TestToQdrantJSON_Basic (0.00s)
=== RUN   TestToQdrantJSON_WithHashDiversity
    qdrant_test.go:127: ✅ 多样性启用：Limit 从 20 扩大到 100（5倍过度获取）
--- PASS: TestToQdrantJSON_WithHashDiversity (0.00s)
=== RUN   TestSqlOfVectorSearch_IgnoresDiversity
    qdrant_test.go:284: ✅ 多样性参数被正确忽略（PostgreSQL 不支持）
--- PASS: TestSqlOfVectorSearch_IgnoresDiversity (0.00s)
=== RUN   TestQdrant_FullWorkflow
    qdrant_test.go:318: ✅ 一份代码，两种后端：PostgreSQL 和 Qdrant
    qdrant_test.go:319: ✅ 优雅降级：不支持的功能自动忽略
--- PASS: TestQdrant_FullWorkflow (0.00s)
PASS
ok      github.com/x-ream/sqlxb 0.830s
```

---

## 💡 设计亮点

### 1. 零侵入式多样性

```go
// 不需要修改现有代码
existing := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    Build()

// 只需添加一行
enhanced := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    WithHashDiversity("semantic_hash").  // ⭐ 新增一行
    Build()

// 两者在 PostgreSQL 中行为完全一致
```

---

### 2. 向后兼容

```go
// 旧代码（v0.8.1）
VectorSearchParams {
    QueryVector: vec,
    TopK: 20,
    DistanceMetric: CosineDistance,
}

// 新代码（v0.9.0）
VectorSearchParams {
    QueryVector: vec,
    TopK: 20,
    DistanceMetric: CosineDistance,
    Diversity: nil,  // ⭐ 新增，默认 nil
}

// 完全兼容：nil Diversity 不影响任何逻辑
```

---

### 3. 错误容忍

```go
// bbToQdrantCondition 中不支持的操作
case LIKE:
    return nil, fmt.Errorf("LIKE not supported in Qdrant")

// buildQdrantFilter 中的处理
cond, err := bbToQdrantCondition(bb)
if err != nil {
    // ⭐ 关键：不支持的操作不报错，忽略即可
    continue
}
```

**结果**：

```go
// 即使使用了 LIKE，也不会报错
sqlxb.Of(&CodeVector{}).
    Like("content", "%login%").  // ⭐ 被忽略
    VectorSearch("embedding", vec, 10).
    Build().
    ToQdrantJSON()

// 生成的 JSON 中不包含 LIKE 条件，但不报错
```

---

### 4. 扩展性

```go
// 未来可以轻松添加新的多样性策略

const (
    DiversityByHash     DiversityStrategy = "hash"
    DiversityByDistance DiversityStrategy = "distance"
    DiversityByMMR      DiversityStrategy = "mmr"
    DiversityByCluster  DiversityStrategy = "cluster"  // ⭐ 未来
)

// 或新的后端
func (built *Built) ToMilvusJSON() (string, error) {
    // ...
}
```

---

## 🎯 总结

### 完成的目标

✅ **目标 1**：设计 API，支持 Qdrant  
   - 链式 API：`WithHashDiversity()`, `WithMinDistance()`, `WithMMR()`
   - JSON 生成：`ToQdrantJSON()`, `ToQdrantRequest()`

✅ **目标 2**：其他数据库自动忽略，不报错  
   - PostgreSQL：`SqlOfVectorSearch()` 忽略 `Diversity`
   - 不支持的操作：`LIKE`, `BETWEEN` 被忽略，不报错

✅ **目标 3**：转 Qdrant 需要的 JSON  
   - 完整的 JSON 生成：`vector`, `limit`, `filter`, `params`
   - 多样性自动应用：`limit` 自动扩大

---

### 核心价值

```
1. 优雅降级（Graceful Degradation）
   → 同一份代码，多种后端

2. 零学习成本
   → 链式 API，符合 sqlxb 风格

3. 类型安全
   → 编译时检查，减少运行时错误

4. AI-First
   → 清晰的模块边界，AI 易于理解和扩展
```

---

### 文件清单

```
新增文件：
✅ sqlxb/to_qdrant_json.go           (Qdrant JSON 生成)
✅ sqlxb/qdrant_test.go              (测试)
✅ sqlxb/VECTOR_DIVERSITY_QDRANT.md  (用户文档)
✅ sqlxb/VECTOR_DIVERSITY_API_DESIGN.md  (设计文档)

修改文件：
✅ sqlxb/vector_types.go             (添加 DiversityParams)
✅ sqlxb/cond_builder_vector.go      (添加 WithDiversity 等方法)
✅ sqlxb/builder_vector.go           (添加 BuilderX 扩展)

测试结果：
✅ 所有测试通过 (9/9)
✅ 向后兼容
✅ 优雅降级验证
```

---

**设计完成！** 🎉

用户可以立即开始使用：

```go
import "github.com/x-ream/sqlxb"

// PostgreSQL
sql, args := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    WithHashDiversity("semantic_hash").
    Build().
    SqlOfVectorSearch()

// Qdrant
json, _ := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    WithHashDiversity("semantic_hash").
    Build().
    ToQdrantJSON()
```

