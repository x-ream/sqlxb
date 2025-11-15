# Qdrant 指南（中文）

`JsonOfSelect()` 现已统一处理 Search、Recommend、Discover、Scroll。本文说明 xb 如何选择对应 API 以及每种场景的配置方式。

---

## 1. 基础用法

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

- 使用向量功能必须调用 `Custom()`，否则 `JsonOfSelect()` 会报错。
- 条件被映射到 `filter.must`。
- `VectorSearch(field, vector, limit)` 控制 `vector` 与 `limit`。

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

- `RecommendBuilder` 存储正/负样本、limit、payload 选择器。
- 一旦 builder 中存在 recommend 配置，`JsonOfSelect()` 自动输出 `/points/recommend` 请求体。

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

- `TargetVector` 设置主向量，`Filter` 补充额外限制。
- Discover 与 Search/Recommend 共用同一个 `JsonOfSelect()` 入口。

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

- Scroll 用于在大集合中分页遍历。
- `OffsetID` 对应 Qdrant 的 `offset` 字段。

---

## 5. 多样性与附加参数

```go
custom := xb.NewQdrantCustom().
    WithHashDiversity(func(h *xb.HashDiversity) {
        h.Field = "category"
        h.Modulo = 6
    }).
    WithMinDistance(0.35)
```

- Hash diversity 避免相似分类挤在顶部。
- `WithMinDistance` 设置分数下限。

---

## 6. Payload Selector

```go
custom := xb.NewQdrantCustom().
    WithPayloadSelector(map[string]any{
        "include": []string{"id", "title", "lang"},
        "exclude": []string{"debug"},
    })
```

- 可全局设置，也能在 Recommend/Discover builder 内单独覆盖。

---

## 7. 排障

| 问题 | 解决方法 |
|------|----------|
| `JsonOfSelect failed: Custom is nil` | 在 `Build()` 前附加 `xb.NewQdrantCustom()` |
| 触发错误的 API | 确保同一 builder 中只配置一种高级 API |
| 过滤条件丢失 | 检查是否被自动跳过（空/零值） |
| limit 不正确 | 同时在 `VectorSearch` 和高级 builder 中设置 |

---

## 8. 相关文档

- `VECTOR_GUIDE.md`：嵌入规范与混合模式
- `CUSTOM_INTERFACE.md`：自定义其他向量数据库方言
- `FILTERING.md`：条件被忽略的原因

若你扩展了新的高级 API（如 rerank、payload projection），也请记得补充本文件与 `xb/qdrant_custom.go` 的测试。

