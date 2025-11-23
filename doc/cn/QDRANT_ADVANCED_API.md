# Qdrant 高级 API

本文档是 `xb/doc/QDRANT_ADVANCED_API.md` 的中文版本。它记录了 xb 如何在 `JsonOfSelect()` 下统一 Recommend、Discover 和 Scroll 工作流。

---

## 1. 功能矩阵

| API | 何时使用 | 构建器入口 |
|-----|---------|-----------|
| Recommend | 对正面与负面反馈进行重排序 | `NewQdrantBuilder().Recommend(func(*RecommendBuilder)).Build()` |
| Discover | 围绕上下文向量探索内容 | `NewQdrantBuilder().Discover(func(*DiscoverBuilder)).Build()` |
| Scroll | 对超大集合进行分页 | `NewQdrantBuilder().ScrollID(string).Build()` |

每个 API 配置都附加到同一个构建器。`JsonOfSelect()` 检查状态并自动发出正确的 JSON 模式。

---

## 2. Recommend 速查表

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

- Positive/negative ID 映射到 Qdrant 的 `positive` / `negative` 字段。
- `Limit`、`WithPayloadSelector`、`WithScoreThreshold` 和多样性辅助方法都可用。

---

## 3. Discover 速查表

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

- 选择目标向量字段，可选地覆盖策略，并像 SQL 条件一样附加过滤器。
- 可与 `WithPayloadSelector`、多样性和元数据注入一起使用。

---

## 4. Scroll 速查表

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

- 使用 `OffsetID` 从最后一条记录恢复。
- 与 `Eq`、`In` 或 `Meta` 结合用于多租户遍历。

---

## 5. 测试和回归

- 添加调用 `JsonOfSelect()` 并比较 JSON 快照的单元测试。
- 覆盖所有三种 API 类型以及默认的 Search 分支。
- 确保如果同时配置 Recommend 和 Discover，`Custom` 返回有意义的错误（xb 选择第一个匹配项）。

---

## 6. 延伸阅读

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/VECTOR_GUIDE.md`
- `doc/cn/AI_APPLICATION.md`

