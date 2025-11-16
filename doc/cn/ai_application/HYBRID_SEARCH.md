# 混合搜索

本文档是 `xb/doc/ai_application/HYBRID_SEARCH.md` 的中文版本。它涵盖将标量过滤器与向量排序相结合。

---

## 模式

- **SQL 优先** – 使用 SQL 过滤候选，然后按 ID 进行向量搜索。
- **向量优先** – 检索前 N 个向量，然后与 SQL 连接以进行丰富。
- **混合** – 同时运行两者，使用业务评分合并结果。

---

## 实现技巧

- 使用 `Cond` 块包装可选过滤器。
- 链接到 SQL 时保持向量限制较小（避免大的 `IN` 列表）。
- 监控总延迟；分解 SQL vs 向量成本。

---

## 参考资料

- `doc/cn/VECTOR_GUIDE.md`
- `doc/cn/QDRANT_GUIDE.md`

