# Vector Guide (English)

This guide consolidates lessons learned while building hybrid SQL + vector pipelines with xb. Use it alongside the Qdrant guide for implementation details.

---

## 1. Embedding hygiene

- Normalize your vectors (L2) before calling `VectorSearch`. xb does not modify the payload, so mismatched normalization leads to noisy scores.
- Store vector dimension alongside the blob. `VectorSearch("embedding", vec, limit)` will not validate lengths—fail fast in your data layer.
- For multi-vector documents (e.g., text + code), use `VectorSearchMulti` or create multiple builders and `UNION ALL` the results with score weighting.

---

## 2. Typical select flow

```go
built := xb.Of(&CodeVector{}).
    Eq("language", "go").
    VectorSearch("embedding", queryVector, 20).
    Sort("score", xb.Desc).
    Offset(0).
    Build()

json, err := built.JsonOfSelect()
```

- Conditions (`Eq`, `In`, `Meta`) work the same as SQL; they become `filter.must` fragments inside the generated JSON.
- Sorting on `score` is optional—Qdrant already sorts by score, but adding it keeps SQL + vector snippets consistent.

---

## 3. Hybrid query patterns

### 3.1 SQL first, vector second

1. Build an SQL query that narrows candidates (e.g., `SELECT id FROM articles WHERE status = 1 LIMIT 200`).
2. Feed those IDs into `VectorSearchByIDs` to re-rank in Qdrant.

### 3.2 Vector first, SQL enrich

1. Use `VectorSearch` to retrieve top-N IDs.
2. Use `xb.Of("articles").In("id", ids...)` to fetch details and join with other tables.
3. Combine both steps inside a transaction to keep latency measurable.

### 3.3 Multi-tenant

```go
xb.Of(&FeedVector{}).
    Meta(func(meta *interceptor.Metadata) {
        meta.Set("tenant_id", tenantID)
    }).
    Eq("tenant_id", tenantID).
    VectorSearch("embedding", vec, 40)
```

- Metadata flows through interceptors so you can log tenant / trace IDs even for vector-only calls.

---

## 4. Diversity helpers

`QdrantCustom` exposes helpers like `WithHashDiversity`, `WithMinDistance`, and `WithPayloadSelector`. They map directly to Qdrant’s `with_payload_selector`, `diversity` and other knobs.

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

## 5. Troubleshooting checklist

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Empty `filter.must` | Condition values skipped | Check for zero/empty strings—xb auto-skips them |
| `JsonOfSelect` error: `Custom is nil` | Vector payload without `Custom()` | Attach `xb.NewQdrantBuilder().Build()` before `Build()` |
| Wrong limit | Limit set only on SQL builder | Use `VectorSearch(..., limit)` or `RecommendBuilder.Limit()` |
| Mixed tenants | Missing `Eq("tenant_id", ...)` | Add tenant guard to every builder |

---

## 6. Related docs

- `QDRANT_GUIDE.md` for API-specific parameters
- `CUSTOM_INTERFACE.md` if you plan to target Milvus/Weaviate/etc.
- `FILTERING.md` to understand why some conditions disappear automatically

Have a pattern worth sharing? Open a PR and append it to this guide! 

