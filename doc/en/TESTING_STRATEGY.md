# Testing Strategy (English)

English translation of `xb/doc/TESTING_STRATEGY.md`. It outlines how the project prevents regressions.

---

## Prior incidents

- **Vector vs SQL split** – vector builders lacked `And`/`Or` tests, leading to regressions.
- **Numeric zero handling** – floats were not covered, so `Gt("score", 0.0)` behaved incorrectly.

---

## Current plan

1. Each feature ships with unit tests for both SQL and vector paths.
2. Snapshot tests guard `JsonOfSelect()` payloads.
3. Regression suites run via `go test ./...` and capture coverage numbers.
4. Release checklists require manual verification of demo projects.

---

## Tips

- Test builder AST via `built.Raw()` alongside rendered SQL/JSON.
- Cover boundary values (`0`, `nil`, empty slices).
- Add fixtures for every advanced Qdrant API (Recommend/Discover/Scroll).

---

## Related docs

- `doc/en/ALL_FILTERING_MECHANISMS.md`
- `doc/en/QDRANT_ADVANCED_API.md`

