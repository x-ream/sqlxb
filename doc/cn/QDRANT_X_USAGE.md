# Qdrant X 使用指南

本文档是 `xb/doc/QDRANT_X_USAGE.md` 的中文版本。它描述了如何使用实验性的 `QdrantX` 扩展点进行高级调优。

---

## QdrantX 入口点

```go
xb.Of(&CodeVector{}).
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HnswEf(512).
            ScoreThreshold(0.85).
            WithVector(false)
    })
```

`QdrantX` 在我们将它们正式化为顶级辅助方法之前暴露较低级别的旋钮。

---

## 典型选项

- `HnswEf(int)`
- `ScoreThreshold(float32)`
- `WithVector(bool)`
- `Exact(bool)`
- `ShardKey(string)`

谨慎组合它们；某些选项根据后端版本相互排斥。

---

## 迁移指南

- 每当 `QdrantX` 旋钮升级为一级辅助方法（`WithMinDistance` 等）时，优先使用新的辅助方法。
- 保持 `QdrantX` 使用本地化，以便未来的重构无痛。
- 添加涵盖 `QdrantX` 和官方辅助方法的测试以确保对等。

---

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/QDRANT_OPTIMIZATION_SUMMARY.md`

