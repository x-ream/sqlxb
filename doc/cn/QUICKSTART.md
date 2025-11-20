# xb 快速上手（中文）

本文展示如何在 3 个步骤内用 xb 构建 SQL 与 Qdrant JSON。示例基于 `v1.4.0+`。

---

## 1. 最简 SQL 构建

```go
package main

import "github.com/fndome/xb"

type Cat struct {
    ID    uint64   `db:"id"`
    Name  string   `db:"name"`
    Age   uint     `db:"age"`
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

**要点**

- `Of()` 可接收结构体指针、表名或 Builder 函数。
- 链式方法（`Eq/Gt/Like/In` 等）会自动跳过空字符串、零值或 nil。
- `Build()` 返回 `Built`，需要 SQL 时再调用 `SqlOfSelect()`。

---

## 2. JOIN + CTE 示例

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

**说明**

- `With/WithRecursive` 注册具名子查询。
- `UNION(kind, fn)` 用同一 API 组合多个 builder。
- 所有子 builder 共用自动过滤规则。

---

## 3. Qdrant Recommend / Discover / Scroll

```go
queryVector := xb.Vector{0.1, 0.2, 0.3}

json, err := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantBuilder().
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

**亮点**

- `Custom()` 附加方言生成器。`QdrantCustom` 会根据状态自动输出 Recommend / Discover / Scroll JSON。
- 对向量场景统一使用 `JsonOfSelect()`。需要 SQL 时不调用 `Custom()` 即可。

---

## 4. 打印与调试

```go
built := xb.Of("t_order").
    Eq("tenant_id", ctx.Tenant()).
    Like("user_name", keyword).
    Limit(50).
    Build()

sql, args, _ := built.SqlOfSelect()
fmt.Printf("SQL => %s\nARGS => %#v\n", sql, args)
```

- 向量场景可 `json, _ := built.JsonOfSelect()` 后用 `encoding/json` 美化输出。
- `built.Raw()` 能看到内部 AST，在测试里很实用。

---

## 5. 下一步

- **向量进阶** → `VECTOR_GUIDE.md`
- **Qdrant 高级 API** → `QDRANT_GUIDE.md`
- **Custom 接口** → `CUSTOM_INTERFACE.md`
- **自动过滤与校验** → `FILTERING.md`

遇到缺失的场景，欢迎提 issue 或 PR。

