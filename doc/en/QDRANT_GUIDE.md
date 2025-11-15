# Qdrant Guide (English)

`JsonOfSelect()` is now the single entry point for every Qdrant workflow. This guide explains how xb routes to Search, Recommend, Discover, and Scroll under the hood and how to configure each scenario.

---

## 1. Basics

```go
custom := xb.NewQdrantCustom().
    Namespace("code_vectors").
    WithPayload(true)

json, err := xb.Of(&CodeVector{}).
    Custom(custom).
    Eq("language", "go").
    VectorSearch("embedding", vec, 20).
    Build().
    JsonOfSelect()
```

- `Custom()` is mandatory for vector payloads. Without it, `JsonOfSelect()` returns an error.
- Conditions automatically populate `filter.must`.
- `VectorSearch(field, vector, limit)` controls the `vector` block and `limit`.

---

## 2. Recommend API

```go
json, err := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantCustom().
            Recommend(func(rb *xb.RecommendBuilder) {
                rb.Positive(111, 222).
                    Negative(333).
                    Limit(40).
                    WithPayloadSelector(map[string]any{
                        "include": []string{"id", "title"},
                    })
            }),
    ).
    Build().
    JsonOfSelect()
```

What happens:

- `RecommendBuilder` stores IDs for positive/negative feedback.
- When `JsonOfSelect()` sees `recommend` settings inside `Built`, it emits a `/points/recommend` payload automatically.

---

## 3. Discover API

```go
json, err := xb.Of(&ArticleVector{}).
    Custom(
        xb.NewQdrantCustom().
            Discover(func(db *xb.DiscoverBuilder) {
                db.
                    TargetVector("news_vector", queryVec).
                    Strategy("best_score").
                    Filter(func(f *xb.QFilterBuilder) {
                        f.MustEq("region", "us")
                    })
            }),
    ).
    Build().
    JsonOfSelect()
```

- `TargetVector` sets the primary vector, `Filter` adds extra metadata constraints.
- Discover payloads share the same `JsonOfSelect()` call site as Search and Recommend.

---

## 4. Scroll API

```go
json, err := xb.Of(&FeedVector{}).
    Custom(
        xb.NewQdrantCustom().
            Scroll(func(sb *xb.ScrollBuilder) {
                sb.
                    PayloadSelector([]string{"id", "tags"}).
                    Limit(100).
                    OffsetID("9001:5")
            }),
    ).
    Build().
    JsonOfSelect()
```

- Use Scroll for pagination across large collections.
- `OffsetID` corresponds to Qdrant’s `offset` field. Leave it empty to start from the beginning.

---

## 5. Diversity helpers

```go
custom := xb.NewQdrantCustom().
    WithHashDiversity(func(h *xb.HashDiversity) {
        h.Field = "category"
        h.Modulo = 6
    }).
    WithMinDistance(0.35)
```

- Hash diversity keeps similar categories from clustering at the top.
- `WithMinDistance` enforces a lower bound on cosine distance.

---

## 6. Payload selectors

```go
custom := xb.NewQdrantCustom().
    WithPayloadSelector(map[string]any{
        "include": []string{"id", "title", "lang"},
        "exclude": []string{"debug"},
    })
```

- You can set selectors globally via the custom object or inside the Recommend/Discover builders.
- Selectors merge with builder-level settings (builder overrides global defaults).

---

## 7. Debugging tips

| Issue | Fix |
|-------|-----|
| `JsonOfSelect failed: Custom is nil` | Attach `xb.NewQdrantCustom()` before `Build()` |
| Wrong API called | Ensure only one of Recommend/Discover/Scroll is configured per builder |
| Missing filters | Check `should_skip` rules—empty strings or zero values are ignored |
| Unexpected limit | Set limit both in `VectorSearch` and the advanced builder |

---

## 8. Related docs

- `VECTOR_GUIDE.md` – embedding hygiene & hybrid patterns
- `CUSTOM_INTERFACE.md` – how to implement your own vector DB custom
- `FILTERING.md` – explains why some filters might be skipped automatically

If you add support for another advanced API (e.g., rerank), document it here and update `xb/qdrant_custom.go` tests accordingly.

