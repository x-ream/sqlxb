# Vector DB Extension Guide (English)

English version of `xb/doc/VECTOR_DB_EXTENSION_GUIDE.md`. It walks through extending xb to new vector engines.

---

## Extension workflow

1. Study the target DB’s JSON/GRPC API.
2. Implement a `Custom` adapter that converts `Built` into the DB payload.
3. Write snapshot tests for `JsonOfSelect/Insert/Update/Delete`.
4. Document presets and limitations.

---

## Considerations

- **Authentication** – let callers inject tokens/headers.
- **Batching** – expose helpers for bulk upserts if the backend benefits from it.
- **Failover** – plan for retries or multi-endpoint deployments.
- **Versioning** – keep adapter semantics stable; use `WithCompatMode` if needed.

---

## Template

- Copy `xb/doc/MILVUS_TEMPLATE.go` as a starting point.
- Replace payload structs with your DB schema.
- Wire the adapter into examples/tests so others can learn from it.

---

## Related docs

- `doc/en/CUSTOM_VECTOR_DB_GUIDE.md`
- `doc/en/VECTOR_DB_INTERFACE_DESIGN.md`

