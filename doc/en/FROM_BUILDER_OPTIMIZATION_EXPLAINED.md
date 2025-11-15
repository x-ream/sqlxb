# From Builder Optimization Explained (English)

English companion to `xb/doc/FROM_BUILDER_OPTIMIZATION_EXPLAINED.md`. It dives into how `FromBuilder` rewrites complex JOIN/CTE graphs while keeping SQL deterministic.

---

## Goals

1. **Predictable aliasing** – every subquery/CTE gets a deterministic alias.
2. **Streaming JOIN planning** – `FromX` receives a function so it can inline or hoist subqueries.
3. **Tenant safety** – automatically injects guards into nested sources when helpers call `WITH` blocks.

---

## Optimization passes

- **Flattening** – merge consecutive `FromX` blocks when they only add projections.
- **Pruning** – remove unused CTEs referenced by no downstream clause.
- **Hint propagation** – carry optimizer hints (e.g., `FINAL`, `SAMPLE`) down to sub-joins.

---

## Debugging

- Use `builder.DebugFrom()` (internal helper) to log the staged AST.
- Unit tests can snapshot the intermediate representation before SQL rendering.
- When adding new join kinds, ensure both the AST and SQL layers support them.

---

## See also

- `doc/en/CUSTOM_JOINS_GUIDE.md`
- `doc/en/BUILDER_BEST_PRACTICES.md`

