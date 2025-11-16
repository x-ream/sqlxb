# Custom 快速开始

本文档是 `xb/doc/CUSTOM_QUICKSTART.md` 的中文版本。它引导你在不到 30 分钟内实现新适配器。

---

## 1. 骨架

```go
type MyVectorCustom struct {
    Endpoint string
}

func NewMyVectorCustom(endpoint string) *MyVectorCustom {
    return &MyVectorCustom{Endpoint: endpoint}
}

func (c *MyVectorCustom) Generate(built *xb.Built) (any, error) {
    req := buildPayload(built) // 将条件转换为 JSON
    return req, nil
}
```

---

## 2. 构建器辅助方法

- `WithCollection(name string)`
- `WithPayloadSelector(selector any)`
- `WithScoreThreshold(th float32)`

在你的自定义结构体上暴露辅助方法，以便调用者不接触原始字段。

---

## 3. 测试

```go
func TestMyVectorCustom_Select(t *testing.T) {
    built := xb.Of("vectors").
        Eq("tenant_id", 42).
        VectorSearch("embedding", xb.Vector{0.1, 0.2}, 5).
        Build()

    custom := NewMyVectorCustom("http://localhost:9000")
    payload, err := custom.Generate(built)
    require.NoError(t, err)
    snapshot.AssertJSON(t, payload)
}
```

---

## 4. 文档

- 描述预设（`NewMyVectorCustom`、`MyVectorCustomHighPrecision`）。
- 链接到 JSON 样本和行为表。
- 如果支持，解释与 `JsonOfInsert/Update/Delete` 的兼容性。

---

## 5. 下一步

- 发布你的适配器或保持内部。
- 为拦截器添加指标/日志记录以进行可观测性。
- 将经验报告贡献给 `doc/cn/CUSTOM_VECTOR_DB_GUIDE.md`。

