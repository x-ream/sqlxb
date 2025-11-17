# xb Quickstart (English)

This quickstart shows how to build SQL and Qdrant JSON with xb in three steps: define a model, chain conditions, inspect the output. Every snippet compiles against `v1.4.0+`.

---

## 1. Minimal SQL builder

```go
package main

import "github.com/fndome/xb"

type Cat struct {
    ID    uint64  `db:"id"`
    Name  string  `db:"name"`
    Age   uint    `db:"age"`
    Price *float64 `db:"price"`
}

func main() {
    built := xb.Of(&Cat{}).
        Eq("status", 1).
        Gte("age", 3).
        Build()

    sql, args, _ := built.SqlOfSelect()
    // SELECT * FROM t_cat WHERE status = ? AND age >= ?
    // args => [1 3]
    _ = sql
    _ = args
}
```

**Key points**

- `Of()` accepts structs, table names, or builder functions.
- Chainable methods (`Eq`, `Gt`, `Like`, `In`, etc.) auto-skip empty/zero values.
- `Build()` returns a `Built` object; call `SqlOfSelect()` to get SQL + args later.

---

## 2. JOIN + CTE example

```go
report := xb.Of("recent_orders").
    With("recent_orders", func(sb *xb.BuilderX) {
        sb.From("orders o").
            Select("o.id", "o.user_id").
            Gt("o.created_at", since30Days)
    }).
    WithRecursive("team_hierarchy", func(sb *xb.BuilderX) {
        sb.From("users u").
            Select("u.id", "u.manager_id").
            Eq("u.active", true)
    }).
    UNION(xb.ALL, func(sb *xb.BuilderX) {
        sb.From("archived_orders ao").
            Select("ao.id", "ao.user_id")
    }).
    Build()

sql, args, _ := report.SqlOfSelect()
```

**What happened**

- `With()` / `WithRecursive()` register named sub-builders.
- `UNION(kind, fn)` chains multiple builders without raw SQL.
- All builders share the same condition auto-filtering.

---

## 3. Qdrant Recommend / Discover / Scroll

```go
queryVector := xb.Vector{0.1, 0.2, 0.3}

json, err := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantCustom().
            Recommend(func(rb *xb.RecommendBuilder) {
                rb.Positive(123, 456).
                    Negative(789).
                    Limit(20)
            }),
    ).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    Build().
    JsonOfSelect()
```

**Highlights**

- `Custom()` attaches database-specific generators. `QdrantCustom` inspects builder state and emits the right JSON schema automatically.
- `JsonOfSelect()` is the single public entry point for vector payloads. Internally it routes to Recommend / Discover / Scroll / Search.
- Want SQL instead? Skip `Custom()` and call `SqlOfSelect()`.

---

## 4. Printing & debugging

```go
built := xb.Of("t_order").
    Eq("tenant_id", ctx.Tenant()).
    Like("user_name", keyword).
    Limit(50).
    Build()

sql, args, _ := built.SqlOfSelect()
fmt.Printf("SQL => %s\nARGS => %#v\n", sql, args)
```

- You can also call `built.Raw()` to inspect the internal AST for testing.
- For vector payloads, use `json, _ := built.JsonOfSelect()` and pretty-print with `encoding/json`.

---

## 5. Next steps

- **Vector patterns** → `VECTOR_GUIDE.md`
- **Advanced Qdrant usage** → `QDRANT_GUIDE.md`
- **Custom interfaces** → `CUSTOM_INTERFACE.md`
- **Auto-filtering and validation** → `FILTERING.md`

Feel free to open an issue if a scenario is missing from this quickstart. Contributions are always welcome!


