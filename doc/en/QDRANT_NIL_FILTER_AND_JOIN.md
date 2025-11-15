# Qdrant Nil Filter & Join Notes (English)

Summarizes `xb/doc/QDRANT_NIL_FILTER_AND_JOIN.md`. It documents how xb treats nil/empty filters and how JOIN-style constructs map to Qdrant filters.

---

## Nil filter behavior

- Empty strings, zero values, and `nil` pointers are removed before generating Qdrant filters.
- `InRequired` enforces that at least one ID/field value remains.
- `Meta` data is still emitted even when no filter clauses survive, enabling multi-tenant logging.

---

## JOIN analogy

While Qdrant does not support SQL joins, xb simulates related behavior via filter composition:

- Use nested `Cond` blocks to represent `AND`/`OR` groups.
- For parent-child documents, store parent IDs in payload fields and filter by them.
- Diversity helpers act like implicit joins by ensuring different payload buckets appear in the result.

---

## Recommendations

- Normalize payload keys up front to avoid mismatched filters.
- Keep boolean fields explicit (`true`/`false`) since empty values are skipped.
- Write regression tests covering nil filter + vector combos to catch edge cases.

---

## Related docs

- `doc/en/ALL_FILTERING_MECHANISMS.md`
- `doc/en/QDRANT_GUIDE.md`

