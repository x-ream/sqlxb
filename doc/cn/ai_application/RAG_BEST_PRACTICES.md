# RAG 最佳实践

本文档是 `xb/doc/ai_application/RAG_BEST_PRACTICES.md` 的中文版本。

---

## 捕获管道

1. 规范化文本（小写，删除样板）。
2. 在 512–1024 个令牌处分块。
3. 存储向量 + 元数据（租户、文档类型、updated_at）。
4. 通过 CDC 保持 SQL + 向量存储同步。

---

## 查询管道

```go
json, _ := xb.Of(&DocVector{}).
    Custom(qdrant).
    Eq("tenant_id", tenant).
    VectorSearch("embedding", queryVec, 8).
    Build().
    JsonOfSelect()
```

将结果提供给你的 LLM，并附上引用，按文档 ID 去重，必要时重新排序。

