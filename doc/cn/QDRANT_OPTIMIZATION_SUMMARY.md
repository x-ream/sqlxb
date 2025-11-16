# Qdrant 优化总结

本文档是 `xb/doc/QDRANT_OPTIMIZATION_SUMMARY.md` 的中文版本。它总结了延迟、召回率和多样性的调优技巧。

---

## 性能杠杆

- **HNSW 参数** – 通过 `QdrantCustom` 暴露 `HnswEf`、`ScoreThreshold`、`Exact` 标志。
- **Payload 投影** – 仅包含必要字段以减少 payload 大小。
- **向量归一化** – 预归一化以保持评分可预测。
- **批处理** – 发送批量插入/更新而不是单条记录调用。

---

## 召回率 vs 延迟

| 目标 | 推荐设置 |
|------|---------|
| 低延迟 | 降低 `limit`，适度的 `HnswEf`，启用 payload 修剪 |
| 高召回率 | 增加 `HnswEf`、`limit`，禁用 payload 修剪 |
| 平衡 | 默认 `NewQdrantCustom()` 加上 `WithHashDiversity` |

---

## 可观测性

- 在日志中记录 `TraceID`、`TenantID` 和适配器配置。
- 跟踪每个 API（Search、Recommend、Discover、Scroll）的 P95 延迟。
- 监控 `score` 分布以捕获向量漂移。

---

## 相关文档

- `doc/cn/VECTOR_GUIDE.md`
- `doc/cn/QDRANT_ADVANCED_API.md`
- `doc/cn/AI_APPLICATION.md`

