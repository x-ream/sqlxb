# Common Errors (English)

Quick reference for mistakes originally documented in `xb/doc/COMMON_ERRORS.md`. Use this checklist before opening an issue.

---

## 1. Builder usage

- **Missing `Build()`** – chaining returns a builder, not SQL. Always end with `.Build()`.
- **Calling `SqlOfSelect()` too early** – only `Built` has SQL generation methods.
- **Mixing pointer/value builders** – in Go, keep a single pattern (`xb.Of(&User{})`) to avoid unexpected copies.
- **Nil pointer DTOs** – check input before passing to `Of()`.

---

## 2. Auto-filter surprises

- `Eq("status", 0)` disappears because zero is auto-filtered.
- `In("id")` with empty slices collapses; switch to `InRequired`.
- `Like("name", "")` is ignored; sanitize input first.

---

## 3. Custom adapters

- `JsonOfSelect failed: Custom is nil` → attach `xb.NewQdrantCustom()` or your own adapter.
- `Custom.Generate` returning unsupported types → stick to `string`, `[]byte`, or driver-ready structs.
- Ensure your `Custom` respects builder state (`Built.Table`, `Built.Conds`, etc.).

---

## 4. SQL execution

- Use the argument slice returned by `SqlOfSelect` directly; do not re-order manually.
- For pagination, always call both `Limit` and `Offset` to avoid driver defaults; And SqlOfPage() will return: countSQL, dataSQL, vs, metaMap.
- When joining custom snippets via `X()`, sanitize them to prevent SQL injection.

---

## 5. Debug workflow

1. Print `built.SqlOfSelect()` and `built.Args` (or `built.ArgsAsStrings()` in other languages).
2. Inspect `built.Raw()` to ensure the AST matches expectations.
3. Add regression tests around the failing scenario.

More edge cases live in `doc/en/TESTING_STRATEGY.md` and `doc/en/FILTERING.md`.

