# Qdrant Optimization Summary (English)

English adaptation of `xb/doc/QDRANT_OPTIMIZATION_SUMMARY.md`. It captures tuning tips for latency, recall, and diversity.

---

## Performance levers

- **HNSW params** – expose `HnswEf`, `ScoreThreshold`, `Exact` flags via `QdrantCustom`.
- **Payload projection** – include only necessary fields to reduce payload size.
- **Vector normalization** – pre-normalize to keep scoring predictable.
- **Batching** – send bulk inserts/updates instead of single record calls.

---

## Recall vs latency

| Goal | Recommended settings |
|------|----------------------|
| Low latency | Lower `limit`, moderate `HnswEf`, enable payload trimming |
| High recall | Increase `HnswEf`, `limit`, disable payload trimming |
| Balanced | Default `NewQdrantBuilder().Build()` plus `WithHashDiversity` |

---

## Observability

- Record `TraceID`, `TenantID`, and adapter config in logs.
- Track P95 latency per API (Search, Recommend, Discover, Scroll).
- Monitor `score` distributions to catch vector drift.

---

## Related docs

- `doc/en/VECTOR_GUIDE.md`
- `doc/en/QDRANT_ADVANCED_API.md`
- `doc/en/AI_APPLICATION.md`

