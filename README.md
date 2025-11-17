# xb (eXtensible Builder)
[![OSCS Status](https://www.oscs1024.com/platform/badge/fndome/xb.svg?size=small)](https://www.oscs1024.com/project/fndome/xb?ref=badge_small)
![workflow build](https://github.com/fndome/xb/actions/workflows/go.yml/badge.svg)
[![GitHub tag](https://img.shields.io/github/tag/fndome/xb.svg?style=flat)](https://github.com/fndome/xb/tags)
[![Go Report Card](https://goreportcard.com/badge/github.com/fndome/xb)](https://goreportcard.com/report/github.com/fndome/xb)

> Languages: **English** | [中文](./doc/cn/README.md)

`xb` is an AI-first SQL/JSON builder for relational + vector databases. One fluent API builds:

- SQL for `database/sql`, `sqlx`, `gorm`, any raw driver
- JSON for Qdrant (Recommend / Discover / Scroll) and other vector stores
- Hybrid pipelines that mix SQL tables with vector similarity

Everything flows through `Custom()` + `Build()` so the surface stays tiny even as capabilities grow.

> **Notes**
> - Persistence structs mirror the database schema, so numeric primary keys can stay as plain values.
> - Request/filter DTOs should declare non-primary numeric and boolean fields as pointers to distinguish “unset” from “0/false” and to leverage autobypass logic.
> - Need to bypass optimizations? Use `X("...")` to inject raw SQL (the clause will never be auto-skipped), and pick explicit JOIN helpers (e.g., `JOIN(NON_JOIN)` or custom builders) when you want to keep every JOIN even if it looks redundant. For `BuilderX`, call `WithoutOptimization()` to disable the JOIN/CTE optimizer entirely.
> - For non-functional control flow inside fluent chains, use `Any(func(*BuilderX))` to run loops or helper functions without breaking chaining, and `Bool(func() bool, func(*CondBuilder))` to conditionally add blocks while reusing the auto-filtered DSL.

---

## Highlights

- **Unified vector entry** — `JsonOfSelect()` now covers all Qdrant search/recommend/discover/scroll flows. Legacy `ToQdrant*JSON()` methods were retired.
- **Composable SQL** — `With/WithRecursive` and `UNION(kind, fn)` let you express ClickHouse-style analytics directly in Go.
- **Smart condition DSL** — auto-filter nil/zero, guard rails via `InRequired`, raw expressions via `X()`, reusable subqueries via `CondBuilderX.Sub()`, and inline conditional blocks.
- **Adaptive JOIN planner** — `FromX` + `JOIN(kind)` skip meaningless joins automatically (e.g., empty ON blocks), keeping SQL lean.
- **Observability-first** — `Meta(func)` plus interceptors carry TraceID/UserID across builder stages.
- **AI-assisted maintenance** — code, tests, docs co-authored by AI and reviewed by humans every release.

📦 **Latest**: [v1.4.0](./RELEASE_v1.4.0.md) — QdrantBuilder API + documentation improvements.

---

## Quickstart

### Build SQL
```go
package main

import "github.com/fndome/xb"

type Cat struct {
    ID    uint64   `db:"id"`
    Name  string   `db:"name"`
    Age   *uint     `db:"age"`
    Price *float64 `db:"price"`
}

func main() {
    built := xb.Of(&Cat{}).
        Eq("status", 1).
        Gte("age", 3).
        Build()

    sql, args, _ := built.SqlOfSelect()
    // SELECT * FROM t_cat WHERE status = ? AND age >= ?
    _ = sql
    _ = args
}
```

### Qdrant vector search
```go
queryVector := xb.Vector{0.1, 0.2, 0.3}

json, err := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantCustom().
            Recommend(func(rb *xb.RecommendBuilder) {
                rb.Positive(123, 456).Negative(789).Limit(20)
            }),
    ).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    Build().
    JsonOfSelect()

if err != nil {
    panic(err)
}
// POST json to /collections/{name}/points/recommend
```

---

## Advanced Capabilities

### CTE + UNION pipelines
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
    Meta(func(meta *interceptor.Metadata) {
        meta.TraceID = traceID
        meta.Set("source", "dashboard")
    }).
    Build()

sql, args, _ := report.SqlOfSelect()
```

### JOIN builder with subqueries
```go
builder := xb.X().
    Select("p.id", "p.weight").
    FromX(func(fb *xb.FromBuilder) {
        fb.Sub(func(sb *xb.BuilderX) {
            sb.Select("id", "type").
                From("t_pet").
                Gt("id", 10000)
        }).As("p").
            JOIN(xb.INNER).Of("t_dog").As("d").On("d.pet_id = p.id").
            JOIN(xb.LEFT).Of("t_cat").As("c").On("c.pet_id = p.id")
    }).
    Ne("p.type", "PIG")

sql, args, _ := builder.Build().SqlOfSelect()
```

### Qdrant Recommend / Discover / Scroll
- Configure via `QdrantCustom.Recommend/Discover/ScrollID`.
- `JsonOfSelect()` inspects builder state and emits the correct JSON schema.
- Compatible with diversity helpers (`WithHashDiversity`, `WithMinDistance`) and standard filters.

### Interceptors & Metadata
- Register global `BeforeBuild` / `AfterBuild` hooks (see `xb/interceptor`).
- `Meta(func)` injects metadata before hooks run — perfect for tracing, tenancy, or experiments.

### Dialect & Custom
- Dialects (`dialect.go`) let you swap quoting rules, placeholder styles, and vendor-specific predicates without rewriting builders — see [`doc/en/DIALECT_CUSTOM_DESIGN.md`](./doc/en/DIALECT_CUSTOM_DESIGN.md) / [`doc/cn/DIALECT_CUSTOM_DESIGN.md`](./doc/cn/DIALECT_CUSTOM_DESIGN.md).
- `Custom()` is the escape hatch for vector DBs and bespoke backends: plug in `Custom` implementations, emit JSON via `JsonOfSelect()`, or mix SQL + vector calls in one fluent chain. Deep dives live in [`doc/en/CUSTOM_VECTOR_DB_GUIDE.md`](./doc/en/CUSTOM_VECTOR_DB_GUIDE.md) / [`doc/cn/CUSTOM_VECTOR_DB_GUIDE.md`](./doc/cn/CUSTOM_VECTOR_DB_GUIDE.md).
- Need Oracle/Milvus/other dialects? Implement a tiny interface `Custom`, register it once, and the fluent chains instantly start outputting those drivers’ SQL/JSON schemas without forking the builder core.

---

## Documentation

| Topic | English | Chinese |
|-------|---------|---------|
| Overview & Index | [doc/en/README.md](./doc/en/README.md) | [doc/cn/README.md](./doc/cn/README.md) |
| Quickstart | [doc/en/QUICKSTART.md](./doc/en/QUICKSTART.md) | [doc/cn/QUICKSTART.md](./doc/cn/QUICKSTART.md) |
| Qdrant Guide | [doc/en/QDRANT_GUIDE.md](./doc/en/QDRANT_GUIDE.md) | [doc/cn/QDRANT_GUIDE.md](./doc/cn/QDRANT_GUIDE.md) |
| Vector Guide | [doc/en/VECTOR_GUIDE.md](./doc/en/VECTOR_GUIDE.md) | [doc/cn/VECTOR_GUIDE.md](./doc/cn/VECTOR_GUIDE.md) |
| Custom Interface | [doc/en/CUSTOM_INTERFACE.md](./doc/en/CUSTOM_INTERFACE.md) | [doc/cn/CUSTOM_INTERFACE.md](./doc/cn/CUSTOM_INTERFACE.md) |
| Auto-filter (nil/0 skip) | [doc/en/ALL_FILTERING_MECHANISMS.md](./doc/en/ALL_FILTERING_MECHANISMS.md) | [doc/cn/FILTERING.md](./doc/cn/FILTERING.md) |
| Join optimization | [doc/en/CUSTOM_JOINS_GUIDE.md](./doc/en/CUSTOM_JOINS_GUIDE.md) | _(coming soon)_ |
| AI Application Starter | [doc/en/AI_APPLICATION.md](./doc/en/AI_APPLICATION.md) | [doc/cn/AI_APPLICATION.md](./doc/cn/AI_APPLICATION.md) |

> We are migrating docs into `doc/en/` + `doc/cn/`. Legacy files remain under `doc/` until the move completes.

---

## Contributing

We welcome issues, discussions, PRs!

- Issues & features: [GitHub Issues](https://github.com/fndome/xb/issues)
- Roadmap & ideas: [GitHub Discussions](https://github.com/fndome/xb/discussions)
- Contribution steps: [CONTRIBUTING](./doc/CONTRIBUTING.md)
- Vision & philosophy: [VISION.md](./VISION.md)

Before opening a PR:
1. Run `go test ./...`
2. Update docs/tests related to your change
3. Describe behavior changes clearly in the PR template

---

## License

Apache License 2.0 — see [LICENSE](./LICENSE).
