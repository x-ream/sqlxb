# JsonOf 方法设计说明

## 🎯 核心设计原则

**统一命名**: 与 SQL 保持一致的命名规范

```
SqlOfSelect()        → 生成 SQL SELECT 查询
SqlOfInsert()        → 生成 SQL INSERT 语句
SqlOfUpdate()        → 生成 SQL UPDATE 语句
SqlOfDelete()        → 生成 SQL DELETE 语句

JsonOfQdrantSelect() → 生成 Qdrant 搜索 JSON
JsonOfMilvusSelect() → 生成 Milvus 搜索 JSON
JsonOfMilvusInsert() → 生成 Milvus 插入 JSON（如果支持）
...
```

**统一设计**: 所有方法都在 `Built` 上调用，参数从 `VectorSearch()` 中获取。

---

## ✅ 正确的设计（Qdrant 已实现）

### API 调用流程

```go
// Step 1: 构建查询（链式调用）
built := xb.C().
    VectorScoreThreshold(0.8).           // 通用参数
    QdrantHnswEf(128).                   // Qdrant 专属参数
    Eq("language", "golang").            // 过滤条件
    VectorSearch(
        "code_vectors",                  // 集合名称
        "embedding",                     // 向量字段
        []float32{0.1, 0.2, 0.3},       // 查询向量
        20,                              // Top K
        xb.CosineDistance,              // 距离度量
    ).
    Build()

// Step 2: 转换为 JSON（与 SQL 命名一致）
json, err := built.JsonOfQdrantSelect()
```

### 实现细节

```go
// ⭐ 与 SQL 命名一致：JsonOfQdrantSelect (类似 SqlOfSelect)
func (built *Built) JsonOfQdrantSelect() (string, error) {
    // 1. 从 Built.Conds 中提取 VectorSearch 参数
    req, err := built.ToQdrantRequest()
    if err != nil {
        return "", err
    }
    
    // 2. 应用 Qdrant 专属参数
    applyQdrantParams(built.Conds, req)
    
    // 3. 序列化为 JSON
    return mergeAndSerialize(req, built.Conds)
}
```

**优势**:
- ✅ **命名统一**: `JsonOfQdrantSelect()` 与 `SqlOfSelect()` 命名一致
- ✅ **参数集中**: 所有参数（集合名、向量、TopK、距离度量）都在 `VectorSearch()` 中
- ✅ **无需重复**: 调用时不需要再传参数
- ✅ **类型安全**: `Built` 包含了所有需要的信息
- ✅ **易于测试**: 可以多次调用生成相同 JSON

---

## ❌ 错误的设计（Milvus 模板之前的版本）

### 错误示例 1: 在 BuilderX 上调用

```go
// ❌ 错误：在 BuilderX 上调用，需要传参数
func (b *BuilderX) ToMilvusSearchJSON(
    collectionName string,    // ❌ 重复：与 VectorSearch 重复
    vectors [][]float32,      // ❌ 重复：与 VectorSearch 重复
    topK int,                 // ❌ 重复：与 VectorSearch 重复
    metricType string,        // ❌ 重复：与 VectorSearch 重复
) (string, error) {
    built := b.Build()
    // ...
}

// 调用时很混乱
json, err := xb.C().
    VectorSearch("users", "embedding", vec, 10, L2Distance).  // 已经指定了参数
    ToMilvusSearchJSON("users", [][]float32{vec}, 10, "L2")   // ❌ 又要重复一遍
```

**问题**:
- ❌ 参数重复：集合名、向量、TopK、距离度量要指定两次
- ❌ 容易出错：两次指定的参数可能不一致
- ❌ API 不一致：与 Qdrant 的设计不一致

---

### 错误示例 2: 参数分散

```go
// ❌ 错误：参数分散在多个地方
builder := xb.C().
    VectorScoreThreshold(0.8)   // 在这里

json, err := builder.ToMilvusSearchJSON(
    "users",                    // 在这里
    [][]float32{{0.1, 0.2}},   // 在这里
    10,                         // 在这里
    "L2",                       // 在这里
)
```

**问题**:
- ❌ 不知道向量搜索参数在哪里
- ❌ 无法复用 `Built` 对象

---

## ✅ 正确的 Milvus 设计（修正后）

### API 调用流程

```go
// Step 1: 构建查询（与 Qdrant 完全一致）
built := xb.C().
    VectorScoreThreshold(0.8).           // 通用参数
    MilvusNProbe(64).                    // Milvus 专属参数
    MilvusExpr("age > 18").              // Milvus 过滤表达式
    VectorSearch(
        "users",                         // 集合名称
        "embedding",                     // 向量字段
        []float32{0.1, 0.2, 0.3},       // 查询向量
        10,                              // Top K
        xb.L2Distance,                  // 距离度量
    ).
    Build()

// Step 2: 转换为 JSON（与 SQL 命名一致）
json, err := built.JsonOfMilvusSelect()
```

### 实现细节

```go
// ⭐ 与 SQL 命名一致：JsonOfMilvusSelect (类似 SqlOfSelect)
func (built *Built) JsonOfMilvusSelect() (string, error) {
    // 1. 从 Built.Conds 中提取 VectorSearch 参数
    vectorBb := findVectorSearchBb(built.Conds)
    if vectorBb == nil {
        return "", fmt.Errorf("no VECTOR_SEARCH found")
    }
    
    params := vectorBb.Value.(VectorSearchParams)
    
    // 2. 创建 Milvus 请求对象
    req := &MilvusSearchRequest{
        CollectionName: params.TableName,
        Vectors:        [][]float32{params.Vector},
        TopK:           params.Limit,
        MetricType:     milvusDistanceMetric(params.Distance),
    }
    
    // 3. 应用 Milvus 专属参数
    applyMilvusParams(built.Conds, req)
    
    // 4. 序列化为 JSON
    return milvusMergeAndSerialize(req, built.Conds)
}
```

**优势**:
- ✅ **命名统一**: `JsonOfMilvusSelect()` 与 `SqlOfSelect()` 命名一致
- ✅ **与 Qdrant 一致**: API 设计完全相同
- ✅ **参数不重复**: 所有参数都在 `VectorSearch()` 中
- ✅ **易于理解**: 用户只需学习一次
- ✅ **可复用**: `Built` 可以多次调用不同的 `JsonOfXxx()`

---

## 🔄 对比：不同向量数据库

### 统一的 API 设计

```go
// ⭐ 所有数据库都使用相同的调用方式
built := xb.C().
    // 通用参数
    VectorScoreThreshold(0.8).
    VectorWithVector(true).
    
    // 数据库专属参数
    QdrantHnswEf(128).           // Qdrant 专属
    // OR
    MilvusNProbe(64).            // Milvus 专属
    // OR
    WeaviateAlpha(0.5).          // Weaviate 专属
    
    // 向量搜索（通用）
    VectorSearch("collection", "field", vector, 10, distance).
    Build()

// ⭐ 转换为不同数据库的 JSON（命名统一）
qdrantJSON, _ := built.JsonOfQdrantSelect()
milvusJSON, _ := built.JsonOfMilvusSelect()
weaviateJSON, _ := built.JsonOfWeaviateSelect()

// ⭐ 对比 SQL（命名完全一致）
sql, args, _ := built.SqlOfSelect()
```

---

## 📊 设计对比总结

| 设计 | 方法命名 | 调用位置 | 参数传递 | 优势 | 劣势 |
|------|---------|---------|---------|------|------|
| **SQL（标准）** | `SqlOfSelect()` | `Built` 上 | 从 `Eq()/Gt()` 等获取 | ✅ 命名清晰 | - |
| **Qdrant（统一）** | `JsonOfQdrantSelect()` | `Built` 上 | 从 `VectorSearch()` 获取 | ✅ 与 SQL 一致<br>✅ 参数不重复 | - |
| **Milvus（统一）** | `JsonOfMilvusSelect()` | `Built` 上 | 从 `VectorSearch()` 获取 | ✅ 与 SQL 一致<br>✅ 易于理解 | - |
| **Milvus（修正前）** | `ToMilvusSearchJSON()` | `BuilderX` 上 | 手动传参 | - | ❌ 命名不一致<br>❌ 参数重复 |

---

## 🎯 实现检查清单

添加新的向量数据库支持时，请确保：

- [ ] **命名规范**: 使用 `JsonOfXxxSelect()`（与 `SqlOfSelect()` 一致）
- [ ] **定义位置**: 方法定义在 `Built` 上（不是 `BuilderX`）
- [ ] **参数传递**: 方法无需参数（从 `VectorSearch` 获取）
- [ ] **参数提取**: 使用 `findVectorSearchBb()` 提取向量搜索参数
- [ ] **通用参数**: 调用 `ApplyCommonVectorParams()` 应用通用参数
- [ ] **自定义参数**: 使用 `ExtractCustomParams()` 提取自定义参数
- [ ] **一致性**: API 与 SQL/Qdrant 保持一致

**支持的操作**（按需实现）:
- `JsonOfXxxSelect()` - 向量搜索/查询（必须）
- `JsonOfXxxInsert()` - 向量插入（如果数据库支持）
- `JsonOfXxxUpdate()` - 向量更新（如果数据库支持）
- `JsonOfXxxDelete()` - 向量删除（如果数据库支持）

---

## 📖 示例：完整的用户代码

### SQL 查询

```go
built := xb.C().
    Eq("language", "golang").
    Gt("score", 0.8).
    Build()

sql, args, _ := built.SqlOfSelect()  // ✅ 标准命名
```

### Qdrant（与 SQL 一致）

```go
built := xb.C().
    VectorScoreThreshold(0.8).
    QdrantHnswEf(128).
    VectorSearch("code_vectors", "embedding", vec, 20, xb.CosineDistance).
    Build()

json, _ := built.JsonOfQdrantSelect()  // ✅ 与 SqlOfSelect 一致
```

### Milvus（与 SQL 一致）

```go
built := xb.C().
    VectorScoreThreshold(0.8).
    MilvusNProbe(64).
    VectorSearch("code_vectors", "embedding", vec, 20, xb.L2Distance).
    Build()

json, _ := built.JsonOfMilvusSelect()  // ✅ 与 SqlOfSelect 一致
```

### 跨数据库查询（同一个 Built）

```go
built := xb.C().
    VectorScoreThreshold(0.8).
    VectorSearch("code_vectors", "embedding", vec, 20, xb.CosineDistance).
    Build()

// ⭐ 可以同时生成多个后端的查询（命名统一）
sql, args, _ := built.SqlOfSelect()           // PostgreSQL + pgvector
qdrantJSON, _ := built.JsonOfQdrantSelect()   // Qdrant
milvusJSON, _ := built.JsonOfMilvusSelect()   // Milvus

// 根据部署环境选择
switch env {
case "postgres":
    results := db.Query(sql, args...)
case "qdrant":
    results := qdrantClient.Search(qdrantJSON)
case "milvus":
    results := milvusClient.Search(milvusJSON)
}
```

---

## 🚀 总结

**核心原则**: 

1. **命名统一**: `JsonOfXxxSelect()` 与 `SqlOfSelect()` 保持一致
2. **调用统一**: 所有方法都在 `Built` 上调用
3. **参数统一**: 从 `VectorSearch()` 获取参数，无需重复传递

**优势**:
1. ✅ **命名一致性**: `SqlOfSelect()` / `JsonOfQdrantSelect()` / `JsonOfMilvusSelect()` 模式统一
2. ✅ **API 一致性**: SQL 和所有向量数据库使用相同的设计
3. ✅ **参数不重复**: 避免在多个地方指定相同参数
4. ✅ **易于理解**: 用户只需学习一次，到处适用
5. ✅ **可复用性**: `Built` 可以生成多个数据库的查询（SQL/JSON）
6. ✅ **类型安全**: 编译时检查，运行时零错误
7. ✅ **按需实现**: 只实现数据库支持的操作（Select/Insert/Update/Delete）

**避免**:
- ❌ 使用 `ToXxxJSON()` 命名（不符合 `SqlOfXxx()` 规范）
- ❌ 在 `BuilderX` 上定义方法（应在 `Built` 上）
- ❌ 需要手动传递集合名、向量、TopK 等参数
- ❌ 与 SQL 设计不一致

**命名规范表**:

| 操作 | SQL | Qdrant | Milvus | Weaviate |
|------|-----|--------|--------|----------|
| 查询 | `SqlOfSelect()` | `JsonOfQdrantSelect()` | `JsonOfMilvusSelect()` | `JsonOfWeaviateSelect()` |
| 插入 | `SqlOfInsert()` | - | `JsonOfMilvusInsert()` | `JsonOfWeaviateInsert()` |
| 更新 | `SqlOfUpdate()` | - | `JsonOfMilvusUpdate()` | - |
| 删除 | `SqlOfDelete()` | - | `JsonOfMilvusDelete()` | - |

*注: `-` 表示数据库不支持该操作，无需实现*

