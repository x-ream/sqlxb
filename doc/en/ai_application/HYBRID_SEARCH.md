# Hybrid Search (English)

English summary of `xb/doc/ai_application/HYBRID_SEARCH.md`. It covers combining scalar filters with vector ranking.

---

## Patterns

- **SQL first** – filter candidates with SQL, then vector search by IDs.
- **Vector first** – retrieve top-N vectors, then join with SQL for enrichment.
- **Mixed** – run both, merge results with business scoring.

---

## Implementation tips

- Use `Cond` blocks to wrap optional filters.
- Keep vector limits small when chaining to SQL (avoid large `IN` lists).
- Monitor total latency; break down SQL vs vector cost.

---

## References

- `doc/en/VECTOR_GUIDE.md`
- `doc/en/QDRANT_GUIDE.md`

