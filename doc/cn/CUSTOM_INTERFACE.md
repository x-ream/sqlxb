# Custom 接口（中文）

`Custom` 接口是 xb 扩展能力的核心。框架不再内置大量方言，而是让用户在几分钟内实现任意数据库（SQL 或向量）的序列化逻辑。

---

## 1. 接口定义

```go
type Custom interface {
    Generate(built *xb.Built) (any, error)
}
```

- `built` 包含 Builder 产出的 AST（select、from、conditions 等）。
- 返回字符串（JSON/SQL）或自定义结构体，供驱动层消费。
- 错误会直接传递给 `built.JsonOfSelect()` 或 `built.CustomGenerate()`。

---

## 2. 最小示例

```go
type MyVectorCustom struct{}

func (c *MyVectorCustom) Generate(built *xb.Built) (any, error) {
    payload := map[string]any{
        "collection": built.Table,
        "filter":     built.FilterJSON(),
        "vector":     built.Vector,
    }
    return payload, nil
}

json, err := xb.Of("vectors").
    Custom(&MyVectorCustom{}).
    VectorSearch("embedding", vec, 10).
    Build().
    JsonOfSelect()
```

---

## 3. 设计目标

| 目标 | 说明 |
|------|------|
| 极简接口 | 一个接口、一个方法 |
| 隔离性 | 用户可独立迭代，无需 fork xb |
| 性能 | 接口调用约 1ns，无额外分支 |
| 面向未来 | 新特性出现时，用户可立即实现 |

---

## 4. 推荐结构

```go
type MilvusCustom struct {
    Endpoint string
    Timeout  time.Duration
}

func NewMilvusCustom() *MilvusCustom {
    return &MilvusCustom{
        Endpoint: "http://localhost:19530",
        Timeout:  3 * time.Second,
    }
}

func (c *MilvusCustom) Generate(built *xb.Built) (any, error) {
    req := buildMilvusPayload(built)
    return req, nil
}
```

- 提供构造函数（默认/高精度/高速）供用户选择。
- 在注释中说明各 preset 的使用场景。

---

## 5. 测试建议

```go
func TestMilvusCustom_Generate(t *testing.T) {
    built := xb.Of("code_vectors").
        VectorSearch("embedding", xb.Vector{0.1, 0.2}, 5).
        Build()

    custom := NewMilvusCustom()
    payload, err := custom.Generate(built)

    require.NoError(t, err)
    snapshot.AssertJSON(t, payload)
}
```

- 测试放在你自己的项目中，xb 无需引用你的 Custom。
- 可以使用 snapshot/golden 文件保证 JSON 不被意外改动。

---

## 6. 适用场景

| 场景 | 示例 |
|------|------|
| 向量数据库 JSON | Qdrant（官方）、Milvus、Pinecone |
| SQL 方言扩展 | Oracle upsert、ClickHouse `FORMAT` |
| 内部 API | 公司自研搜索服务 |
| 混合模式 | 同一 builder 输出 SQL + JSON |

---

## 7. 相关文档

- `QDRANT_GUIDE.md`：官方实现
- `VECTOR_GUIDE.md`：嵌入实践
- `FILTERING.md`：理解自动过滤

若你发布了可复用的 Custom 适配，欢迎在 `doc/en/` 或 `doc/cn/` 中追加说明并提交 PR。

