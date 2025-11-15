# Custom Quickstart (English)

English translation of `xb/doc/CUSTOM_QUICKSTART.md`. It walks through implementing a new adapter in less than 30 minutes.

---

## 1. Skeleton

```go
type MyVectorCustom struct {
    Endpoint string
}

func NewMyVectorCustom(endpoint string) *MyVectorCustom {
    return &MyVectorCustom{Endpoint: endpoint}
}

func (c *MyVectorCustom) Generate(built *xb.Built) (any, error) {
    req := buildPayload(built) // convert conditions to JSON
    return req, nil
}
```

---

## 2. Builder helpers

- `WithCollection(name string)`
- `WithPayloadSelector(selector any)`
- `WithScoreThreshold(th float32)`

Expose helper methods on your custom struct so callers do not touch raw fields.

---

## 3. Testing

```go
func TestMyVectorCustom_Select(t *testing.T) {
    built := xb.Of("vectors").
        Eq("tenant_id", 42).
        VectorSearch("embedding", xb.Vector{0.1, 0.2}, 5).
        Build()

    custom := NewMyVectorCustom("http://localhost:9000")
    payload, err := custom.Generate(built)
    require.NoError(t, err)
    snapshot.AssertJSON(t, payload)
}
```

---

## 4. Documentation

- Describe presets (`NewMyVectorCustom`, `MyVectorCustomHighPrecision`).
- Link to JSON samples and behaviour tables.
- Explain compatibility with `JsonOfInsert/Update/Delete` if supported.

---

## 5. Next steps

- Publish your adapter or keep it internal.
- Add metrics/logging to interceptors for observability.
- Contribute experience reports to `doc/en/CUSTOM_VECTOR_DB_GUIDE.md`.

