# Custom 接口 README

本文档是 `xb/doc/CUSTOM_INTERFACE_README.md` 的中文版本，作为与 `Custom` 相关的所有内容的英文入口页面。

---

## 为什么选择 Custom？

- 数据库生态系统的发展速度比任何单个 ORM 都能跟上的速度更快。
- 单一接口让我们能够覆盖关系型 SQL、向量存储、时间序列和专有服务。
- 用户可以在几分钟内发布适配器，而不是几周。

---

## 快速开始

```go
type MyDBCustom struct{}

func (c *MyDBCustom) Generate(built *xb.Built) (any, error) {
    return map[string]any{
        "query": built.Table,
        "filter": built.FilterJSON(),
    }, nil
}

json, _ := xb.Of("docs").
    Custom(&MyDBCustom{}).
    Eq("tenant_id", tenant).
    Build().
    JsonOfSelect()
```

---

## 建议阅读顺序

1. `doc/cn/CUSTOM_INTERFACE_PHILOSOPHY.md`
2. `doc/cn/CUSTOM_QUICKSTART.md`
3. `doc/cn/CUSTOM_VECTOR_DB_GUIDE.md`
4. `doc/cn/CUSTOM_ARCHITECTURE_VALIDATION.md`

---

## 社区适配器

- Qdrant（官方）– 参见 `xb/qdrant_custom.go`
- Milvus、Weaviate、Pinecone – 社区实验
- 内部服务 – 团队通常包装 HTTP/GRPC 端点

一旦你的适配器稳定，请在 `doc/cn/CUSTOM_VECTOR_DB_GUIDE.md` 中记录它，以便其他人可以从中学习。

