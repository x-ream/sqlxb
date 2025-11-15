# RAG Best Practices (English)

English version of `xb/doc/ai_application/RAG_BEST_PRACTICES.md`.

---

## Capture pipeline

1. Normalize text (lowercase, remove boilerplate).
2. Chunk at 512–1024 tokens.
3. Store vectors + metadata (tenant, doc type, updated_at).
4. Keep SQL + vector stores in sync via CDC.

---

## Query pipeline

```go
json, _ := xb.Of(&DocVector{}).
    Custom(qdrant).
    Eq("tenant_id", tenant).
    VectorSearch("embedding", queryVec, 8).
    Build().
    JsonOfSelect()
```

Feed results to your LLM with citations, dedupe by doc ID, and rerank if necessary.


