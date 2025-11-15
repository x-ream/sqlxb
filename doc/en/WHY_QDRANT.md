# Why Qdrant? (English)

English adaptation of `xb/doc/WHY_QDRANT.md`. It explains why Qdrant became the first officially supported vector backend.

---

## Reasons to choose Qdrant

1. **Feature velocity** – frequent releases with Recommend, Discover, Scroll, payload selectors.
2. **Operational maturity** – horizontal scaling, snapshots, cloud + self-hosted options.
3. **JSON-first design** – easy to integrate with xb’s `Custom.Generate`.
4. **Open-source friendly** – permissive license and strong community.

---

## Comparison snapshot

| Engine | Notes |
|--------|-------|
| Qdrant | Balanced features, predictable API |
| Milvus | Great at scale but more complex to operate |
| Pinecone | Managed only; closed APIs |
| Weaviate | Also JSON-first but heavier schema management |

---

## Roadmap alignment

- Qdrant’s advanced APIs map cleanly to xb’s DSL.
- Diversity helpers mirror Qdrant’s payload selectors.
- Future features (rerank, multi-vector) can be added via `Custom`.

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/QDRANT_ADVANCED_API.md`
- `doc/en/VECTOR_GUIDE.md`

