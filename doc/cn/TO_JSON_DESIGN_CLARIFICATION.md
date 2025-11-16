# To JSON 设计说明

本文档是 `xb/doc/TO_JSON_DESIGN_CLARIFICATION.md` 的中文版本。它阐明了 `JsonOfSelect/Insert/Update/Delete` 的行为以及它们如何通过 `Built` 收敛。

---

## 设计目标

1. **一个入口点** – `Built.JsonOfSelect()` 替换多个 `ToQdrant*JSON()` 辅助方法。
2. **适配器所有权** – 具体 JSON 模式存在于 `Custom` 实现内部。
3. **面向未来** – 新后端继承相同的生命周期，而无需接触构建器代码。

---

## 执行流程

1. 构建 DSL → `Built`
2. `Built.JsonOfSelect()` 检查是否附加了 `Custom`。
3. 如果是，委托给 `Custom.Generate`。
4. 如果不是，回退到 SQL 生成（或为仅向量功能返回错误）。

---

## 错误处理

- 当存在向量子句时缺少 `Custom` → 描述性错误。
- 不支持的操作（例如，没有适配器的向量插入）应在测试中 panic，以便团队可以及早捕获不匹配。
- 始终向上冒泡适配器错误；除非添加可操作的元数据，否则不要包装。

---

## 测试技巧

- 为每个适配器快照 JSON 输出。
- 断言没有 `Custom` 的 SQL 构建器在调用 `JsonOfSelect()` 时仍然有效（应该可预测地错误）。
- 在同一套件中涵盖多种 API 类型（Search、Recommend、Discover、Scroll）。

---

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/CUSTOM_INTERFACE.md`

