# Vector Quickstart (English)

English summary of `xb/doc/VECTOR_QUICKSTART.md`. It focuses on starting a vector project with xb + Qdrant/pgvector.

---

## Setup checklist

1. Define DTOs (`struct CodeVector { ... }`).
2. Create embeddings (OpenAI, local model, etc.).
3. Store vectors + metadata in your database of choice.
4. Use xb to build both SQL and vector queries.

---

## Example workflow

```go
queryVec := embedder.Encode(prompt)

json, _ := xb.Of(&CodeVector{}).
    Custom(xb.NewQdrantBuilder().Build()).
    Eq("tenant_id", tenant).
    VectorSearch("embedding", queryVec, 10).
    Build().
    JsonOfSelect()
```

For pgvector:

```go
sql, args, _ := xb.Of(&CodeVector{}).
    Eq("tenant_id", tenant).
    X("embedding <#> ?", queryVec). // custom operator
    Limit(10).
    Build().
    SqlOfSelect()
```

---

## Tips

- Normalize vectors upfront.
- Store timestamps and durability metadata to support re-indexing.
- Add regression tests to ensure vector + scalar filters compose correctly.

---

## Related docs

- `doc/en/VECTOR_GUIDE.md`
- `doc/en/QDRANT_GUIDE.md`

