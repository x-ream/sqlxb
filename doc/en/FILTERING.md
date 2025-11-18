# Filtering & Validation (English)

xb aggressively skips empty inputs so you don’t have to write `if value == "" { ... }` before every condition. This document explains what gets filtered, how to override the defaults, and how to plug in validation logic.

---

## 1. Single-value conditions

Methods: `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `Like`, `LikeLeft`.

| Type | Skipped values |
|------|----------------|
| `string` | `""` |
| `int`, `uint`, `float` | `0` |
| `bool` | `false` |
| pointers | `nil` |
| `time.Time` | never skipped (converted to string) |

Example:

```go
builder := xb.Of("t_user").
    Eq("status", 0).   // skipped
    Eq("status", 1).   // added
    Like("name", "").  // skipped
    Like("name", "ai") // added
```

---

## 2. `IN` / `NOT IN`

`In`, `NotIn`, `InRequired`.

- Entire call skipped if the variadic list is empty or contains only zero/empty/nil values.
- Nil pointers and zero numbers are removed from the list.
- `InRequired` panics if the final list is empty (useful for enforcing guards).

```go
builder := xb.Of("t").
    In("id", 0, nil, 9, 10) // becomes IN (9, 10)
```

---

## 3. Compound blocks

- `Or(func(cb *CondBuilder))`
- `And(func(cb *CondBuilder))`
- `Cond(func(cb *CondBuilder))`

If the nested builder produces zero valid conditions, the block is discarded (no empty `()` in SQL).

---

## 4. Validation hooks

Use interceptors or `Meta(func(meta *interceptor.Metadata))` to run custom validation.

```go
xb.RegisterBeforeBuild(func(built *xb.Built) error {
    if built.HasInRequiredViolation() {
        return errors.New("missing required IDs")
    }
    return nil
})
```

Useful scenarios:

- Ensure tenant or org ID is present.
- Enforce paging limits.
- Audit changes by inspecting `built.Meta`.

---

## 5. Debugging skipped conditions

| Symptom | Fix |
|---------|-----|
| Condition missing from SQL | Check if the value was zero/empty |
| IN collapsed to empty | Switch to `InRequired` to catch it early |
| OR block gone | All inner conditions were skipped |
| Validation not triggered | Register interceptor before calling `Build()` |

To inspect the internal state:

```go
built := xb.Of("t").Eq("name", "").Build()
fmt.Printf("%#v\n", built.Conds)
```

---

## 6. Related docs

- `QUICKSTART.md` for basic chaining
- `CUSTOM_INTERFACE.md` if you need custom validation in vector adapters
- `QDRANT_GUIDE.md` to see how filters map to Qdrant payloads

If you discover a common pattern that should not be auto-skipped, open an issue so we can document or extend the behavior.

