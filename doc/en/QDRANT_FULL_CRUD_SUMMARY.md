# Qdrant Full CRUD Summary (English)

English version of `xb/doc/QDRANT_FULL_CRUD_SUMMARY.md`. It outlines how xb mirrors SQL semantics across Qdrant’s CRUD APIs.

---

## Operations

| Operation | Builder entry | Notes |
|-----------|---------------|-------|
| Insert | `Insert(func(*InsertBuilder))` + `JsonOfInsert()` | Supports batch inserts |
| Update | `Update(func(*UpdateBuilder))` + `JsonOfUpdate()` | Requires filter/ids |
| Delete | `Build()` + `JsonOfDelete()` | Respects same filter DSL as SQL |
| Select | `Build()` + `JsonOfSelect()` | Unified Search/Recommend/Discover/Scroll |

---

## Common patterns

- Use `Eq("id", id)` for single-record operations.
- `InRequired` prevents unintended mass updates/deletes.
- Metadata (`Meta(func)`) is propagated so audit logs know who ran the operation.

---

## Feature parity matrix

| SQL concept | Qdrant mapping |
|-------------|----------------|
| WHERE clauses | `filter.must` / `filter.should` |
| ORDER BY score | inherent to vector search; explicit sorts optional |
| LIMIT/OFFSET | `limit`, `offset`, `scroll_id` |
| With/CTE | Not applicable; keep logic in app layer |

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/CUSTOM_INTERFACE.md`

