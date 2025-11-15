# NL2SQL Experiments (English)

English rewrite of `xb/doc/ai_application/NL2SQL.md`. It tracks experiments converting natural language to xb builders.

---

## Approach

- Use LLMs to produce structured intents (filters, sorts, limits).
- Validate intents before applying them to a builder.
- Provide fallback templates for unsupported queries.

---

## Safety

- Whitelist allowed columns to avoid arbitrary SQL injection.
- Require explicit tenant/user filters.
- Log both the natural query and rendered SQL for audits.

---

## References

- `doc/en/AI_APPLICATION.md`

