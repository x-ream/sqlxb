# Custom Interface README (English)

This file mirrors `xb/doc/CUSTOM_INTERFACE_README.md` and serves as an English landing page for everything related to `Custom`.

---

## Why Custom?

- Database ecosystems grow faster than any single ORM can keep up.
- A single interface lets us cover relational SQL, vector stores, time-series, and proprietary services.
- Users can ship adapters in minutes, not weeks.

---

## Quick start

```go
type MyDBCustom struct{}

func (c *MyDBCustom) Generate(built *xb.Built) (any, error) {
    return map[string]any{
        "query": built.Table,
        "filter": built.FilterJSON(),
    }, nil
}

json, _ := xb.Of("docs").
    Custom(&MyDBCustom{}).
    Eq("tenant_id", tenant).
    Build().
    JsonOfSelect()
```

---

## Suggested reading order

1. `doc/en/CUSTOM_INTERFACE_PHILOSOPHY.md`
2. `doc/en/CUSTOM_QUICKSTART.md`
3. `doc/en/CUSTOM_VECTOR_DB_GUIDE.md`
4. `doc/en/CUSTOM_ARCHITECTURE_VALIDATION.md`

---

## Community adapters

- Qdrant (official) – see `xb/qdrant_custom.go`
- Milvus, Weaviate, Pinecone – community experiments
- Internal services – teams typically wrap HTTP/GRPC endpoints

Document your adapter in `doc/en/CUSTOM_VECTOR_DB_GUIDE.md` once it stabilizes so others can learn from it.

