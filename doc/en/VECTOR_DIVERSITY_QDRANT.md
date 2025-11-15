# Vector Diversity & Qdrant (English)

English remake of `xb/doc/VECTOR_DIVERSITY_QDRANT.md`. It explains how xb balances recommendation diversity using Qdrant helpers.

---

## Diversity strategies

| Strategy | Builder helper | Description |
|----------|----------------|-------------|
| Hash diversity | `WithHashDiversity(func(*HashDiversity))` | Spreads results across payload buckets |
| Score floor | `WithMinDistance(float32)` | Drops items below a similarity threshold |
| Payload projection | `WithPayloadSelector` | Controls fields returned for downstream rerankers |

---

## Example

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

## Best practices

- Choose hash fields that correlate with business diversity (brand, category, tenant).
- Tune modulo to match the number of slots you want.
- Combine with server-side rerankers if you need deterministic diversity.

---

## Related docs

- `doc/en/VECTOR_GUIDE.md`
- `doc/en/QDRANT_OPTIMIZATION_SUMMARY.md`

