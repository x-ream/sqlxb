# Naming Convention (English)

English version of `xb/doc/NAMING_CONVENTION.md`. It standardizes how builders, structs, tables, and files are named across the project.

---

## Packages and files

- Use lowercase with underscores for files in Go (`builder_select.go`, `qdrant_custom.go`).
- Keep package names short (`xb`, `interceptor`, `qdrant`).
- Mirror test files (`builder_select_test.go`) near their implementation.

---

## Structs and DTOs

- Exported structs: `CamelCase` (`Builder`, `Built`, `QdrantCustom`).
- Request/response DTOs include suffixes (`ListUserRequest`, `SearchResult`).
- Use `RO` or `VO` suffixes for read-only view objects if needed.

---

## Builder methods

- Verb-style for operations (`Select`, `Sort`, `Limit`, `VectorSearch`).
- Short nouns for DSL helpers (`Or`, `And`, `Meta`, `Cond`).
- Keep chain order consistent to aid readability.

---

## Database artifacts

- Tables: `snake_case` (`t_user`, `recent_orders`).
- Columns: `snake_case`.
- CTE aliases: short but descriptive (`recent_orders ro`, `team_hierarchy th`).

---

## Related docs

- `doc/en/BUILDER_BEST_PRACTICES.md`
- `doc/en/TESTING_STRATEGY.md`

