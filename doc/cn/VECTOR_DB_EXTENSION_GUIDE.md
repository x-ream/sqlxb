# 向量数据库扩展指南

本文档是 `xb/doc/VECTOR_DB_EXTENSION_GUIDE.md` 的中文版本。它引导你扩展 xb 到新的向量引擎。

---

## 扩展工作流

1. 研究目标数据库的 JSON/GRPC API。
2. 实现将 `Built` 转换为数据库负载的 `Custom` 适配器。
3. 为 `JsonOfSelect/Insert/Update/Delete` 编写快照测试。
4. 记录预设和限制。

---

## 注意事项

- **身份验证** – 让调用者注入令牌/标头。
- **批处理** – 如果后端从中受益，暴露批量 upsert 的辅助方法。
- **故障转移** – 为重试或多端点部署做计划。
- **版本控制** – 保持适配器语义稳定；如需要，使用 `WithCompatMode`。

---

## 模板

- 复制 `xb/doc/MILVUS_TEMPLATE.go` 作为起点。
- 用你的数据库模式替换负载结构体。
- 将适配器连接到示例/测试中，以便其他人可以从中学习。

---

## 相关文档

- `doc/cn/CUSTOM_VECTOR_DB_GUIDE.md`
- `doc/cn/VECTOR_DB_INTERFACE_DESIGN.md`

