# Semantic Kernel 集成

本文档是 `xb/doc/ai_application/SEMANTIC_KERNEL_INTEGRATION.md` 的中文版本。

---

## 模式

- 实现包装 xb 构建器的 `IQueryFunction` 或 `ISKFunction`。
- 通过 `Meta` 传递租户/上下文信息。
- 返回水合的 JSON，以便下游技能可以推理分数和负载。

---

## 测试

- 模拟 Semantic Kernel 上下文并断言 xb 构建器接收预期的输入。
- 端到端验证多步骤技能（检索 → 总结）。

---

## 参考资料

- `doc/cn/AI_APPLICATION.md`
- `doc/cn/ai_application/AGENT_TOOLKIT.md`

