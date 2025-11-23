# Custom Vector DB Guide (English)

Based on `xb/doc/CUSTOM_VECTOR_DB_GUIDE.md`. It captures best practices for building vector-database adapters with the `Custom` interface.

---

## 1. Adapter blueprint

| Concern | Recommendation |
|---------|----------------|
| Payload schema | Define Go structs that mirror the DB JSON to avoid map juggling |
| Filter encoding | Reuse `built.FilterJSON()` where possible |
| Vector serialization | Accept both `[]float32` and pre-normalized vectors |
| Pagination | Support `limit`, `offset`, `scroll_id` according to backend |

---

## 2. Preset constructors

- `NewQdrantBuilder()` – official reference
- `NewMilvusBuilder()` – include server endpoint + GRPC options (future implementation)
- `NewWeaviateBuilder()` – handles class names + where filters (future implementation)

Expose variations like `HighPrecision`, `HighRecall`, or `HighSpeed` tuned for specific workloads.

---

## 3. Mapping builder features

| Builder feature | Vector DB mapping |
|-----------------|------------------|
| `VectorSearch` | `search`/`recommend` API |
| `Eq`/`In` | filter `must` clauses |
| `Meta` | request-level metadata (trace, tenant) |
| `Limit`, `Offset` | `limit`, `offset`, `scroll` |

Document unsupported combinations so callers know when to fall back to raw APIs.

---

## 4. Operational tips

- Validate collection/namespace existence before executing.
- Surface backend errors verbatim; it helps with misconfigured payloads.
- Add optional logging hooks for debugging payloads without touching application code.

---

## 5. Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/VECTOR_GUIDE.md`
- `doc/en/AI_APPLICATION.md`

