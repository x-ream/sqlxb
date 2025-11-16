# Qdrant Custom 设计分析

本文档是 `xb/doc/QDRANT_CUSTOM_DESIGN_ANALYSIS.md` 的中文版本。它剖析了 Qdrant 适配器的结构以及做出某些权衡的原因。

---

## 架构层

1. **Builder** – 收集 DSL 状态（过滤器、向量、排序、payload 选择器）。
2. **Custom 配置** – 暴露有用的旋钮（`Recommend`、`Discover`、`Scroll`、多样性、payload 选择器）。
3. **Generator** – 将构建器状态映射到 JSON 负载。
4. **Transport** – 留给调用者（HTTP 客户端、SDK 等）。

---

## 关键决策

- **单一入口点** – `JsonOfSelect()` 检查配置以选择 Search/Recommend/Discover/Scroll。
- **可组合构建器** – 高级 API 为过滤器重用相同的 `CondBuilder` DSL。
- **预设函数** – `NewQdrantCustom()` 暴露默认设置；可以在此基础上叠加高级预设（高召回率、高多样性）。

---

## 可扩展性钩子

- `WithPayloadSelector`
- `WithHashDiversity`
- `WithMinDistance`
- `WithNamespace`、`WithTenant`

每个辅助方法更新适配器状态，并可与高级 API 组合。

---

## 测试策略

- 每个分支（search + 3 个高级 API）的 JSON 快照。
- 回归测试确保元数据传播和自动过滤行为一致。
- 每当 Qdrant 发布新的可选字段时添加针对性测试。

---

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/QDRANT_ADVANCED_API.md`
- `doc/cn/TO_JSON_DESIGN_CLARIFICATION.md`

