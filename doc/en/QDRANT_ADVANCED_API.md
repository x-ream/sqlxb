# Qdrant Advanced API (English)

English rewrite of `xb/doc/QDRANT_ADVANCED_API.md`. It documents how xb unifies Recommend, Discover, and Scroll workflows under `JsonOfSelect()`.

---

## 1. Feature matrix

| API | When to use | Builder entry |
|-----|-------------|---------------|
| Recommend | Rerank positive vs negative feedback | `QdrantCustom.Recommend(func(*RecommendBuilder))` |
| Discover | Explore content around a context vector | `QdrantCustom.Discover(func(*DiscoverBuilder))` |
| Scroll | Paginate extremely large collections | `QdrantCustom.Scroll(func(*ScrollBuilder))` |

Each API config attaches to the same builder. `JsonOfSelect()` inspects the state and emits the proper JSON schema automatically.

---

## 2. Recommend cheat sheet

```go
json, _ := xb.Of(&FeedVector{}).
    Custom(
        xb.NewQdrantBuilder().
            Recommend(func(rb *xb.RecommendBuilder) {
                rb.Positive(501, 502).
                    Negative(999).
                    Limit(40).
                    WithPayloadSelector(map[string]any{
                        "include": []string{"id", "title"},
                    })
            }).
            Build()
    ).
    Build().
    JsonOfSelect()
```

- Positive/negative IDs map to Qdrant’s `positive` / `negative` fields.
- `Limit`, `WithPayloadSelector`, `WithScoreThreshold`, and diversity helpers are available.

---

## 3. Discover cheat sheet

```go
json, _ := xb.Of(&ArticleVector{}).
    Custom(
        xb.NewQdrantBuilder().
            Discover(func(db *xb.DiscoverBuilder) {
                db.TargetVector("topic_vec", queryVec).
                    Strategy("best_score").
                    Filter(func(f *xb.QFilterBuilder) {
                        f.MustEq("region", "us")
                    })
            }).
            Build()
    ).
    Build().
    JsonOfSelect()
```

- Choose a target vector field, optionally override strategy, and attach filters just like SQL conditions.
- Works with `WithPayloadSelector`, diversity, and metadata injection.

---

## 4. Scroll cheat sheet

```go
json, _ := xb.Of(&FeedVector{}).
    Custom(
        xb.NewQdrantBuilder().
            Scroll(func(sb *xb.ScrollBuilder) {
                sb.PayloadSelector([]string{"id", "tags"}).
                    Limit(100).
                    OffsetID("9012:5")
            })
    ).
    Build().
    JsonOfSelect()
```

- Use `OffsetID` to resume from the last record.
- Combine with `Eq`, `In`, or `Meta` for multi-tenant traversal.

---

## 5. Testing and regression

- Add unit tests that call `JsonOfSelect()` and compare JSON snapshots.
- Cover all three API types plus the default Search branch.
- Ensure `Custom` returns meaningful errors if both Recommend and Discover are configured simultaneously (xb picks the first match).

---

## 6. Further reading

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/VECTOR_GUIDE.md`
- `doc/en/AI_APPLICATION.md`

