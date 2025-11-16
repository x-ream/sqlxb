# AI Agent 工具包

本文档是 `xb/doc/ai_application/AGENT_TOOLKIT.md` 的中文版本。它涵盖如何将 xb 构建器暴露为 GPT 风格 agent 的函数调用工具。

---

## JSON 模式技巧

- 定义清晰的输入字段（`tenant_id`、`vector`、`limit`）。
- 在调用 xb 之前验证枚举和数值范围。
- 返回原始 `JsonOfSelect` 输出，以便 agent 可以解释分数/负载。

---

## 安全性

- 添加 `Meta` 信息（agent ID、会话 ID）。
- 使用 `InRequired` 或租户守卫以避免广泛扫描。
- 记录每个工具调用以进行审计。

---

## 相关文档

- `doc/cn/AI_APPLICATION.md`
- `doc/cn/QDRANT_GUIDE.md`

