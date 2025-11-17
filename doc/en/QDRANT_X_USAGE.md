# Qdrant Builder Usage (English)

This document describes how to use `QdrantBuilder` for advanced tuning configuration.

---

## QdrantBuilder entrypoint

```go
xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantBuilder().
            HnswEf(512).
            ScoreThreshold(0.85).
            WithVector(false).
            Build(),
    ).
    Build()
```

`QdrantBuilder` provides a fluent API to configure advanced Qdrant parameters.

---

## Available options

- `HnswEf(int)` - Set HNSW algorithm ef parameter (recommended: 64-256)
- `ScoreThreshold(float32)` - Set minimum similarity threshold (range: 0.0-1.0)
- `WithVector(bool)` - Set whether to return vector data

These options can be combined via method chaining.

---

## Complete examples

```go
// Basic usage
json, _ := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantBuilder().
            HnswEf(512).
            ScoreThreshold(0.85).
            WithVector(false).
            Build(),
    ).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    Build().
    JsonOfSelect()

// Combined with advanced APIs
custom := xb.NewQdrantBuilder().
    HnswEf(256).
    ScoreThreshold(0.8).
    Build()

custom = custom.Recommend(func(rb *xb.RecommendBuilder) {
    rb.Positive(101, 102).Negative(203).Limit(20)
})

json, _ := xb.Of(&CodeVector{}).
    Custom(custom).
    Build().
    JsonOfSelect()
```

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/QDRANT_OPTIMIZATION_SUMMARY.md`

