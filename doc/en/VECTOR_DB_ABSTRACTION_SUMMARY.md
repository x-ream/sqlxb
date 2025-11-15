# Vector DB Abstraction Summary (English)

English digest of `xb/doc/VECTOR_DB_ABSTRACTION_SUMMARY.md`. It explains how xb abstracts vector databases without sacrificing power.

---

## Goals

- Keep the fluent builder identical for SQL and vector workloads.
- Expose vector-specific controls via `Custom` instead of polluting the base API.
- Enable users to run hybrid pipelines (SQL → vector → SQL).

---

## Abstraction layers

| Layer | Responsibility |
|-------|----------------|
| Builder | Collects filters, sorts, vector ops |
| Built | Immutable snapshot for SQL/JSON generation |
| Custom adapter | Renders JSON payloads |
| Transport | Caller’s HTTP/GRPC client |

---

## What stays in adapters

- Payload schema
- Authentication/transport
- Advanced knobs (diversity, multi-vector, payload projection)
- Backend-specific limits

---

## Related docs

- `doc/en/CUSTOM_INTERFACE.md`
- `doc/en/QDRANT_GUIDE.md`

