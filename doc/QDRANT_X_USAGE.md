# QdrantX 使用指南

## 🎯 设计目的

`QdrantX` 提供了更高层次的 Qdrant 专属 API 封装，让所有 Qdrant 配置集中在一个地方。

**优势**：
- ✅ 语义清晰：明确这是 Qdrant 专属查询
- ✅ 配置集中：所有 Qdrant 参数在一起
- ✅ 链式调用：流畅的 API
- ✅ 向后兼容：可以不使用

**性能**：
- ⚠️ 会有轻微的性能开销（封装层）
- ✅ 但适当的封装是有必要的（可读性和可维护性）

---

## 🚀 快速开始

### 基础用法

```go
import "github.com/fndome/xb"

queryVector := xb.Vector{0.1, 0.2, 0.3, 0.4}

// 推荐用法：VectorSearch 在外部，QdrantX 只配置 Qdrant 专属参数
built := xb.Of(&CodeVector{}).
    Eq("language", "golang").                    // 通用条件
    VectorSearch("embedding", queryVector, 20).  // ⭐ 通用向量检索
    QdrantX(func(qx *xb.QdrantBuilderX) {
        // ⭐ 只有 Qdrant 专属配置
        qx.HnswEf(256).
            ScoreThreshold(0.8)
    }).
    Build()

// 生成 Qdrant JSON
json, err := built.ToQdrantJSON()
```

---

## 📚 API 详解

### QdrantX 专属方法

**注意**: `VectorSearch`, `VectorDistance`, `WithHashDiversity` 等是**通用方法**，应该在 `QdrantX` **外部**调用！

```go
// ⭐ Qdrant 专属方法（只在 QdrantX 内使用）
qx.HnswEf(ef int)                    // HNSW ef 参数
qx.ScoreThreshold(threshold float32) // 最小相似度阈值
qx.Exact(exact bool)                 // 精确/近似搜索
qx.WithVector(withVector bool)       // 是否返回向量
qx.X(key string, value interface{})  // 自定义参数
```

---

### 性能参数方法

```go
// HnswEf 设置 HNSW 算法的 ef 参数
// ef 越大 = 精度越高，速度越慢
// 推荐值: 64-256
qx.HnswEf(ef int)

// 快捷模式
qx.HighPrecision()  // ef=512（慢但准）
qx.Balanced()       // ef=128（默认，平衡）
qx.HighSpeed()      // ef=32（快但略不准）
```

---

### 过滤参数方法

```go
// ScoreThreshold 设置最小相似度阈值
// 只返回相似度 >= threshold 的结果
qx.ScoreThreshold(threshold float32)

// Exact 设置是否精确搜索（不使用索引）
// true: 精确（慢但完全准确）
// false: 近似（快但使用索引）
qx.Exact(exact bool)
```

---

### 结果控制方法

```go
// WithVector 设置是否返回向量数据
// true: 返回向量（占用带宽，可用于后续计算）
// false: 不返回（节省带宽）⭐ 推荐
qx.WithVector(withVector bool)

// X() 自定义 Qdrant 参数（扩展点）
// 用于未封装的 Qdrant 参数
qx.X(key string, value interface{})
```

**⚠️ 分页说明**: 不使用 `QdrantX` 设置分页，应该使用 `xb` 的 `Paged()` 方法！

```go
// ✅ 正确的分页方式
xb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    Paged(3, 20).  // ⭐ 第 3 页，每页 20 条
    QdrantX(func(qx *QdrantBuilderX) {
        qx.HnswEf(128)
    }).
    Build()
```

---

## 💡 实际应用场景

### 场景 1: 高精度代码搜索

```go
// 需求：查找最相似的代码，要求高精度
queryVector := embedding.Encode("用户登录逻辑")

built := xb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).  // ⭐ 通用向量检索
    WithHashDiversity("semantic_hash").           // ⭐ 通用多样性
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HighPrecision().         // ⭐ Qdrant 专属：高精度模式
            ScoreThreshold(0.9)     // ⭐ Qdrant 专属：高阈值
    }).
    Build()

json, _ := built.ToQdrantJSON()
```

**生成的 JSON**：

```json
{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "limit": 50,
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}}
    ]
  },
  "score_threshold": 0.9,
  "params": {
    "hnsw_ef": 512
  }
}
```

---

### 场景 2: 快速推荐（牺牲精度）

```go
// 需求：快速推荐相关文章，可以容忍轻微误差
articleVector := currentArticle.Embedding

built := xb.Of(&Article{}).
    Eq("category", "tech").
    VectorSearch("embedding", articleVector, 20).  // ⭐ 通用向量检索
    WithMMR(0.6).                                   // ⭐ 通用多样性
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HighSpeed()                              // ⭐ Qdrant 专属：高速模式
    }).
    Build()

json, _ := built.ToQdrantJSON()
```

**生成的 JSON**：

```json
{
  "vector": [...],
  "limit": 100,
  "filter": {...},
  "params": {
    "hnsw_ef": 32
  }
}
```

---

### 场景 3: 分页查询

```go
// 需求：分页展示向量搜索结果
page := 3      // 第 3 页
pageSize := 20 // 每页 20 条

built := xb.Of(&Document{}).
    VectorSearch("embedding", queryVector, pageSize).  // ⭐ 通用向量检索
    Paged(page, pageSize).                             // ⭐ 使用 xb 的 Paged()
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.Balanced()                                  // ⭐ Qdrant 专属：平衡模式
    }).
    Build()

json, _ := built.ToQdrantJSON()
```

**生成的 JSON**：

```json
{
  "vector": [...],
  "limit": 20,
  "offset": 40,
  "params": {
    "hnsw_ef": 128
  }
}
```

---

### 场景 4: 完整配置（生产环境）

```go
// 需求：生产级查询，需要精细控制
built := xb.Of(&CodeVector{}).
    Eq("language", "golang").
    Gt("quality_score", 0.7).
    VectorSearch("embedding", queryVector, 20).       // ⭐ 通用向量检索
    VectorDistance(xb.CosineDistance).             // ⭐ 通用距离度量
    WithHashDiversity("semantic_hash").               // ⭐ 通用多样性
    Paged(1, 20).                                     // ⭐ 通用分页
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HnswEf(256).                               // ⭐ Qdrant 专属：精度
            ScoreThreshold(0.75).                     // ⭐ Qdrant 专属：阈值
            WithVector(false)                         // ⭐ Qdrant 专属：不返回向量
    }).
    Build()

json, _ := built.ToQdrantJSON()
```

---

## 🎨 正确的用法

### ✅ 推荐用法：清晰分离

```go
// VectorSearch 和多样性在外部（通用方法）
// QdrantX 只配置 Qdrant 专属参数
built := xb.Of(&CodeVector{}).
    Eq("language", "golang").                     // 通用条件
    VectorSearch("embedding", vec, 20).           // ⭐ 通用向量检索
    WithHashDiversity("semantic_hash").           // ⭐ 通用多样性
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HnswEf(256).                           // ⭐ Qdrant 专属
            ScoreThreshold(0.8).                  // ⭐ Qdrant 专属
            WithVector(false)                     // ⭐ Qdrant 专属
    }).
    Build()
```

**优点**：
- ✅ 清晰的职责分离（通用 vs Qdrant 专属）
- ✅ 可以同时生成 PostgreSQL SQL 和 Qdrant JSON
- ✅ 与 xb 风格一致

---

## 📊 性能模式选择

### 模式对比

| 模式 | HnswEf | 精度 | 速度 | 适用场景 |
|------|--------|------|------|---------|
| `HighSpeed()` | 32 | ⭐⭐ | ⭐⭐⭐⭐⭐ | 推荐系统、实时查询 |
| `Balanced()` | 128 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 默认，适合大多数场景 ⭐ |
| `HighPrecision()` | 512 | ⭐⭐⭐⭐⭐ | ⭐⭐ | 精确搜索、关键业务 |

### 选择建议

```go
// 推荐系统、内容发现
qx.HighSpeed()  // 快速返回，轻微误差可接受

// 代码搜索、文档检索
qx.Balanced()   // ⭐ 默认，平衡精度和速度

// 法律文书、医疗诊断
qx.HighPrecision()  // 精度优先，性能其次
```

---

## 🎯 最佳实践

### 1. VectorSearch 必须在 QdrantX 外部调用

```go
// ✅ 正确：VectorSearch 在外部（通用方法）
VectorSearch("embedding", vec, 20).
QdrantX(func(qx *QdrantBuilderX) {
    qx.HnswEf(256)  // 只配置 Qdrant 专属参数
})

// ❌ 错误：QdrantBuilderX 没有 VectorSearch 方法！
QdrantX(func(qx *QdrantBuilderX) {
    qx.VectorSearch("embedding", vec, 20)  // ❌ 编译错误
})
```

---

### 2. 使用快捷模式而非手动设置

```go
// ✅ 推荐：使用快捷模式
qx.HighPrecision()

// ⚠️ 不推荐：手动设置（容易出错）
qx.HnswEf(512).Exact(false)
```

---

### 3. 生产环境设置 ScoreThreshold

```go
// ✅ 推荐：设置阈值，避免不相关结果
qx.ScoreThreshold(0.7)  // 只返回相似度 >= 0.7 的结果

// ❌ 不推荐：不设置阈值（可能返回不相关结果）
```

---

### 4. 节省带宽：不返回向量

```go
// ✅ 推荐：不返回向量数据（节省带宽）
qx.WithVector(false)

// ⚠️ 只在需要向量时才设置 true
// 例如：需要计算向量距离、二次检索等
qx.WithVector(true)
```

---

## 📖 完整示例

### 示例 1: 代码语义搜索

```go
package main

import (
    "github.com/fndome/xb"
    "github.com/qdrant/go-client/qdrant"
)

type CodeVector struct {
    Id           int64  `db:"id"`
    Content      string `db:"content"`
    Embedding    xb.Vector `db:"embedding"`
    Language     string `db:"language"`
    SemanticHash string `db:"semantic_hash"`
}

func (CodeVector) TableName() string {
    return "code_vectors"
}

func searchCode(query string, language string) ([]CodeVector, error) {
    // 1. 生成查询向量（调用 Python 服务）
    queryVector := callEmbeddingService(query)
    
    // 2. 构建 Qdrant 查询
    built := xb.Of(&CodeVector{}).
        Eq("language", language).
        VectorSearch("embedding", queryVector, 20).     // ⭐ 通用向量检索
        WithHashDiversity("semantic_hash").             // ⭐ 通用多样性
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.Balanced().                              // ⭐ Qdrant 专属：平衡模式
                ScoreThreshold(0.7).                    // ⭐ Qdrant 专属：最低相似度
                WithVector(false)                       // ⭐ Qdrant 专属：不返回向量
        }).
        Build()
    
    // 3. 执行查询
    jsonStr, _ := built.ToQdrantJSON()
    results := qdrantClient.Search("code_vectors", jsonStr)
    
    // 4. 应用层多样性过滤（基于 semantic_hash）
    uniqueResults := deduplicateByHash(results, "semantic_hash", 20)
    
    return uniqueResults, nil
}
```

---

### 示例 2: 分页查询

```go
func searchCodesPaged(query string, page, pageSize int) ([]CodeVector, error) {
    queryVector := callEmbeddingService(query)
    
    built := xb.Of(&CodeVector{}).
        VectorSearch("embedding", queryVector, pageSize).  // ⭐ 通用向量检索
        Paged(page, pageSize).                             // ⭐ 通用分页（xb 方法）
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.Balanced()                                  // ⭐ Qdrant 专属：平衡模式
        }).
        Build()
    
    jsonStr, _ := built.ToQdrantJSON()
    results := qdrantClient.Search("code_vectors", jsonStr)
    
    return results, nil
}

// 使用
results1 := searchCodesPaged("user login", 1, 20) // 第 1 页
results2 := searchCodesPaged("user login", 2, 20) // 第 2 页
results3 := searchCodesPaged("user login", 3, 20) // 第 3 页
```

---

### 示例 3: 高精度 + 高阈值（关键业务）

```go
// 法律文书检索：需要非常精确的匹配
func searchLegalCases(query string) ([]LegalCase, error) {
    queryVector := callEmbeddingService(query)
    
    built := xb.Of(&LegalCase{}).
        Eq("court_level", "最高法院").
        VectorSearch("case_embedding", queryVector, 10).  // ⭐ 通用向量检索
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.HighPrecision().         // ⭐ Qdrant 专属：高精度（ef=512）
                ScoreThreshold(0.95)    // ⭐ Qdrant 专属：高阈值
        }).
        Build()
    
    jsonStr, _ := built.ToQdrantJSON()
    results := qdrantClient.Search("legal_cases", jsonStr)
    
    return results, nil
}
```

---

### 示例 4: 高速推荐（容忍误差）

```go
// 内容推荐：速度优先，轻微误差可接受
func recommendArticles(userVector xb.Vector) ([]Article, error) {
    built := xb.Of(&Article{}).
        Eq("status", "published").
        VectorSearch("embedding", userVector, 50).     // ⭐ 通用向量检索
        WithMMR(0.6).                                   // ⭐ 通用多样性（避免重复推荐）
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.HighSpeed().                             // ⭐ Qdrant 专属：高速模式（ef=32）
                ScoreThreshold(0.5).                    // ⭐ Qdrant 专属：较低阈值（扩大范围）
                WithVector(false)                       // ⭐ Qdrant 专属：不返回向量
        }).
        Build()
    
    jsonStr, _ := built.ToQdrantJSON()
    results := qdrantClient.Search("articles", jsonStr)
    
    // 应用层 MMR 过滤
    diverse := applyMMR(results, userVector, 0.6, 20)
    
    return diverse, nil
}
```

---

## ⚖️ 性能权衡

### HnswEf 参数的影响

| ef 值 | 精度 | 速度 | 内存 | 推荐场景 |
|-------|------|------|------|---------|
| 32 | 90% | 非常快 | 低 | 推荐系统、实时查询 |
| 64 | 95% | 快 | 中 | 一般搜索 |
| 128 | 98% | 中等 | 中 | 默认值 ⭐ |
| 256 | 99% | 较慢 | 较高 | 精确搜索 |
| 512 | 99.5% | 慢 | 高 | 关键业务 |

**建议**：
- 开发/测试：64-128
- 生产环境：128-256
- 关键业务：256-512

---

## 🔧 高级技巧

### 技巧 1: 动态调整精度

```go
func search(query string, precision string) {
    queryVector := callEmbeddingService(query)
    
    built := xb.Of(&CodeVector{}).
        VectorSearch("embedding", queryVector, 20).  // ⭐ 通用向量检索
        QdrantX(func(qx *xb.QdrantBuilderX) {
            // 根据用户选择调整精度
            switch precision {
            case "high":
                qx.HighPrecision()
            case "low":
                qx.HighSpeed()
            default:
                qx.Balanced()
            }
        }).
        Build()
}
```

---

### 技巧 2: 条件性多样性

```go
func search(query string, needDiversity bool) {
    queryVector := callEmbeddingService(query)
    
    builder := xb.Of(&CodeVector{}).
        VectorSearch("embedding", queryVector, 20)  // ⭐ 通用向量检索
    
    // 条件性应用多样性（在外部）
    if needDiversity {
        builder.WithHashDiversity("semantic_hash")
    }
    
    built := builder.
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.Balanced()  // ⭐ Qdrant 专属配置
        }).
        Build()
}
```

---

## 📝 总结

### QdrantX 的价值

```
优势:
  ✅ 配置集中（所有 Qdrant 参数在一起）
  ✅ 语义清晰（明确这是 Qdrant 专属）
  ✅ 链式调用（流畅的 API）
  ✅ 快捷模式（HighPrecision, HighSpeed, Balanced）

性能:
  ⚠️ 有轻微封装开销
  ✅ 但适当的封装是必要的
  ✅ 提高可读性和可维护性

向后兼容:
  ✅ 可以不使用 QdrantX
  ✅ 原有 API 仍然有效
```

---

### 推荐用法

```go
// ⭐ 推荐：清晰分离通用方法和 Qdrant 专属配置
xb.Of(&Model{}).
    Eq("common_field", value).       // 通用条件
    VectorSearch("embedding", vec, 20).  // ⭐ 通用向量检索
    WithHashDiversity("hash_field").     // ⭐ 通用多样性
    QdrantX(func(qx *QdrantBuilderX) {
        qx.HnswEf(256).                  // ⭐ Qdrant 专属
            ScoreThreshold(0.8)          // ⭐ Qdrant 专属
    })
```

---

**开始使用 QdrantX，让 Qdrant 查询更清晰！** 🚀

