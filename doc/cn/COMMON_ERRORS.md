# 常见错误

本文档是 `xb/doc/COMMON_ERRORS.md` 的中文版本。快速参考常见错误，在提交 issue 前使用此检查清单。

---

## 1. 构建器使用

- **缺少 `Build()`** – 链式调用返回的是构建器，不是 SQL。始终以 `.Build()` 结尾。
- **过早调用 `SqlOfSelect()`** – 只有 `Built` 有 SQL 生成方法。
- **混合指针/值构建器** – 在 Go 中，保持单一模式（`xb.Of(&User{})`）以避免意外的副本。
- **Nil 指针 DTO** – 在传递给 `Of()` 之前检查输入。

---

## 2. 自动过滤的意外情况

- `Eq("status", 0)` 消失，因为零值被自动过滤。
- `In("id")` 使用空切片时会崩溃；切换到 `InRequired`。
- `Like("name", "")` 被忽略；首先清理输入。

---

## 3. Custom 适配器

- `JsonOfSelect failed: Custom is nil` → 附加 `xb.NewQdrantBuilder().Build()` 或你自己的适配器。
- `Custom.Generate` 返回不支持的类型 → 坚持使用 `string`、`[]byte` 或驱动程序就绪的结构体。
- 确保你的 `Custom` 尊重构建器状态（`Built.Table`、`Built.Conds` 等）。

---

## 4. SQL 执行

- 直接使用 `SqlOfSelect` 返回的参数切片；不要手动重新排序。
- 对于分页，始终同时调用 `Limit` 和 `Offset` 以避免驱动程序默认值; 而且SqlOfPage(), 可以返回：countSql, dataSql, vs, metaMap
- 通过 `X()` 连接自定义片段时，清理它们以防止 SQL 注入。

---

## 5. 调试工作流

1. 打印 `built.SqlOfSelect()` 和 `built.Args`（或其他语言中的 `built.ArgsAsStrings()`）。
2. 检查 `built.Raw()` 以确保 AST 符合预期。
3. 围绕失败场景添加回归测试。

更多边缘情况请参阅 `doc/cn/TESTING_STRATEGY.md` 和 `doc/cn/FILTERING.md`。

