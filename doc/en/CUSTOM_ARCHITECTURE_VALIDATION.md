# Custom Architecture Validation (English)

English counterpart of `xb/doc/CUSTOM_ARCHITECTURE_VALIDATION.md`. It explains the guardrails for adding new `Custom` adapters (vector DB, SQL dialect, internal API).

---

## Checklist

1. **Interface contract** – `Generate(*Built)` must be pure, deterministic, and side-effect free.
2. **Error story** – return actionable errors (missing namespace, unsupported clause, etc.).
3. **Test surface** – unit tests for `JsonOfSelect`, `JsonOfInsert`, `SqlOfSelect` (if applicable).
4. **Docs** – mention the adapter in README + release notes once stable.

---

## Architecture concerns

- **Deterministic JSON** – keep map ordering predictable or marshal using structs.
- **Feature gating** – guard database-specific flags behind builder methods (`WithPayloadSelector`, `WithMinDistance`, etc.).
- **Observability** – propagate `Built.Meta` (TraceID, tenant) into your payload for logging.
- **Extensibility** – expose preset constructors (`Default`, `HighPrecision`, `HighSpeed`) rather than raw struct fields.

---

## Migration tips

- When removing legacy adapters, leave compatibility packages or announce in `MIGRATION.md`.
- Use semantic versioning: minor versions for additive features, major when breaking `Custom` behavior.
- Document upgrade paths for downstream teams building on top of your adapter.

---

## Related docs

- `doc/en/CUSTOM_INTERFACE.md`
- `doc/en/CUSTOM_INTERFACE_PHILOSOPHY.md`
- `doc/en/CUSTOM_VECTOR_DB_GUIDE.md`


