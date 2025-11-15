# Performance Guide (English)

English summary of `xb/doc/ai_application/PERFORMANCE.md`. It explains how to benchmark RAG/agent pipelines powered by xb.

---

## Metrics

- Vector query latency (P50/P95)
- SQL enrichment latency
- LLM invocation time
- End-to-end throughput (QPS)

---

## Optimization tips

- Cache embeddings and vector results when possible.
- Limit payload fields for Qdrant to reduce JSON size.
- Use goroutine pools when issuing multiple xb builders simultaneously.
- Profile both Go (pprof) and external services.

---

## Related docs

- `doc/en/QDRANT_OPTIMIZATION_SUMMARY.md`
- `doc/en/AI_APPLICATION.md`

