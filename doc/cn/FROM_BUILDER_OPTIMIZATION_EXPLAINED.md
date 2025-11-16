# From Builder 优化说明

本文档是 `xb/doc/FROM_BUILDER_OPTIMIZATION_EXPLAINED.md` 的中文版本。它深入探讨了 `FromBuilder` 如何在保持 SQL 确定性的同时重写复杂的 JOIN/CTE 图。

---

## 目标

1. **可预测别名** – 每个子查询/CTE 获得确定性别名。
2. **流式 JOIN 规划** – `FromX` 接收一个函数，以便它可以内联或提升子查询。
3. **租户安全** – 当辅助方法调用 `WITH` 块时，自动将守卫注入到嵌套源中。

---

## 优化过程

- **扁平化** – 当它们只添加投影时，合并连续的 `FromX` 块。
- **修剪** – 删除没有下游子句引用的未使用 CTE。
- **提示传播** – 将优化器提示（例如，`FINAL`、`SAMPLE`）传递到子 join。

---

## 调试

- 使用 `builder.DebugFrom()`（内部辅助方法）记录分阶段的 AST。
- 单元测试可以在 SQL 渲染之前快照中间表示。
- 添加新的 join 类型时，确保 AST 和 SQL 层都支持它们。

---

## 另请参阅

- `doc/cn/CUSTOM_JOINS_GUIDE.md`
- `doc/cn/BUILDER_BEST_PRACTICES.md`

