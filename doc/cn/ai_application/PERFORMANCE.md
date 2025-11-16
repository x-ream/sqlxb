# 性能指南

本文档是 `xb/doc/ai_application/PERFORMANCE.md` 的中文版本。它解释了如何对由 xb 驱动的 RAG/agent 管道进行基准测试。

---

## 指标

- 向量查询延迟（P50/P95）
- SQL 丰富延迟
- LLM 调用时间
- 端到端吞吐量（QPS）

---

## 优化技巧

- 尽可能缓存嵌入和向量结果。
- 限制 Qdrant 的 payload 字段以减少 JSON 大小。
- 同时发出多个 xb 构建器时使用 goroutine 池。
- 分析 Go（pprof）和外部服务。

---

## 相关文档

- `doc/cn/QDRANT_OPTIMIZATION_SUMMARY.md`
- `doc/cn/AI_APPLICATION.md`

