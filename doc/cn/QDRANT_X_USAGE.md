# Qdrant Builder 使用指南

本文档描述了如何使用 `QdrantBuilder` 进行高级调优配置。

---

## QdrantBuilder 入口点

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

`QdrantBuilder` 提供了链式 API 来配置 Qdrant 的高级参数。

---

## 可用选项

- `HnswEf(int)` - 设置 HNSW 算法的 ef 参数（推荐值: 64-256）
- `ScoreThreshold(float32)` - 设置最小相似度阈值（范围: 0.0-1.0）
- `WithVector(bool)` - 设置是否返回向量数据

这些选项可以通过链式调用组合使用。

---

## 完整示例

```go
// 基础用法
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

// 与高级 API 结合使用
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

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/QDRANT_OPTIMIZATION_SUMMARY.md`

