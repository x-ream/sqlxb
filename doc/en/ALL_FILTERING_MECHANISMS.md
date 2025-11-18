# All Filtering Mechanisms (English)

This document is the English counterpart of `xb/doc/ALL_FILTERING_MECHANISMS.md`. It explains how xb automatically skips empty inputs, how `InRequired` protects destructive operations, and how higher level blocks behave when their internal conditions collapse.

---

## 1. Single-value guards

- `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `Like`, `LikeLeft`
- Empty strings, zero numbers, `false` booleans, and `nil` pointers are ignored.
- `time.Time` values are always serialized (converted to `YYYY-MM-DD HH:MM:SS`).

```go
xb.Of("t_user").
    Eq("status", 0).      // skipped
    Eq("status", 1).      // included
    Like("name", "").     // skipped
    Like("name", "ai").   // included
    Build()
```

---

## 2. `IN` / `NOT IN`

- Entire clauses are removed if every argument is zero, empty, or `nil`.
- Mixed lists are cleaned before rendering.
- `InRequired` panics when the remaining list is empty, which is perfect for admin bulk actions.

```go
xb.Of("orders").
    In("id", 0, nil, 9, 10).         // renders IN (9,10)

xb.Of("orders").
    InRequired("id", ids...).        // panic if ids collapses to empty
```

---

## 3. Compound blocks

- `Or(func(*CondBuilder))`, `And(func(*CondBuilder))`, `Cond(func(*CondBuilder))`
- If the nested builder produces zero effective conditions, the entire block is dropped, preventing `WHERE ()`.

---

## 4. LIKE helpers

- `Like`, `NotLike`, `LikeLeft` all skip empty strings.
- `doLike` injects the proper wildcard placement (`%foo%`,  `foo%`).

---

## 5. Debugging tips

| Symptom | Explanation |
|---------|-------------|
| Condition missing | Value was auto-filtered; inspect `built.Conds` |
| IN disappeared | Input collapsed to zero elements |
| OR block missing | Every nested condition skipped |
| Need strict enforcement | Swap to `InRequired` or add `X()` raw expressions |

Use `fmt.Printf("%#v\n", built.Conds)` in tests to inspect the final AST.

---

## 6. Related material

- `doc/en/FILTERING.md` – condensed overview
- `xb/doc/EMPTY_OR_AND_FILTERING.md` – legacy Chinese reference
- `doc/en/TESTING_STRATEGY.md` – ideas for asserting automatic filtering in tests

