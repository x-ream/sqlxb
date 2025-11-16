# Custom JOINs 指南

本文档是 `xb/doc/CUSTOM_JOINS_GUIDE.md` 的中文版本。它解释了如何扩展 xb 的 JOIN DSL 以支持方言特定的语法，例如 ClickHouse `GLOBAL JOIN` 或 `ASOF JOIN`。

---

## 内置 JOIN 目录

```go
const (
    inner_join      = "INNER JOIN"
    left_join       = "LEFT JOIN"
    right_join      = "RIGHT JOIN"
    cross_join      = "CROSS JOIN"
    asof_join       = "ASOF JOIN"
    global_join     = "GLOBAL JOIN"
    full_outer_join = "FULL OUTER JOIN"
)
```

使用 `FromX(func(*FromBuilder))` 和 `JOIN(kind)` 辅助方法来组合多步骤管道。

---

## 添加你自己的 JOIN

```go
func LATERAL() xb.JOIN {
    return func() string {
        return "LATERAL JOIN"
    }
}
```

- 返回一个闭包，以便构建器可以延迟调用它。
- 如果需要额外配置，注册简短的辅助方法（`func (fb *FromBuilder) Lateral(...)`）。

---

## 最佳实践

- 保持 JOIN 辅助方法纯函数；它们应该只格式化 SQL 关键字。
- 重用 `Cond(func(*ON))` 来定义谓词；这保持自动过滤完整。
- 对于方言特定的提示（`GLOBAL`、`FINAL` 等），通过 `FromX` 配置暴露它们，以便链的其余部分保持可移植。

---

## 示例

- 混合子查询和 ON 子句的多 join 管道。
- 注入优化器提示的自定义 join。
- 在每个 ON 块内强制执行租户过滤器的跨分片 join。

更多模式请参阅 `doc/cn/BUILDER_BEST_PRACTICES.md`。

