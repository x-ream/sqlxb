# 自定义向量数据库支持指南 (v1.1.0)

## 🎯 概述

本指南演示如何为 `xb` 添加自定义向量数据库支持（如 Milvus, Weaviate, Pinecone 等）。

**核心思路**：实现 `Custom` 接口，提供数据库专属的 JSON 生成逻辑。

---

## 🚀 快速开始

### Custom 接口（极简设计）

```go
// 定义在 xb/dialect.go
type Custom interface {
    // ToJSON 生成查询 JSON
    // 参数: built - Built 对象（包含所有查询条件）
    // 返回: JSON 字符串, error
    ToJSON(built *Built) (string, error)
}
```

**就这一个方法！** 简单、直接、实用。

---

## 📋 实现步骤（以 Milvus 为例）

### Step 1: 定义 Milvus Custom

```go
// milvus_custom.go
package xb

// MilvusCustom Milvus 专属配置
type MilvusCustom struct {
    // 默认参数
    DefaultNProbe     int
    DefaultRoundDec   int
    DefaultMetricType string
}

// NewMilvusCustom 创建 Milvus Custom（默认配置）
func NewMilvusCustom() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:     64,
        DefaultRoundDec:   4,
        DefaultMetricType: "L2",
    }
}

// ToJSON 实现 Custom 接口
func (c *MilvusCustom) ToJSON(built *Built) (string, error) {
    // 委托给内部实现
    return built.toMilvusJSON()
}
```

---

### Step 2: 实现 JSON 生成逻辑

```go
// to_milvus_json.go（在 xb 包内或自己的项目中）
package xb

import (
    "encoding/json"
    "fmt"
)

// MilvusSearchRequest Milvus 搜索请求结构
type MilvusSearchRequest struct {
    CollectionName string          `json:"collection_name"`
    Data           [][]float32     `json:"data"`
    Limit          int             `json:"limit"`
    SearchParams   MilvusSearchParams `json:"search_params"`
    Expr           string          `json:"expr,omitempty"`
}

type MilvusSearchParams struct {
    MetricType   string                 `json:"metric_type"`
    Params       map[string]interface{} `json:"params"`
    RoundDecimal int                    `json:"round_decimal,omitempty"`
}

// toMilvusJSON 内部实现（私有方法）
func (built *Built) toMilvusJSON() (string, error) {
    // 1. 从 Built.Conds 中提取 VectorSearch 参数
    vectorBb := findVectorSearchBb(built.Conds)
    if vectorBb == nil {
        return "", fmt.Errorf("no VECTOR_SEARCH found")
    }
    
    params := vectorBb.Value.(VectorSearchParams)
    
    // 2. 创建 Milvus 请求对象
    req := &MilvusSearchRequest{
        CollectionName: params.TableName,
        Data:           [][]float32{params.Vector},
        Limit:          params.Limit,
        SearchParams: MilvusSearchParams{
            MetricType: milvusDistanceMetric(params.Distance),
            Params:     make(map[string]interface{}),
        },
    }
    
    // 3. 应用 Milvus 专属参数
    applyMilvusParams(built.Conds, req)
    
    // 4. 序列化为 JSON
    bytes, err := json.MarshalIndent(req, "", "  ")
    if err != nil {
        return "", fmt.Errorf("failed to marshal Milvus request: %w", err)
    }
    
    return string(bytes), nil
}

// applyMilvusParams 应用 Milvus 专属参数
func applyMilvusParams(bbs []Bb, req *MilvusSearchRequest) {
    for _, bb := range bbs {
        switch bb.Op {
        case "MILVUS_NPROBE":
            req.SearchParams.Params["nprobe"] = bb.Value
        case "MILVUS_ROUND_DEC":
            req.SearchParams.RoundDecimal = bb.Value.(int)
        case "MILVUS_METRIC_TYPE":
            req.SearchParams.MetricType = bb.Value.(string)
        }
    }
}

func milvusDistanceMetric(metric VectorDistance) string {
    switch metric {
    case CosineDistance:
        return "IP"  // Inner Product
    case L2Distance:
        return "L2"
    case InnerProduct:
        return "IP"
    default:
        return "L2"
    }
}
```

---

### Step 3: 添加 Builder 方法（可选）

```go
// cond_builder_milvus.go
package xb

// MilvusNProbe 设置 Milvus nprobe 参数
func (b *CondBuilder) MilvusNProbe(nprobe int) *CondBuilder {
    return b.append(Bb{Op: "MILVUS_NPROBE", Value: nprobe})
}

// MilvusRoundDec 设置小数位
func (b *CondBuilder) MilvusRoundDec(dec int) *CondBuilder {
    return b.append(Bb{Op: "MILVUS_ROUND_DEC", Value: dec})
}

// MilvusX 自定义参数
func (b *CondBuilder) MilvusX(key string, value interface{}) *CondBuilder {
    return b.append(Bb{Op: "MILVUS_XX", Key: key, Value: value})
}
```

---

## 💡 使用示例

### 方式 1: 使用 Custom（推荐）

```go
// Milvus
built := xb.Of("code_vectors").
    Custom(xb.NewMilvusCustom()).  // ⭐ 设置 Milvus Custom
    VectorSearch("embedding", vec, 20).
    Eq("language", "golang").
    Build()

json, _ := built.JsonOfSelect()  // ⭐ 统一接口
```

### 方式 2: 便捷方法

```go
// Milvus（如果实现了 ToMilvusJSON 便捷方法）
built := xb.Of("code_vectors").
    MilvusNProbe(64).
    VectorSearch("embedding", vec, 20).
    Build()

json, _ := built.ToMilvusJSON()  // 自动使用默认 Custom
```

### 方式 3: 运行时切换

```go
// 根据配置动态选择数据库
var custom xb.Custom
switch config.VectorDB {
case "qdrant":
    custom = xb.NewQdrantCustom()
case "milvus":
    custom = xb.NewMilvusCustom()
case "weaviate":
    custom = xb.NewWeaviateCustom()
}

built := xb.Of("code_vectors").
    Custom(custom).  // ⭐ 运行时切换
    VectorSearch("embedding", vec, 20).
    Build()

json, _ := built.JsonOfSelect()  // ✅ 自动适配
```

---

## 🎨 设计模式对比

### v1.0.x（旧模式）：BuilderX 扩展

```go
// ❌ 复杂：需要定义专属 BuilderX
type MilvusBuilderX struct {
    builder *xb.BuilderX
}

func (x *xb.BuilderX) MilvusX(f func(mx *MilvusBuilderX)) *xb.BuilderX {
    mx := &MilvusBuilderX{builder: x}
    f(mx)
    return x
}

// 使用
built := xb.Of("t").
    MilvusX(func(mx *MilvusBuilderX) {
        mx.Nprobe(64).RoundDec(4)
    }).
    Build()
```

### v1.1.0（新模式）：Custom 接口

```go
// ✅ 简单：只需实现一个接口
type MilvusCustom struct {
    DefaultNProbe int
}

func (c *MilvusCustom) ToJSON(built *Built) (string, error) {
    return built.toMilvusJSON()
}

// 使用
built := xb.Of("t").
    Custom(xb.NewMilvusCustom()).
    Build()

json, _ := built.JsonOfSelect()  // 统一接口
```

---

## 📊 完整示例：Weaviate 支持

### 1. 定义 Weaviate Custom

```go
// weaviate_custom.go
package xb

type WeaviateCustom struct {
    DefaultCertainty float32
    DefaultAlpha     float32
}

func NewWeaviateCustom() *WeaviateCustom {
    return &WeaviateCustom{
        DefaultCertainty: 0.7,
        DefaultAlpha:     0.5,
    }
}

func (c *WeaviateCustom) ToJSON(built *Built) (string, error) {
    return built.toWeaviateJSON()
}

// 预设模式
func WeaviateSemanticMode() *WeaviateCustom {
    return &WeaviateCustom{
        DefaultCertainty: 0.8,
        DefaultAlpha:     0.0,  // 纯向量搜索
    }
}

func WeaviateHybridMode() *WeaviateCustom {
    return &WeaviateCustom{
        DefaultCertainty: 0.7,
        DefaultAlpha:     0.5,  // 混合搜索
    }
}
```

### 2. 实现 JSON 生成

```go
// to_weaviate_json.go
func (built *Built) toWeaviateJSON() (string, error) {
    // 提取参数
    vectorBb := findVectorSearchBb(built.Conds)
    if vectorBb == nil {
        return "", fmt.Errorf("no VECTOR_SEARCH found")
    }
    
    params := vectorBb.Value.(VectorSearchParams)
    
    // 构建 Weaviate GraphQL 查询
    query := fmt.Sprintf(`{
  Get {
    %s(
      nearVector: {
        vector: %v
      }
      limit: %d
    ) {
      _additional { certainty }
      # 字段列表
    }
  }
}`, params.TableName, params.Vector, params.Limit)
    
    return query, nil
}
```

### 3. 使用

```go
built := xb.Of("CodeVector").
    Custom(xb.WeaviateSemanticMode()).
    VectorSearch("embedding", vec, 20).
    Build()

graphql, _ := built.JsonOfSelect()
```

---

## 🔄 对比：Qdrant vs Milvus vs Weaviate

### 统一的 API

```go
// ⭐ 完全相同的调用方式
built := xb.Of("code_vectors").
    VectorSearch("embedding", vec, 20).
    Eq("language", "golang").
    Build()

// ⭐ 只需切换 Custom
qdrantJSON, _ := built.Custom(xb.NewQdrantCustom()).JsonOfSelect()
milvusJSON, _ := built.Custom(xb.NewMilvusCustom()).JsonOfSelect()
weaviateJSON, _ := built.Custom(xb.NewWeaviateCustom()).JsonOfSelect()
```

---

## 📝 检查清单

添加新的向量数据库支持时，请确保：

- [ ] **定义 Custom 结构体**（如 `MilvusCustom`）
- [ ] **实现 ToJSON 方法**（实现 Custom 接口）
- [ ] **创建内部实现**（如 `toMilvusJSON()`）
- [ ] **提供预设模式**（如 `NewMilvusCustom()`、`MilvusHighPrecision()`）
- [ ] **添加便捷方法**（可选：`ToMilvusJSON()` 自动使用默认 Custom）
- [ ] **编写测试用例**（验证 Custom 接口和 JSON 生成）
- [ ] **文档注释完整**

---

## 🎯 最佳实践

### 1. Custom 结构体设计

```go
// ✅ 好的设计：包含默认配置
type MilvusCustom struct {
    DefaultNProbe     int     // 用户可以自定义
    DefaultRoundDec   int
    DefaultMetricType string
}

// ❌ 不好的设计：空结构体
type MilvusCustom struct {
    // 什么都没有
}
```

### 2. 提供预设模式

```go
// ✅ 必须提供
func NewMilvusCustom() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:     64,
        DefaultRoundDec:   4,
        DefaultMetricType: "L2",
    }
}

// ✅ 推荐提供多个预设
func MilvusHighPrecision() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:     256,
        DefaultRoundDec:   6,
        DefaultMetricType: "IP",
    }
}

func MilvusHighSpeed() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:     16,
        DefaultRoundDec:   2,
        DefaultMetricType: "L2",
    }
}
```

### 3. 内部实现分离

```go
// ✅ 公开 Custom
type MilvusCustom struct { ... }

func (c *MilvusCustom) ToJSON(built *Built) (string, error) {
    return built.toMilvusJSON()  // ⭐ 委托给私有实现
}

// ✅ 私有实现（小写开头）
func (built *Built) toMilvusJSON() (string, error) {
    // 实际的 JSON 生成逻辑
}

// ✅ 便捷方法（可选）
func (built *Built) ToMilvusJSON() (string, error) {
    if built.Custom != nil {
        return built.JsonOfSelect()
    }
    built.Custom = NewMilvusCustom()  // 自动设置默认 Custom
    return built.JsonOfSelect()
}
```

---

## 🔧 实战：完整的 Milvus 实现

### 完整代码（约 150 行）

```go
// ============================================================================
// milvus_custom.go
// ============================================================================
package xb

type MilvusCustom struct {
    DefaultNProbe     int
    DefaultRoundDec   int
    DefaultMetricType string
}

func NewMilvusCustom() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:     64,
        DefaultRoundDec:   4,
        DefaultMetricType: "L2",
    }
}

func (c *MilvusCustom) ToJSON(built *Built) (string, error) {
    return built.toMilvusJSON()
}

func MilvusHighPrecision() *MilvusCustom {
    return &MilvusCustom{DefaultNProbe: 256, DefaultRoundDec: 6}
}

func MilvusHighSpeed() *MilvusCustom {
    return &MilvusCustom{DefaultNProbe: 16, DefaultRoundDec: 2}
}

// ============================================================================
// to_milvus_json.go
// ============================================================================

// ToMilvusJSON 便捷方法
func (built *Built) ToMilvusJSON() (string, error) {
    if built.Custom != nil {
        return built.JsonOfSelect()
    }
    built.Custom = NewMilvusCustom()
    return built.JsonOfSelect()
}

// toMilvusJSON 内部实现
func (built *Built) toMilvusJSON() (string, error) {
    vectorBb := findVectorSearchBb(built.Conds)
    if vectorBb == nil {
        return "", fmt.Errorf("no VECTOR_SEARCH found")
    }
    
    params := vectorBb.Value.(VectorSearchParams)
    
    req := &MilvusSearchRequest{
        CollectionName: params.TableName,
        Data:           [][]float32{params.Vector},
        Limit:          params.Limit,
        SearchParams: MilvusSearchParams{
            MetricType: milvusDistanceMetric(params.Distance),
            Params:     make(map[string]interface{}),
        },
    }
    
    // 应用专属参数
    applyMilvusParams(built.Conds, req)
    
    // 序列化
    bytes, err := json.MarshalIndent(req, "", "  ")
    if err != nil {
        return "", err
    }
    
    return string(bytes), nil
}
```

---

## 📖 使用示例

### 示例 1: 基础用法

```go
// Milvus 搜索
built := xb.Of("code_vectors").
    Custom(xb.NewMilvusCustom()).
    VectorSearch("embedding", queryVector, 20).
    Eq("language", "golang").
    Build()

json, _ := built.JsonOfSelect()
```

### 示例 2: 预设模式

```go
// 高精度模式
built := xb.Of("code_vectors").
    Custom(xb.MilvusHighPrecision()).
    VectorSearch("embedding", vec, 20).
    Build()

json, _ := built.JsonOfSelect()
```

### 示例 3: 便捷方法

```go
// 使用便捷方法（自动使用默认 Custom）
built := xb.Of("code_vectors").
    VectorSearch("embedding", vec, 20).
    Build()

json, _ := built.ToMilvusJSON()
```

### 示例 4: 跨数据库部署

```go
func SearchDocuments(config Config, query string) ([]Document, error) {
    embedding := embed(query)
    
    // 根据配置选择 Custom
    var custom xb.Custom
    switch config.VectorDB {
    case "qdrant":
        custom = xb.NewQdrantCustom()
    case "milvus":
        custom = xb.NewMilvusCustom()
    case "weaviate":
        custom = xb.NewWeaviateCustom()
    }
    
    // 统一的查询构建
    built := xb.Of("documents").
        Custom(custom).
        VectorSearch("embedding", embedding, 10).
        Eq("status", "published").
        Build()
    
    // 统一的接口
    json, _ := built.JsonOfSelect()
    
    // 调用对应的客户端
    switch config.VectorDB {
    case "qdrant":
        return qdrantClient.Search(json)
    case "milvus":
        return milvusClient.Search(json)
    case "weaviate":
        return weaviateClient.Search(json)
    }
}
```

---

## 🎯 设计优势

### v1.1.0 Custom 接口 vs v1.0.x BuilderX 扩展

| 特性 | v1.0.x (BuilderX) | v1.1.0 (Custom) |
|------|------------------|-----------------|
| **接口方法数** | 需要多个方法 | 1个方法 ✅ |
| **代码量** | ~300 行 | ~150 行 ✅ |
| **预设模式** | 不支持 | 支持 ✅ |
| **运行时切换** | 困难 | 简单 ✅ |
| **统一 API** | `ToMilvusJSON()` | `JsonOfSelect()` ✅ |
| **类型复杂度** | 高 | 低 ✅ |

---

## 📚 参考实现

### Qdrant Custom（官方实现）

查看 `xb/qdrant_custom.go`：

```go
type QdrantCustom struct {
    DefaultHnswEf         int
    DefaultScoreThreshold float32
    DefaultWithVector     bool
}

func (c *QdrantCustom) ToJSON(built *Built) (string, error) {
    return built.toQdrantJSON()
}

// 使用说明：
// 1. 基础构造函数：NewQdrantCustom()
// 2. 手动配置字段或使用 QdrantX() 闭包
```

---

## ⚠️ 注意事项

### 1. 不要在 xb 核心添加所有数据库的支持

```go
// ❌ 错误：在 xb 核心添加所有数据库
// xb/milvus_custom.go ❌
// xb/weaviate_custom.go ❌
// xb/pinecone_custom.go ❌

// ✅ 正确：只添加常用的（如 Qdrant）
// xb/qdrant_custom.go ✅

// ✅ 其他数据库在用户项目中实现
// your-project/vectordb/milvus_custom.go ✅
```

### 2. Custom 接口只有一个方法

```go
// ✅ 保持简单
type Custom interface {
    ToJSON(built *Built) (string, error)
}

// ❌ 不要过度设计
type Custom interface {
    GetDialect() Dialect          // ❌ 多余
    ApplyParams(bbs, req) error   // ❌ 多余
    ToJSON(built) (string, error) // ✅ 只需这个
}
```

### 3. 类型本身就是标识

```go
// ✅ Go 的接口多态
var custom xb.Custom

custom = &QdrantCustom{...}   // 类型本身说明是 Qdrant
custom = &MilvusCustom{...}   // 类型本身说明是 Milvus

// ❌ 不需要额外的枚举
type Dialect string
const Qdrant Dialect = "qdrant"  // ❌ 多余
```

---

## 🎉 总结

### 添加自定义向量数据库支持的 3 步

1. ✅ **定义 Custom**：实现 `ToJSON(built *Built) (string, error)`
2. ✅ **创建预设**：提供 `NewXxxCustom()` 和预设模式
3. ✅ **编写测试**：验证功能正常

### 核心优势

- ✅ **极简接口**：只需一个方法
- ✅ **类型安全**：编译时检查
- ✅ **预设模式**：开箱即用
- ✅ **运行时切换**：灵活部署
- ✅ **统一 API**：`JsonOfSelect()` 适用于所有向量数据库

---

**参考**：
- `xb/dialect.go` - Custom 接口定义
- `xb/qdrant_custom.go` - Qdrant 官方实现
- `xb/doc/MILVUS_TEMPLATE.go` - Milvus 实现模板
- `xb/doc/DIALECT_CUSTOM_DESIGN.md` - Custom 设计文档

**开始实现你的向量数据库支持！** 🚀
