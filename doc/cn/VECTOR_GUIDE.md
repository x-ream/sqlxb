# 向量指南（中文）

本指南汇总在 xb 中实现 SQL + 向量混合查询时的经验，可与 `QDRANT_GUIDE.md` 搭配使用。

---

## 1. 嵌入向量健康度

- 在调用 `VectorSearch` 之前做 L2 归一化，xb 不会自动处理。
- 存储向量时务必记录维度，`VectorSearch("embedding", vec, limit)` 不会检查长度。
- 多向量文档（如文本 + 代码）可使用 `VectorSearchMulti`，或构建多个 builder 再 `UNION ALL`。

---

## 2. 典型查询流程

```go
built := xb.Of(&CodeVector{}).
    Eq("language", "go").
    VectorSearch("embedding", queryVector, 20).
    Sort("score", xb.Desc).
    Offset(0).
    Build()

json, err := built.JsonOfSelect()
```

- 条件与 SQL 完全一致，会被映射到 Qdrant `filter.must`。
- 可以显式调用 `Sort("score", xb.Desc)`，方便与 SQL 语句保持一致。

---

## 3. 常见混合模式

### 3.1 先 SQL 后向量

1. 用 SQL 过滤候选（如 `SELECT id FROM articles ... LIMIT 200`）。
2. 将 ID 传入 `VectorSearchByIDs` 进行二次排序。

### 3.2 先向量后 SQL 补全

1. `VectorSearch` 获取 top-N ID。
2. 再用 `xb.Of("articles").In("id", ids...)` 查询详情并联表。

### 3.3 多租户

```go
xb.Of(&FeedVector{}).
    Meta(func(meta *interceptor.Metadata) {
        meta.Set("tenant_id", tenantID)
    }).
    Eq("tenant_id", tenantID).
    VectorSearch("embedding", vec, 40)
```

- Metadata 会经过拦截器，可用于日志与审计。

---

## 4. 多样性辅助

`QdrantCustom` 可配置 `WithHashDiversity`、`WithMinDistance`、`WithPayloadSelector` 等 helper。

```go
custom := xb.NewQdrantBuilder().
    WithHashDiversity(func(h *xb.HashDiversity) {
        h.Field = "category"
        h.Modulo = 4
    })

built := xb.Of(&ProductVector{}).
    Custom(custom).
    VectorSearch("embedding", vec, 12).
    Build()
```

---

## 5. 排障清单

| 症状 | 可能原因 | 解决办法 |
|------|----------|----------|
| `filter.must` 为空 | 条件值被自动跳过 | 检查是否为零值/空字符串 |
| `Custom is nil` | 未调用 `Custom()` | 添加 `xb.NewQdrantBuilder().Build()` |
| limit 不生效 | 只在 SQL Builder 里设置 | 在 `VectorSearch` 或 Recommend Builder 里设置 |
| 租户串线 | 忘记 `Eq("tenant_id", ...)` | 所有 builder 加租户守卫 |

---

## 6. 相关文档

- `QDRANT_GUIDE.md`：API 细节
- `CUSTOM_INTERFACE.md`：自定义向量数据库适配
- `FILTERING.md`：理解自动跳过规则

若你有值得分享的模式，欢迎提交 PR 补充本指南。

