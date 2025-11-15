# Dialect vs Custom Design (English)

English version of `xb/doc/DIALECT_CUSTOM_DESIGN.md`. It compares traditional ORM dialect switches with xb’s `Custom` architecture.

---

## Dialect model (legacy ORMs)

- Central enum listing every database (`MySQL`, `PostgreSQL`, `Oracle`, ...).
- Framework implements vendor-specific SQL generation internally.
- Adding a new database requires touching the core repository.
- Release cadence slows as more dialects pile up.

---

## Custom model (xb)

- Single `Custom` interface; adapters live in user space.
- Core stays small and stable.
- Teams ship adapters without waiting for upstream review.
- Works for SQL, JSON, GRPC, HTTP, CLI, or any other execution layer.

---

## When to choose each

| Scenario | Recommended model |
|----------|------------------|
| Commodity SQL with standard syntax | Built-in SQL generator |
| Specialized engines (ClickHouse, Qdrant) | Custom |
| Proprietary internal APIs | Custom |
| Need instant experimentation | Custom |

---

## Guidance for adapter authors

1. Provide typed builders/configs.
2. Document limitations (unsupported clauses, max vectors, etc.).
3. Keep tests close to the adapter.
4. Follow semantic versioning when exposing APIs to other teams.

See `doc/en/CUSTOM_INTERFACE_PHILOSOPHY.md` for the underlying design motivations.

