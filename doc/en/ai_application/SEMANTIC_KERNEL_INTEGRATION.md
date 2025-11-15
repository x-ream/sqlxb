# Semantic Kernel Integration (English)

English notes for `xb/doc/ai_application/SEMANTIC_KERNEL_INTEGRATION.md`.

---

## Pattern

- Implement `IQueryFunction` or `ISKFunction` that wraps an xb builder.
- Pass tenant/context info via `Meta`.
- Return hydrated JSON so downstream skills can reason about scores and payloads.

---

## Testing

- Mock Semantic Kernel context and assert the xb builder receives the expected inputs.
- Validate multi-step skills (retrieve → summarize) end-to-end.

---

## References

- `doc/en/AI_APPLICATION.md`
- `doc/en/ai_application/AGENT_TOOLKIT.md`

