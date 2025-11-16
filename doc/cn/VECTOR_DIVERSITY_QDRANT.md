# 向量多样性与 Qdrant

本文档是 `xb/doc/VECTOR_DIVERSITY_QDRANT.md` 的中文版本。它解释了 xb 如何使用 Qdrant 辅助方法平衡推荐多样性。

---

## 多样性策略

| 策略 | 构建器辅助方法 | 描述 |
|------|--------------|------|
| Hash 多样性 | `WithHashDiversity(func(*HashDiversity))` | 将结果分散到 payload 桶中 |
| 分数下限 | `WithMinDistance(float32)` | 丢弃低于相似度阈值的项目 |
| Payload 投影 | `WithPayloadSelector` | 控制为下游重排序器返回的字段 |

---

## 示例

```go
custom := xb.NewQdrantCustom().
    WithHashDiversity(func(h *xb.HashDiversity) {
        h.Field = "category"
        h.Modulo = 4
    }).
    WithMinDistance(0.35)

json, _ := xb.Of(&ProductVector{}).
    Custom(custom).
    VectorSearch("embedding", vec, 12).
    Build().
    JsonOfSelect()
```

---

## 最佳实践

- 选择与业务多样性相关的 hash 字段（品牌、类别、租户）。
- 调整 modulo 以匹配所需的槽数。
- 如果需要确定性多样性，与服务器端重排序器结合使用。

---

## 相关文档

- `doc/cn/VECTOR_GUIDE.md`
- `doc/cn/QDRANT_OPTIMIZATION_SUMMARY.md`

