# To JSON Design Clarification (English)

English summary of `xb/doc/TO_JSON_DESIGN_CLARIFICATION.md`. It clarifies how `JsonOfSelect/Insert/Update/Delete` behave and why they converge through `Built`.

---

## Design goals

1. **One entry point** – `Built.JsonOfSelect()` replaces multiple `ToQdrant*JSON()` helpers.
2. **Adapter ownership** – concrete JSON schemas live inside `Custom` implementations.
3. **Future-proof** – new backends inherit the same lifecycle without touching builder code.

---

## Execution flow

1. Build DSL → `Built`
2. `Built.JsonOfSelect()` checks if a `Custom` is attached.
3. If yes, delegate to `Custom.Generate`.
4. If no, fall back to SQL generation (or return an error for vector-only features).

---

## Error handling

- Missing `Custom` when vector clauses exist → descriptive error.
- Unsupported operations (e.g., vector insert without adapter) should panic in tests so teams can catch mismatches early.
- Always bubble adapter errors up; do not wrap unless you add actionable metadata.

---

## Testing tips

- Snapshot JSON outputs for each adapter.
- Assert that SQL builders without `Custom` still work when calling `JsonOfSelect()` (should error predictably).
- Cover multiple API types (Search, Recommend, Discover, Scroll) in the same suite.

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/CUSTOM_INTERFACE.md`

