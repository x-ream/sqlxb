# Custom JOINs Guide (English)

English version of `xb/doc/CUSTOM_JOINS_GUIDE.md`. It explains how to extend xb’s JOIN DSL for dialect-specific syntax such as ClickHouse `GLOBAL JOIN` or `ASOF JOIN`.

---

## Built-in JOIN catalog

```go
const (
    inner_join      = "INNER JOIN"
    left_join       = "LEFT JOIN"
    right_join      = "RIGHT JOIN"
    cross_join      = "CROSS JOIN"
    asof_join       = "ASOF JOIN"
    global_join     = "GLOBAL JOIN"
    full_outer_join = "FULL OUTER JOIN"
)
```

Use `FromX(func(*FromBuilder))` with `JOIN(kind)` helpers to compose multi-step pipelines.

---

## Adding your own JOIN

```go
func LATERAL() xb.JOIN {
    return func() string {
        return "LATERAL JOIN"
    }
}
```

- Return a closure so the builder can call it lazily.
- Register short helper methods (`func (fb *FromBuilder) Lateral(...)`) if you need extra configuration.

---

## Best practices

- Keep JOIN helpers pure; they should only format SQL keywords.
- Reuse `Cond(func(*ON))` to define predicates; this keeps auto-filtering intact.
- For dialect-specific hints (`GLOBAL`, `FINAL`, etc.) expose them via `FromX` configuration so the rest of the chain stays portable.

---

## Examples

- Multi-join pipeline mixing subqueries and ON clauses.
- Custom join that injects optimizer hints.
- Cross-shard join that enforces tenant filters inside every ON block.

Refer to `doc/en/BUILDER_BEST_PRACTICES.md` for more patterns.

