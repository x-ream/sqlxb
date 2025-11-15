# Qdrant X Usage (English)

English translation of `xb/doc/QDRANT_X_USAGE.md`. It describes how to use the experimental `QdrantX` extension points for advanced tuning.

---

## QdrantX entrypoint

```go
xb.Of(&CodeVector{}).
    QdrantX(func(qx *xb.QdrantBuilderX) {
        qx.HnswEf(512).
            ScoreThreshold(0.85).
            WithVector(false)
    })
```

`QdrantX` exposes lower-level knobs before we formalize them into top-level helpers.

---

## Typical options

- `HnswEf(int)`
- `ScoreThreshold(float32)`
- `WithVector(bool)`
- `Exact(bool)`
- `ShardKey(string)`

Combine them cautiously; some options are mutually exclusive depending on the backend version.

---

## Migration guidance

- Whenever a `QdrantX` knob graduates to a first-class helper (`WithMinDistance`, etc.), prefer the new helper.
- Keep `QdrantX` usage localized so future refactors are painless.
- Add tests covering both `QdrantX` and the official helper to ensure parity.

---

## Related docs

- `doc/en/QDRANT_GUIDE.md`
- `doc/en/QDRANT_OPTIMIZATION_SUMMARY.md`

