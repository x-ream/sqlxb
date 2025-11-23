# Custom Interface Philosophy (English)

Summarizes `xb/doc/CUSTOM_INTERFACE_PHILOSOPHY.md`. It describes why xb chose the `Custom` interface instead of hard-coding dozens of dialects.

---

## Principles

1. **Minimal surface** – one method (`Generate`) handles every operation (SQL or JSON).
2. **User ownership** – teams can implement niche databases without waiting for core releases.
3. **Predictable behavior** – same builder API regardless of backend.
4. **Composability** – Combine SQL + vector workflows with a single fluent chain.

---

## Dialect vs Custom

| Dialect enum | Custom interface |
|--------------|-----------------|
| Framework bundles every database | Framework stays tiny; users extend it |
| Requires core PRs for new DBs | Users ship adapters in their own repos |
| Hard to keep in sync | Adapter owners iterate at their own pace |

---

## Design rules

- Keep adapters stateless or easily clonable.
- Return types must be driver-friendly (`string`, `[]byte`, or thin structs).
- Provide builders (`NewQdrantBuilder()`, `NewMilvusBuilder()`) instead of exposing raw structs.
- Document advanced knobs so users can reason about trade-offs.

---

## Recommended reading

- `doc/en/CUSTOM_INTERFACE.md`
- `doc/en/CUSTOM_QUICKSTART.md`
- `doc/en/CUSTOM_VECTOR_DB_GUIDE.md`

