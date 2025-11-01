# xb 框架命名规范

## 🎯 核心原则

**统一性**: 所有数据库后端（SQL、Qdrant、Milvus、Weaviate 等）使用一致的命名模式。

---

## 📋 命名规范

### 1. 查询生成方法

所有查询生成方法都定义在 `Built` 结构体上，遵循以下模式：

```
<Backend>Of<Operation>()
```

| 后端类型 | 方法命名模式 | 示例 |
|---------|-------------|------|
| SQL | `SqlOf<Operation>()` | `SqlOfSelect()`, `SqlOfInsert()` |
| Qdrant | `JsonOfQdrant<Operation>()` | `JsonOfQdrantSelect()` |
| Milvus | `JsonOfMilvus<Operation>()` | `JsonOfMilvusSelect()`, `JsonOfMilvusInsert()` |
| Weaviate | `JsonOfWeaviate<Operation>()` | `JsonOfWeaviateSelect()` |

---

## 📊 完整方法列表

### SQL 后端

```go
// 定义在 to_sql.go
func (built *Built) SqlOfSelect() (string, []interface{}, map[string]string)
func (built *Built) SqlOfInsert() (string, []interface{})
func (built *Built) SqlOfUpdate() (string, []interface{})
func (built *Built) SqlOfDelete() (string, []interface{})
func (built *Built) SqlOfPage() (string, string, []interface{}, map[string]string)
func (built *Built) SqlOfCond() (string, string, []interface{})

// 向量查询（PostgreSQL + pgvector）
func (built *Built) SqlOfVectorSearch() (string, []interface{})
```

### Qdrant 后端

```go
// 定义在 to_qdrant_json.go
func (built *Built) JsonOfQdrantSelect() (string, error)      // 向量搜索
func (built *Built) JsonOfQdrantRecommend() (string, error)   // 推荐查询
func (built *Built) JsonOfQdrantScroll() (string, error)      // 游标遍历
func (built *Built) JsonOfQdrantDiscover() (string, error)    // 探索查询
// 注: Qdrant 是只读向量数据库，不支持 Insert/Update/Delete
```

### Milvus 后端

```go
// 定义在 to_milvus_json.go（未来实现）
func (built *Built) JsonOfMilvusSelect() (string, error)      // 向量搜索
func (built *Built) JsonOfMilvusInsert() (string, error)      // 向量插入
func (built *Built) JsonOfMilvusUpdate() (string, error)      // 向量更新
func (built *Built) JsonOfMilvusDelete() (string, error)      // 向量删除
```

### Weaviate 后端

```go
// 定义在 to_weaviate_json.go（未来实现）
func (built *Built) JsonOfWeaviateSelect() (string, error)    // 向量搜索
func (built *Built) JsonOfWeaviateInsert() (string, error)    // 向量插入
// 注: Weaviate 不支持独立的 Update/Delete
```

---

## 🔄 操作类型对照表

| 操作 | SQL | Qdrant | Milvus | Weaviate | 说明 |
|------|-----|--------|--------|----------|------|
| **查询** | `SqlOfSelect()` | `JsonOfQdrantSelect()` | `JsonOfMilvusSelect()` | `JsonOfWeaviateSelect()` | 所有后端必须支持 |
| **插入** | `SqlOfInsert()` | - | `JsonOfMilvusInsert()` | `JsonOfWeaviateInsert()` | Qdrant 不支持 |
| **更新** | `SqlOfUpdate()` | - | `JsonOfMilvusUpdate()` | - | 多数向量库不支持 |
| **删除** | `SqlOfDelete()` | - | `JsonOfMilvusDelete()` | - | 多数向量库不支持 |
| **推荐** | - | `JsonOfQdrantRecommend()` | - | - | Qdrant 专属 |
| **滚动** | - | `JsonOfQdrantScroll()` | - | - | Qdrant 专属 |
| **探索** | - | `JsonOfQdrantDiscover()` | - | - | Qdrant 专属 |
| **分页** | `SqlOfPage()` | - | - | - | SQL 专属 |

---

## 💡 使用示例

### 基础用法

```go
import "github.com/fndo-io/xb"

// 构建查询条件
built := xb.C().
    VectorScoreThreshold(0.8).
    VectorSearch("code_vectors", "embedding", vec, 20, xb.CosineDistance).
    Build()

// ⭐ 生成不同后端的查询（命名统一）
sql, args, _ := built.SqlOfSelect()           // PostgreSQL + pgvector
qdrantJSON, _ := built.JsonOfQdrantSelect()   // Qdrant
milvusJSON, _ := built.JsonOfMilvusSelect()   // Milvus（未来）
```

### 跨后端部署

```go
// 同一份业务逻辑，支持多个后端
func SearchCodeVectors(query string, embedding []float32) ([]Result, error) {
    built := xb.C().
        Eq("language", "golang").
        VectorScoreThreshold(0.8).
        VectorSearch("code_vectors", "embedding", embedding, 20, xb.CosineDistance).
        Build()
    
    // 根据配置选择后端
    switch config.VectorDB {
    case "postgres":
        sql, args, _ := built.SqlOfSelect()
        return db.Query(sql, args...)
    
    case "qdrant":
        json, _ := built.JsonOfQdrantSelect()
        return qdrantClient.Search(json)
    
    case "milvus":
        json, _ := built.JsonOfMilvusSelect()
        return milvusClient.Search(json)
    
    default:
        return nil, errors.New("unsupported vector db")
    }
}
```

---

## 🎨 命名规则详解

### 1. 方法前缀

| 前缀 | 返回类型 | 适用后端 | 示例 |
|------|---------|---------|------|
| `SqlOf` | SQL 字符串 | PostgreSQL, MySQL, SQLite 等 | `SqlOfSelect()` |
| `JsonOf` | JSON 字符串 | Qdrant, Milvus, Weaviate 等 | `JsonOfQdrantSelect()` |

### 2. 后端标识

| 标识 | 含义 | 示例 |
|------|------|------|
| (无) | SQL 数据库（通用） | `SqlOfSelect()` |
| `Qdrant` | Qdrant 向量数据库 | `JsonOfQdrantSelect()` |
| `Milvus` | Milvus 向量数据库 | `JsonOfMilvusSelect()` |
| `Weaviate` | Weaviate 向量数据库 | `JsonOfWeaviateSelect()` |

### 3. 操作名称

| 操作 | 说明 | 所有后端通用 |
|------|------|------------|
| `Select` | 查询/搜索 | ✅ 是 |
| `Insert` | 插入 | ✅ 是 |
| `Update` | 更新 | ✅ 是 |
| `Delete` | 删除 | ✅ 是 |
| `Recommend` | 推荐 | ❌ Qdrant 专属 |
| `Scroll` | 游标遍历 | ❌ Qdrant 专属 |
| `Discover` | 探索 | ❌ Qdrant 专属 |
| `Page` | 分页 | ❌ SQL 专属 |

---

## ✅ 命名一致性检查

添加新方法时，请确保：

- [ ] 方法定义在 `Built` 结构体上
- [ ] 使用 `<Backend>Of<Operation>()` 模式
- [ ] 返回值类型符合规范（`string` 或 `(string, error)`）
- [ ] 方法无参数（从 `Built.Conds` 获取）
- [ ] 与现有方法命名风格一致

---

## 🚫 错误示例

```go
// ❌ 错误：不符合命名规范
func (built *Built) ToQdrantJSON() (string, error)
func (built *Built) GetMilvusSearchJSON() (string, error)
func (b *BuilderX) GenerateWeaviateQuery() (string, error)

// ✅ 正确：符合命名规范
func (built *Built) JsonOfQdrantSelect() (string, error)
func (built *Built) JsonOfMilvusSelect() (string, error)
func (built *Built) JsonOfWeaviateSelect() (string, error)
```

---

## 📚 相关文档

- `TO_JSON_DESIGN_CLARIFICATION.md` - JsonOf 方法设计说明
- `VECTOR_DB_EXTENSION_GUIDE.md` - 向量数据库扩展指南
- `MILVUS_TEMPLATE.go` - Milvus 实现模板

---

**版本**: v0.11.0  
**更新时间**: 2025-11-01  
**维护者**: xb Team

