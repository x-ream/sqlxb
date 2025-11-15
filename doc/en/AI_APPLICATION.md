# AI Application Starter (English)

This page consolidates guidance that used to live under `xb/doc/ai_application/`. It focuses on three workflows: Retrieval-Augmented Generation (RAG), Agent tooling, and analytics dashboards backed by xb.

---

## 1. Retrieval-Augmented Generation (RAG)

### 1.1 Data pipeline checklist

1. **Chunking** – split documents into 512-1024 token windows.
2. **Embedding** – store the raw vector, the text, and metadata (`tenant_id`, `lang`, `tags`).
3. **Indexing** – choose a Qdrant collection per tenant or per domain.
4. **Sync metadata** – keep SQL + vector stores in lockstep via CDC or async jobs.

### 1.2 Query pipeline

```go
queryVec := embedder.Encode(prompt)

resultsJSON, _ := xb.Of(&DocVector{}).
    Custom(qdrantCustom).
    Eq("tenant_id", tenantID).
    VectorSearch("embedding", queryVec, 8).
    Build().
    JsonOfSelect()
```

Feed `resultsJSON` to the RAG orchestrator (LangChain, LlamaIndex, Semantic Kernel, etc.). Each toolkit consumes the same payload because xb sticks to raw Qdrant JSON.

---

## 2. Agent tooling

- **Tool schema** – expose builder presets as agent tools (`recommend_feed`, `discover_code`, `scroll_notifications`).
- **Guard rails** – use `Meta(func)` to inject agent/session IDs and enforce quotas in interceptors.
- **Streaming** – for long-running scrolls, paginate via `ScrollID` and send chunks to the agent runtime.

Example tool definition:

```go
func RecommendFeedTool(input RecommendInput) (string, error) {
    json, err := xb.Of(&FeedVector{}).
        Custom(qdrantCustom).
        Eq("tenant_id", input.Tenant).
        VectorSearch("embedding", input.Vector, input.Limit).
        Build().
        JsonOfSelect()
    if err != nil {
        return "", err
    }
    return string(json), nil
}
```

---

## 3. Analytics dashboards

Even if the front-end is SQL-only, xb helps keep AI activity auditable:

- Use `Meta(func)` to log `TraceID`, `UserID`, `Model` per query.
- Register `AfterBuild` interceptors to push query summaries to observability systems.
- Store `built.SqlOfSelect()` alongside `built.JsonOfSelect()` to correlate SQL + vector traces.

---

## 4. Integration notes

| Toolkit | Tip |
|---------|-----|
| LangChain | Wrap xb builders as `Runnable` objects; use `JsonOfSelect` output directly |
| LlamaIndex | Register a custom retriever that calls xb; keep metadata consistent |
| Semantic Kernel | Implement `IQueryFunction` to invoke xb and return embeddings |

---

## 5. Where to go next

- `QUICKSTART.md` for the core builder API
- `QDRANT_GUIDE.md` for advanced vector payloads
- `VECTOR_GUIDE.md` for embedding hygiene and diversity settings
- `FILTERING.md` to understand how inputs are sanitized automatically

If you have a repeatable AI workflow (RAG, agent, analytics) worth sharing, PRs to this file are welcome.

