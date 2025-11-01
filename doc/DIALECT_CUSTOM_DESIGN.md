# Dialect + Custom 设计：统一的 JsonOfSelect()

## 🎯 核心革新

通过 **Dialect（方言）+ Custom（自定义）** 两层抽象，实现真正统一的 API：

```go
// ✅ 理想：统一接口，不需要指定数据库类型
built := xb.C().
    WithCustom(qdrantCustom).  // 设置 Custom
    VectorSearch(...).
    Build()

json, _ := built.JsonOfSelect()  // ⭐ 自动根据 Custom 生成对应的 JSON
```

---

## 📊 两层抽象

### 1️⃣ Dialect（方言）

数据库类型标识：

```go
type Dialect string

const (
    // SQL 数据库
    PostgreSQL Dialect = "postgresql"
    MySQL      Dialect = "mysql"
    SQLite     Dialect = "sqlite"

    // 向量数据库
    Qdrant   Dialect = "qdrant"
    Milvus   Dialect = "milvus"
    Weaviate Dialect = "weaviate"
    Pinecone Dialect = "pinecone"
)
```

### 2️⃣ Custom（自定义配置）

数据库专属逻辑封装：

```go
type Custom interface {
    // 获取方言类型
    GetDialect() Dialect

    // 应用专属参数
    ApplyParams(bbs []Bb, req interface{}) error

    // 生成 JSON
    ToJSON(built *Built) (string, error)
}
```

---

## 🚀 使用示例

### Qdrant 用户

```go
// Step 1: 选择 Qdrant Custom（预设配置）
built := xb.C().
    WithCustom(xb.QdrantHighPrecision()).  // 高精度模式
    VectorScoreThreshold(0.8).
    VectorSearch("code_vectors", "embedding", vec, 20, xb.CosineDistance).
    Build()

// Step 2: 统一接口生成 JSON
json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Qdrant
```

**预设模式**:
- `xb.QdrantHighPrecision()` - 高精度（HnswEf=512）
- `xb.QdrantHighSpeed()` - 高速（HnswEf=32）
- `xb.QdrantBalanced()` - 平衡（HnswEf=128，默认）

### Milvus 用户

```go
// Step 1: 选择 Milvus Custom
built := xb.C().
    WithCustom(xb.NewMilvusCustom()).  // Milvus 默认配置
    VectorScoreThreshold(0.8).
    MilvusNProbe(64).
    VectorSearch("code_vectors", "embedding", vec, 20, xb.L2Distance).
    Build()

// Step 2: 统一接口生成 JSON
json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Milvus
```

### 跨数据库部署

```go
// 同一份业务逻辑，根据配置切换数据库
func SearchCodeVectors(config Config, embedding []float32) ([]Result, error) {
    // Step 1: 根据配置选择 Custom
    var custom xb.Custom
    switch config.VectorDB {
    case "qdrant":
        custom = xb.QdrantBalanced()
    case "milvus":
        custom = xb.NewMilvusCustom()
    case "weaviate":
        custom = xb.NewWeaviateCustom()
    }

    // Step 2: 构建查询（完全相同的代码）
    built := xb.C().
        WithCustom(custom).  // ⭐ 唯一的区别
        VectorScoreThreshold(0.8).
        VectorSearch("code_vectors", "embedding", embedding, 20, xb.CosineDistance).
        Build()

    // Step 3: 统一接口生成 JSON
    json, _ := built.JsonOfSelect()  // ⭐ 自动适配不同数据库

    // Step 4: 调用对应的客户端
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

## 🎨 对比：设计演进

### v0.10.x 之前（数据库专用方法）

```go
// ❌ 问题：需要为每个数据库写不同的方法名
qdrantJSON, _ := built.JsonOfQdrantSelect()
milvusJSON, _ := built.JsonOfMilvusSelect()
weaviateJSON, _ := built.JsonOfWeaviateSelect()
```

### v0.11.0（Dialect + Custom 设计）

```go
// ✅ 优势：统一的方法名，通过 Custom 区分
built := xb.C().
    WithCustom(qdrantCustom).  // 或 milvusCustom
    Build()

json, _ := built.JsonOfSelect()  // ⭐ 自动适配
```

---

## 🔄 架构设计

### 数据流

```
用户代码
   ↓
WithCustom(custom)  ← 设置 Custom
   ↓
Build()            ← 传递 Custom 到 Built
   ↓
JsonOfSelect()     ← 调用 custom.ToJSON(built)
   ↓
Custom.ToJSON()    ← Qdrant/Milvus 专属逻辑
   ↓
JSON 字符串
```

### 类图

```
┌─────────────────────────────────┐
│  Custom (interface)             │
├─────────────────────────────────┤
│  GetDialect() Dialect           │
│  ApplyParams(bbs, req) error    │
│  ToJSON(built) (string, error)  │
└─────────────────────────────────┘
            ↑
            │ 实现
            │
   ┌────────┴────────┬────────────┬─────────────┐
   │                 │            │             │
┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│ Qdrant   │  │ Milvus   │  │ Weaviate │  │ Pinecone │
│ Custom   │  │ Custom   │  │ Custom   │  │ Custom   │
└──────────┘  └──────────┘  └──────────┘  └──────────┘
```

---

## 💡 实现 Custom 的步骤

### 示例：添加 Weaviate 支持

#### Step 1: 定义 Weaviate Custom

```go
// weaviate_custom.go
type WeaviateCustom struct {
    DefaultCertainty float32
}

func NewWeaviateCustom() *WeaviateCustom {
    return &WeaviateCustom{
        DefaultCertainty: 0.7,
    }
}

func (c *WeaviateCustom) GetDialect() Dialect {
    return Weaviate
}

func (c *WeaviateCustom) ApplyParams(bbs []Bb, req interface{}) error {
    weaviateReq, ok := req.(*WeaviateSearchRequest)
    if !ok {
        return fmt.Errorf("req is not WeaviateSearchRequest")
    }

    // 应用 Weaviate 专属参数
    applyWeaviateParams(bbs, weaviateReq)
    return nil
}

func (c *WeaviateCustom) ToJSON(built *Built) (string, error) {
    // 从 Built.Conds 中提取参数
    req, err := built.ToWeaviateRequest()
    if err != nil {
        return "", err
    }

    // 应用参数
    c.ApplyParams(built.Conds, req)

    // 序列化
    return weaviateMergeAndSerialize(req, built.Conds)
}
```

#### Step 2: 用户使用

```go
built := xb.C().
    WithCustom(xb.NewWeaviateCustom()).
    VectorSearch(...).
    Build()

json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Weaviate
```

---

## 📈 优势总结

| 特性 | v0.10.x（专用方法） | v0.11.0（Custom 设计） |
|------|---------------------|----------------------|
| **方法名** | `JsonOfQdrantSelect()` | `JsonOfSelect()` ✅ |
| **切换数据库** | 修改方法名 | 修改 Custom ✅ |
| **预设配置** | 无 | `QdrantHighPrecision()` 等 ✅ |
| **扩展性** | 每个数据库新增方法 | 实现 Custom 接口 ✅ |
| **跨数据库部署** | 需要 if/else 判断方法 | 统一 `JsonOfSelect()` ✅ |
| **学习成本** | 需要记住每个方法名 | 只需学习 Custom ✅ |

---

## 🎯 设计原则

### 1. YAGNI（You Aren't Gonna Need It）

- 只实现当前需要的数据库
- Custom 接口提供扩展能力

### 2. 开闭原则

- 对扩展开放：新增数据库实现 Custom
- 对修改封闭：`JsonOfSelect()` 不需要修改

### 3. 依赖倒置

- 高层模块（Built）依赖抽象（Custom）
- 低层模块（QdrantCustom）实现抽象

---

## 🚧 兼容性

### 向后兼容

```go
// ✅ 旧代码仍然可用
json, _ := built.JsonOfQdrantSelect()

// ✅ 新代码更简洁
built := xb.C().
    WithCustom(xb.QdrantBalanced()).
    Build()

json, _ := built.JsonOfSelect()
```

### 推荐迁移路径

```go
// v0.10.x
json, _ := built.JsonOfQdrantSelect()

// ↓ 迁移到 v0.11.0

// v0.11.0（推荐）
built := xb.C().
    WithCustom(xb.QdrantBalanced()).
    Build()

json, _ := built.JsonOfSelect()
```

---

## 📚 相关文档

- `dialect.go` - Dialect 和 Custom 接口定义
- `qdrant_custom.go` - Qdrant Custom 实现
- `NAMING_CONVENTION.md` - 命名规范
- `TO_JSON_DESIGN_CLARIFICATION.md` - JsonOf 设计说明

---

## 🎉 总结

**Dialect + Custom 设计实现了：**

1. ✅ **统一 API**: `JsonOfSelect()` 适用于所有向量数据库
2. ✅ **配置预设**: `QdrantHighPrecision()` 等预设模式
3. ✅ **轻松切换**: 只需修改 Custom，业务代码不变
4. ✅ **易于扩展**: 新增数据库只需实现 Custom 接口
5. ✅ **向后兼容**: 旧代码仍然可用
6. ✅ **类型安全**: 编译时检查，运行时零错误

**这是 xb 框架的重大革新！** 🚀

---

**版本**: v0.11.0  
**更新时间**: 2025-11-01  
**维护者**: xb Team

