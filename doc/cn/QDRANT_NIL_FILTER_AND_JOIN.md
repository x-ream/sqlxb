# Qdrant Nil 过滤和 JOIN 说明

本文档是 `xb/doc/QDRANT_NIL_FILTER_AND_JOIN.md` 的中文版本。它记录了 xb 如何处理 nil/空过滤器以及 JOIN 风格的构造如何映射到 Qdrant 过滤器。

---

## Nil 过滤行为

- 空字符串、零值和 `nil` 指针在生成 Qdrant 过滤器之前会被移除。
- `InRequired` 强制至少保留一个 ID/字段值。
- 即使没有过滤器子句保留，`Meta` 数据仍会被发出，支持多租户日志记录。

---

## JOIN 类比

虽然 Qdrant 不支持 SQL join，但 xb 通过过滤器组合模拟相关行为：

- 使用嵌套的 `Cond` 块来表示 `AND`/`OR` 组。
- 对于父子文档，将父 ID 存储在 payload 字段中并按它们过滤。
- 多样性辅助方法通过确保不同的 payload 桶出现在结果中来充当隐式 join。

---

## 建议

- 预先规范化 payload 键以避免不匹配的过滤器。
- 保持布尔字段显式（`true`/`false`），因为空值会被跳过。
- 编写涵盖 nil 过滤 + 向量组合的回归测试以捕获边缘情况。

---

## 相关文档

- `doc/cn/ALL_FILTERING_MECHANISMS.md`
- `doc/cn/QDRANT_GUIDE.md`

