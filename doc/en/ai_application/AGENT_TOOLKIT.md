# AI Agent Toolkit (English)

English rewrite of `xb/doc/ai_application/AGENT_TOOLKIT.md`. It covers how to expose xb builders as function-calling tools for GPT-style agents.

---

## JSON schema tips

- Define clear input fields (`tenant_id`, `vector`, `limit`).
- Validate enums and numeric ranges before calling xb.
- Return raw `JsonOfSelect` output so agents can interpret scores/payloads.

---

## Safety

- Add `Meta` info (agent ID, session ID).
- Use `InRequired` or tenant guards to avoid broad scans.
- Log every tool invocation for auditing.

---

## Related docs

- `doc/en/AI_APPLICATION.md`
- `doc/en/QDRANT_GUIDE.md`

