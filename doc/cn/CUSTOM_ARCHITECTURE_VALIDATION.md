# Custom 架构验证

本文档是 `xb/doc/CUSTOM_ARCHITECTURE_VALIDATION.md` 的中文版本。它解释了添加新 `Custom` 适配器（向量数据库、SQL 方言、内部 API）的护栏。

---

## 检查清单

1. **接口契约** – `Generate(*Built)` 必须是纯函数、确定性的且无副作用。
2. **错误处理** – 返回可操作的错误（缺少命名空间、不支持的子句等）。
3. **测试覆盖** – `JsonOfSelect`、`JsonOfInsert`、`SqlOfSelect`（如适用）的单元测试。
4. **文档** – 一旦稳定，在 README + 发布说明中提及适配器。

---

## 架构关注点

- **确定性 JSON** – 保持 map 排序可预测或使用结构体进行编组。
- **功能门控** – 在构建器方法（`WithPayloadSelector`、`WithMinDistance` 等）后面保护数据库特定标志。
- **可观测性** – 将 `Built.Meta`（TraceID、租户）传播到你的负载中以进行日志记录。
- **可扩展性** – 暴露预设构造函数（`Default`、`HighPrecision`、`HighSpeed`）而不是原始结构体字段。

---

## 迁移技巧

- 删除旧适配器时，保留兼容包或在 `MIGRATION.md` 中宣布。
- 使用语义版本控制：次要版本用于添加功能，主要版本用于破坏 `Custom` 行为。
- 为基于你的适配器构建的下游团队记录升级路径。

---

## 相关文档

- `doc/cn/CUSTOM_INTERFACE.md`
- `doc/cn/CUSTOM_INTERFACE_PHILOSOPHY.md`
- `doc/cn/CUSTOM_VECTOR_DB_GUIDE.md`

