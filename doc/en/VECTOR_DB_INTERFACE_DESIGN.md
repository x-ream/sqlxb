# Vector DB Interface Design (English)

This file mirrors `xb/doc/VECTOR_DB_INTERFACE_DESIGN.md`. It details the public interfaces exposed to vector adapters.

---

## Key structs

- `Builder` – fluent DSL entry point.
- `Built` – immutable snapshot consumed by SQL/JSON renderers.
- `Custom` – interface adapters implement.
- `QdrantCustom`, `MilvusCustom` (examples) – wrappers around the interface.

---

## Design principles

| Principle | Description |
|-----------|-------------|
| Minimal API | Keep public surface tiny; rely on builder closures |
| Type safety | Use sum types / enums where possible (in Zig/V) |
| Deterministic output | Same input tree → same SQL/JSON |
| Extensibility | Hooks for metadata, validation, interceptors |

---

## Required methods

- `VectorSearch(field string, vector xb.Vector, limit int)`
- `VectorSearchMulti(...)`
- `VectorSearchByIDs(...)`
- `Meta(func(*interceptor.Metadata))`

Adapters should respect these fields when generating requests.

---

## Related docs

- `doc/en/VECTOR_DB_ABSTRACTION_SUMMARY.md`
- `doc/en/CUSTOM_INTERFACE.md`

