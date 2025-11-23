# Custom Interface (English)

The `Custom` interface is the heart of xb’s extensibility model. Instead of shipping 50 dialects in the core, xb lets you implement your own serialiser for any database (SQL or vector) in minutes.

---

## 1. Interface contract

```go
type Custom interface {
    Generate(built *xb.Built) (any, error)
}
```

- `built` contains the canonical AST produced by the builder (`select`, `from`, `conditions`, etc.).
- Return either a string (JSON / SQL) or a struct understood by your driver.
- Errors bubble up to `built.JsonOfSelect()` / `built.CustomGenerate()`.

---

## 2. Minimal example

```go
type MyVectorCustom struct{}

func (c *MyVectorCustom) Generate(built *xb.Built) (any, error) {
    payload := map[string]any{
        "collection": built.Table,
        "filter":     built.FilterJSON(),
        "vector":     built.Vector,
    }
    return payload, nil
}

json, err := xb.Of("vectors").
    Custom(&MyVectorCustom{}).
    VectorSearch("embedding", vec, 10).
    Build().
    JsonOfSelect()
```

---

## 3. Design goals

| Goal | Description |
|------|-------------|
| Minimal surface | One interface, one method |
| Isolation | Users can ship custom logic without forking xb |
| Performance | Interface call is ~1ns; no runtime type switches in core |
| Future proof | Users unlock new database features immediately |

---

## 4. Recommended structure

```go
type MilvusCustom struct {
    Endpoint string
    Timeout  time.Duration
}

func NewMilvusBuilder() *MilvusBuilder {
    return &MilvusBuilder{
        custom: &MilvusCustom{
            Endpoint: "http://localhost:19530",
            Timeout:  3 * time.Second,
        },
    }
}

func (c *MilvusCustom) Generate(built *xb.Built) (any, error) {
    req := buildMilvusPayload(built)
    return req, nil
}
```

- Provide constructors (default/high-precision/high-speed) so users can pick presets without touching internal fields.
- Add comments explaining when to use each preset.

---

## 5. Testing your custom

```go
func TestMilvusCustom_Generate(t *testing.T) {
    built := xb.Of("code_vectors").
        VectorSearch("embedding", xb.Vector{0.1, 0.2}, 5).
        Build()

    custom := NewMilvusBuilder().Build()
    payload, err := custom.Generate(built)

    require.NoError(t, err)
    snapshot.AssertJSON(t, payload)
}
```

- Keep these tests in your project; xb does not need to import your custom adapter.
- Use snapshots or golden files to guard against regressions.

---

## 6. When to implement a `Custom`

| Scenario | Example |
|----------|---------|
| Vector DB JSON | Qdrant (official), Milvus, Pinecone |
| SQL dialect quirks | Oracle upsert, ClickHouse `FORMAT` |
| Internal APIs | Company-specific search service |
| Mixed modes | Generate both SQL and JSON from the same builder |

---

## 7. Related docs

- `QDRANT_GUIDE.md` – official reference implementation
- `VECTOR_GUIDE.md` – embedding best practices
- `FILTERING.md` – understand builder-side auto filtration

If you publish a reusable custom adapter, consider adding a short README under `doc/en/` or linking to it from this file.

