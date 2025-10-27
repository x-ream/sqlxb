# 自定义向量数据库支持指南

## 🎯 概述

本指南演示如何为 `sqlxb` 添加自定义向量数据库支持（如 Milvus, Weaviate, Pinecone 等）。

**核心思路**：参照 `QdrantBuilderX` 的实现模式，创建自己的 `XxxxBuilderX`。

---

## 🏗️ 实现步骤

### 步骤 1: 创建自定义 BuilderX

```go
// my_project/vectordb/milvus_x.go
package vectordb

import (
    "github.com/x-ream/sqlxb"
)

// MilvusBuilderX Milvus 专属构建器
type MilvusBuilderX struct {
    builder *sqlxb.BuilderX
}

// MilvusX 创建 Milvus 专属构建器
// 用法:
//   sqlxb.Of(&CodeVector{}).
//       Eq("language", "golang").
//       VectorSearch("embedding", vec, 20).
//       MilvusX(func(mx *MilvusBuilderX) {
//           mx.Nprobe(10).
//               RoundDecimal(2)
//       })
func (x *sqlxb.BuilderX) MilvusX(f func(mx *MilvusBuilderX)) *sqlxb.BuilderX {
    mx := &MilvusBuilderX{
        builder: x,
    }
    
    f(mx)
    
    return x
}
```

---

### 步骤 2: 添加专属操作符

```go
// my_project/vectordb/milvus_oper.go
package vectordb

const (
    MILVUS_NPROBE        = "MILVUS_NPROBE"
    MILVUS_ROUND_DECIMAL = "MILVUS_ROUND_DECIMAL"
    MILVUS_METRIC_TYPE   = "MILVUS_METRIC_TYPE"
    MILVUS_XX            = "MILVUS_XX"  // 自定义扩展点
)
```

---

### 步骤 3: 实现专属方法

```go
// MilvusBuilderX 的方法实现
package vectordb

import "github.com/x-ream/sqlxb"

// Nprobe 设置 Milvus 的 nprobe 参数
// nprobe 越大，精度越高，但速度越慢
func (mx *MilvusBuilderX) Nprobe(nprobe int) *MilvusBuilderX {
    if nprobe > 0 {
        bb := sqlxb.Bb{
            Op:    MILVUS_NPROBE,
            Key:   "nprobe",
            Value: nprobe,
        }
        mx.builder.Bbs = append(mx.builder.Bbs, bb)
    }
    return mx
}

// RoundDecimal 设置 Milvus 的距离小数位数
func (mx *MilvusBuilderX) RoundDecimal(decimal int) *MilvusBuilderX {
    bb := sqlxb.Bb{
        Op:    MILVUS_ROUND_DECIMAL,
        Key:   "round_decimal",
        Value: decimal,
    }
    mx.builder.Bbs = append(mx.builder.Bbs, bb)
    return mx
}

// MetricType 设置 Milvus 的距离度量类型
func (mx *MilvusBuilderX) MetricType(metricType string) *MilvusBuilderX {
    bb := sqlxb.Bb{
        Op:    MILVUS_METRIC_TYPE,
        Key:   "metric_type",
        Value: metricType,
    }
    mx.builder.Bbs = append(mx.builder.Bbs, bb)
    return mx
}

// X 自定义 Milvus 参数（扩展点）
// 用于未封装的 Milvus 参数
func (mx *MilvusBuilderX) X(k string, v interface{}) *MilvusBuilderX {
    bb := sqlxb.Bb{
        Op:    MILVUS_XX,
        Key:   k,
        Value: v,
    }
    mx.builder.Bbs = append(mx.builder.Bbs, bb)
    return mx
}

// 快捷方法
func (mx *MilvusBuilderX) HighAccuracy() *MilvusBuilderX {
    return mx.Nprobe(256)
}

func (mx *MilvusBuilderX) Balanced() *MilvusBuilderX {
    return mx.Nprobe(64)
}

func (mx *MilvusBuilderX) HighSpeed() *MilvusBuilderX {
    return mx.Nprobe(16)
}
```

---

### 步骤 4: 实现 JSON 转换器

```go
// my_project/vectordb/to_milvus_json.go
package vectordb

import (
    "encoding/json"
    "github.com/x-ream/sqlxb"
)

// MilvusSearchRequest Milvus 搜索请求结构
type MilvusSearchRequest struct {
    CollectionName string                 `json:"collection_name"`
    Data           [][]float32            `json:"data"`
    Limit          int                    `json:"limit"`
    OutputFields   []string               `json:"output_fields,omitempty"`
    SearchParams   MilvusSearchParams     `json:"search_params"`
    Expr           string                 `json:"expr,omitempty"`
}

type MilvusSearchParams struct {
    MetricType   string `json:"metric_type"`
    Params       map[string]interface{} `json:"params"`
    RoundDecimal int    `json:"round_decimal,omitempty"`
}

// ToMilvusJSON 转换为 Milvus JSON
func (built *sqlxb.Built) ToMilvusJSON(collectionName string) (string, error) {
    req, err := built.ToMilvusRequest(collectionName)
    if err != nil {
        return "", err
    }
    
    bytes, err := json.MarshalIndent(req, "", "  ")
    if err != nil {
        return "", err
    }
    
    return string(bytes), nil
}

// ToMilvusRequest 转换为 Milvus 请求结构
func (built *sqlxb.Built) ToMilvusRequest(collectionName string) (*MilvusSearchRequest, error) {
    req := &MilvusSearchRequest{
        CollectionName: collectionName,
        SearchParams: MilvusSearchParams{
            MetricType: "L2",  // 默认值
            Params:     make(map[string]interface{}),
        },
    }
    
    // 1. 提取向量搜索参数
    for _, bb := range built.Conds {
        if bb.Op == sqlxb.VECTOR_SEARCH {
            params := bb.Value.(sqlxb.VectorSearchParams)
            req.Data = [][]float32{params.QueryVector}
            req.Limit = params.TopK
            
            // 距离度量映射
            switch params.DistanceMetric {
            case sqlxb.CosineDistance:
                req.SearchParams.MetricType = "IP"  // Inner Product
            case sqlxb.L2Distance:
                req.SearchParams.MetricType = "L2"
            }
            break
        }
    }
    
    // 2. 提取 Milvus 专属参数
    for _, bb := range built.Conds {
        switch bb.Op {
        case MILVUS_NPROBE:
            req.SearchParams.Params["nprobe"] = bb.Value
        case MILVUS_ROUND_DECIMAL:
            req.SearchParams.RoundDecimal = bb.Value.(int)
        case MILVUS_METRIC_TYPE:
            req.SearchParams.MetricType = bb.Value.(string)
        case MILVUS_XX:
            // 自定义参数
            req.SearchParams.Params[bb.Key] = bb.Value
        }
    }
    
    // 3. 构建标量过滤表达式（Milvus 的 expr）
    expr := buildMilvusExpr(built.Conds)
    if expr != "" {
        req.Expr = expr
    }
    
    return req, nil
}

// buildMilvusExpr 构建 Milvus 的过滤表达式
func buildMilvusExpr(bbs []sqlxb.Bb) string {
    var conditions []string
    
    for _, bb := range bbs {
        switch bb.Op {
        case sqlxb.EQ:
            conditions = append(conditions, fmt.Sprintf(`%s == "%v"`, bb.Key, bb.Value))
        case sqlxb.GT:
            conditions = append(conditions, fmt.Sprintf(`%s > %v`, bb.Key, bb.Value))
        case sqlxb.LT:
            conditions = append(conditions, fmt.Sprintf(`%s < %v`, bb.Key, bb.Value))
        case sqlxb.IN:
            // 处理 IN 条件
            values := []string{}
            // ... 转换为 Milvus 的 IN 表达式
        }
    }
    
    if len(conditions) == 0 {
        return ""
    }
    
    return strings.Join(conditions, " and ")
}
```

---

## 📚 完整示例

### 示例 1: 代码搜索（Milvus）

```go
package main

import (
    "fmt"
    "github.com/x-ream/sqlxb"
    "your-project/vectordb"
)

func main() {
    queryVector := sqlxb.Vector{0.1, 0.2, 0.3, 0.4}
    
    // 构建查询
    built := sqlxb.Of(&CodeVector{}).
        Eq("language", "golang").                      // 通用条件
        Gt("quality_score", 0.7).                      // 通用条件
        VectorSearch("embedding", queryVector, 20).    // ⭐ 通用向量检索
        WithHashDiversity("semantic_hash").            // ⭐ 通用多样性
        MilvusX(func(mx *vectordb.MilvusBuilderX) {
            mx.HighAccuracy().                         // ⭐ Milvus 专属
                RoundDecimal(4).                       // ⭐ Milvus 专属
                MetricType("IP")                       // ⭐ Milvus 专属
        }).
        Build()
    
    // 生成 Milvus JSON
    jsonStr, err := built.ToMilvusJSON("code_vectors")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(jsonStr)
}
```

**输出**：

```json
{
  "collection_name": "code_vectors",
  "data": [[0.1, 0.2, 0.3, 0.4]],
  "limit": 100,
  "search_params": {
    "metric_type": "IP",
    "params": {
      "nprobe": 256
    },
    "round_decimal": 4
  },
  "expr": "language == \"golang\" and quality_score > 0.7"
}
```

---

## 🎯 设计原则

### 1. 清晰分离：通用 vs 专属

```go
// ✅ 正确设计
sqlxb.Of(&Model{}).
    VectorSearch("embedding", vec, 20).      // ⭐ 通用方法（外部）
    WithHashDiversity("hash").                // ⭐ 通用方法（外部）
    MilvusX(func(mx *MilvusBuilderX) {
        mx.Nprobe(128).                       // ⭐ Milvus 专属（内部）
            RoundDecimal(4)                   // ⭐ Milvus 专属（内部）
    })

// ❌ 错误设计：不要在 BuilderX 内实现 VectorSearch
MilvusX(func(mx *MilvusBuilderX) {
    mx.VectorSearch(...)  // ❌ 不要这样做！
})
```

---

### 2. 保持向后兼容

```go
// ⭐ 通过扩展 BuilderX 而非修改
func (x *sqlxb.BuilderX) MilvusX(f func(mx *MilvusBuilderX)) *sqlxb.BuilderX {
    // 实现...
    return x  // ⭐ 返回 BuilderX，保持链式调用
}
```

---

### 3. 使用 Bb 存储参数

```go
// ✅ 正确：使用 Bb 存储 Milvus 参数
bb := sqlxb.Bb{
    Op:    MILVUS_NPROBE,
    Key:   "nprobe",
    Value: nprobe,
}
mx.builder.Bbs = append(mx.builder.Bbs, bb)
```

---

### 4. 提供扩展点 X()

```go
// ⭐ 必须提供 X() 方法用于未封装的参数
func (mx *MilvusBuilderX) X(k string, v interface{}) *MilvusBuilderX {
    bb := sqlxb.Bb{
        Op:    MILVUS_XX,  // 专属的 XX 操作符
        Key:   k,
        Value: v,
    }
    mx.builder.Bbs = append(mx.builder.Bbs, bb)
    return mx
}

// 使用示例
MilvusX(func(mx *MilvusBuilderX) {
    mx.X("search_k", 100).  // 未封装的参数
        X("ef_construction", 200)
})
```

---

## 💡 最佳实践

### 1. 命名规范

```go
// ✅ 遵循 sqlxb 的 X 后缀命名
QdrantBuilderX   ✅
MilvusBuilderX   ✅
WeaviateBuilderX ✅

// ❌ 不要使用其他命名
MilvusBuilder    ❌
MilvusConfig     ❌
MilvusTemplate   ❌
```

---

### 2. 方法命名风格

```go
// ✅ 简洁命名（无 Set 前缀）
mx.Nprobe(10)          ✅
mx.RoundDecimal(4)     ✅
mx.MetricType("L2")    ✅

// ❌ Java 风格（啰嗦）
mx.SetNprobe(10)       ❌
mx.SetRoundDecimal(4)  ❌
```

---

### 3. 提供快捷方法

```go
// ⭐ 提供高层抽象（快捷方法）
func (mx *MilvusBuilderX) HighAccuracy() *MilvusBuilderX {
    return mx.Nprobe(256).RoundDecimal(6)
}

func (mx *MilvusBuilderX) Balanced() *MilvusBuilderX {
    return mx.Nprobe(64).RoundDecimal(4)
}

func (mx *MilvusBuilderX) HighSpeed() *MilvusBuilderX {
    return mx.Nprobe(16).RoundDecimal(2)
}
```

---

## 🔧 实际案例：Weaviate 支持

### 完整实现

```go
// your_project/vectordb/weaviate_x.go
package vectordb

import "github.com/x-ream/sqlxb"

// Weaviate 专属操作符
const (
    WEAVIATE_CERTAINTY = "WEAVIATE_CERTAINTY"
    WEAVIATE_ALPHA     = "WEAVIATE_ALPHA"
    WEAVIATE_XX        = "WEAVIATE_XX"
)

// WeaviateBuilderX Weaviate 专属构建器
type WeaviateBuilderX struct {
    builder *sqlxb.BuilderX
}

// WeaviateX 创建 Weaviate 专属构建器
func (x *sqlxb.BuilderX) WeaviateX(f func(wx *WeaviateBuilderX)) *sqlxb.BuilderX {
    wx := &WeaviateBuilderX{builder: x}
    f(wx)
    return x
}

// Certainty 设置 Weaviate 的确定性阈值（0-1）
func (wx *WeaviateBuilderX) Certainty(certainty float32) *WeaviateBuilderX {
    if certainty > 0 && certainty <= 1 {
        bb := sqlxb.Bb{
            Op:    WEAVIATE_CERTAINTY,
            Key:   "certainty",
            Value: certainty,
        }
        wx.builder.Bbs = append(wx.builder.Bbs, bb)
    }
    return wx
}

// Alpha 设置混合搜索的权重（0=纯向量, 1=纯关键词）
func (wx *WeaviateBuilderX) Alpha(alpha float32) *WeaviateBuilderX {
    bb := sqlxb.Bb{
        Op:    WEAVIATE_ALPHA,
        Key:   "alpha",
        Value: alpha,
    }
    wx.builder.Bbs = append(wx.builder.Bbs, bb)
    return wx
}

// X 自定义参数
func (wx *WeaviateBuilderX) X(k string, v interface{}) *WeaviateBuilderX {
    bb := sqlxb.Bb{
        Op:    WEAVIATE_XX,
        Key:   k,
        Value: v,
    }
    wx.builder.Bbs = append(wx.builder.Bbs, bb)
    return wx
}

// ToWeaviateGraphQL 转换为 Weaviate GraphQL 查询
func (built *sqlxb.Built) ToWeaviateGraphQL(className string) (string, error) {
    // 1. 提取向量搜索参数
    var queryVector []float32
    var limit int
    
    for _, bb := range built.Conds {
        if bb.Op == sqlxb.VECTOR_SEARCH {
            params := bb.Value.(sqlxb.VectorSearchParams)
            queryVector = params.QueryVector
            limit = params.TopK
            break
        }
    }
    
    // 2. 提取 Weaviate 专属参数
    var certainty float32
    var alpha float32
    
    for _, bb := range built.Conds {
        switch bb.Op {
        case WEAVIATE_CERTAINTY:
            certainty = bb.Value.(float32)
        case WEAVIATE_ALPHA:
            alpha = bb.Value.(float32)
        }
    }
    
    // 3. 构建 GraphQL 查询
    graphql := fmt.Sprintf(`{
  Get {
    %s(
      nearVector: {
        vector: %v
        certainty: %.2f
      }
      limit: %d
    ) {
      _additional {
        certainty
      }
      ... 省略字段
    }
  }
}`, className, queryVector, certainty, limit)
    
    return graphql, nil
}
```

---

## 📖 使用示例

### 示例 1: 同时支持 Qdrant 和 Milvus

```go
package main

import (
    "github.com/x-ream/sqlxb"
    "your-project/vectordb"
)

func search(query string, backend string) (interface{}, error) {
    queryVector := embedQuery(query)
    
    // 构建通用查询
    builder := sqlxb.Of(&CodeVector{}).
        Eq("language", "golang").
        VectorSearch("embedding", queryVector, 20).
        WithHashDiversity("semantic_hash")
    
    // 根据后端选择不同的专属配置
    switch backend {
    case "qdrant":
        built := builder.
            QdrantX(func(qx *sqlxb.QdrantBuilderX) {
                qx.HnswEf(256).ScoreThreshold(0.8)
            }).
            Build()
        return built.ToQdrantJSON()
        
    case "milvus":
        built := builder.
            MilvusX(func(mx *vectordb.MilvusBuilderX) {
                mx.Nprobe(128).RoundDecimal(4)
            }).
            Build()
        return built.ToMilvusJSON("code_vectors")
        
    case "weaviate":
        built := builder.
            WeaviateX(func(wx *vectordb.WeaviateBuilderX) {
                wx.Certainty(0.8).Alpha(0.5)
            }).
            Build()
        return built.ToWeaviateGraphQL("CodeVector")
    }
    
    return nil, fmt.Errorf("unsupported backend: %s", backend)
}
```

---

### 示例 2: 嵌入式轻量向量数据库

```go
// 假设你自研了一个轻量级向量数据库
package vectordb

type LiteVectorBuilderX struct {
    builder *sqlxb.BuilderX
}

func (x *sqlxb.BuilderX) LiteVectorX(f func(lx *LiteVectorBuilderX)) *sqlxb.BuilderX {
    lx := &LiteVectorBuilderX{builder: x}
    f(lx)
    return x
}

// 专属方法
func (lx *LiteVectorBuilderX) CacheSize(size int) *LiteVectorBuilderX {
    // 设置向量缓存大小
    // ...
    return lx
}

func (lx *LiteVectorBuilderX) InMemory(inMemory bool) *LiteVectorBuilderX {
    // 是否全内存运行
    // ...
    return lx
}

// 使用
built := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", vec, 20).
    LiteVectorX(func(lx *LiteVectorBuilderX) {
        lx.InMemory(true).CacheSize(10000)
    }).
    Build()
```

---

## ⚠️ 注意事项

### 1. 不要修改 sqlxb 核心代码

```go
// ❌ 错误：修改 sqlxb 核心
// sqlxb/builder_x.go
func (x *BuilderX) MilvusX(...) {  // ❌ 不要在 sqlxb 内添加
}

// ✅ 正确：在自己的包内扩展
// your_project/vectordb/milvus_x.go
func (x *sqlxb.BuilderX) MilvusX(...) {  // ✅ 在自己包内添加
}
```

---

### 2. 操作符常量使用专属前缀

```go
// ✅ 正确：使用专属前缀避免冲突
const (
    MILVUS_NPROBE = "MILVUS_NPROBE"  // ✅
    WEAVIATE_CERTAINTY = "WEAVIATE_CERTAINTY"  // ✅
)

// ❌ 错误：可能与 sqlxb 冲突
const (
    NPROBE = "NPROBE"  // ❌ 太通用
)
```

---

### 3. 优雅降级处理

```go
// ⭐ 如果在 PostgreSQL 环境，Milvus 参数应被忽略
func (built *sqlxb.Built) SqlOfVectorSearch() (string, []interface{}) {
    // 自动忽略 MILVUS_* 操作符
    for _, bb := range built.Conds {
        if strings.HasPrefix(bb.Op, "MILVUS_") {
            continue  // ⭐ 忽略
        }
        // ...
    }
}
```

---

## 📊 支持的向量数据库对比

| 数据库 | 官方支持 | 社区扩展 | 实现难度 | 推荐度 |
|-------|---------|---------|---------|--------|
| **Qdrant** | ✅ (v0.9.0+) | - | - | ⭐⭐⭐⭐⭐ |
| **Milvus** | ❌ | 本文档 | 中等 | ⭐⭐⭐⭐ |
| **Weaviate** | ❌ | 本文档 | 中等 | ⭐⭐⭐ |
| **Pinecone** | ❌ | 可自行实现 | 简单 | ⭐⭐⭐ |
| **pgvector** | ✅ (v0.8.1+) | - | - | ⭐⭐⭐⭐⭐ |
| **自研** | ❌ | 本文档 | 高 | ⭐⭐⭐⭐⭐ |

---

## 🚀 项目结构建议

```
your-project/
├── go.mod
├── vectordb/
│   ├── milvus_x.go              # Milvus 扩展
│   ├── milvus_oper.go           # Milvus 操作符
│   ├── to_milvus_json.go        # JSON 转换
│   ├── milvus_test.go           # 测试
│   │
│   ├── weaviate_x.go            # Weaviate 扩展
│   ├── to_weaviate_graphql.go   # GraphQL 转换
│   │
│   └── lite_vector_x.go         # 自研向量数据库
│
└── main.go
```

---

## 🎯 总结

### 实现自定义向量数据库支持的 5 步

1. ✅ 创建 `XxxxBuilderX` 结构体
2. ✅ 定义专属操作符常量（`XXXX_*`）
3. ✅ 实现专属配置方法
4. ✅ 实现 JSON/GraphQL 转换器
5. ✅ 编写测试用例

### 核心原则

```
1. 清晰分离：通用方法在外部，专属配置在内部
2. 向后兼容：通过扩展而非修改
3. 使用 Bb：所有参数存储为 Bb
4. 提供 X()：支持未封装的参数
5. 遵循风格：简洁命名，链式调用
```

---

**参考实现**: [qdrant_x.go](../qdrant_x.go) 和 [to_qdrant_json.go](../to_qdrant_json.go)

**开始构建你自己的向量数据库支持！** 🚀


