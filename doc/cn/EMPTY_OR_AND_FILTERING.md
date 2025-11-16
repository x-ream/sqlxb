# 空 OR/AND 过滤

本文档是 `xb/doc/EMPTY_OR_AND_FILTERING.md` 的中文版本。它解释了 xb 如何通过丢弃空块来防止 `WHERE ()` 或 `OR ()` 子句。

---

## 行为

- `Or(func(*CondBuilder))` 和 `And(func(*CondBuilder))` 首先评估它们的嵌套构建器。
- 如果嵌套构建器产生零个有效条件，整个块会被完全跳过。
- 这避免了像 `WHERE status = ? AND ()` 这样的 SQL。

---

## 实用技巧

- 使用 `Bool(condition, func(*CondBuilder))` 显式包装可选块。
- 对于调试，使用 `%+v` 记录 `built.Conds` 以查看哪些块保留了下来。
- 如果需要强制插入空块（非常罕见），请手动使用 `X("...")`。

---

## 相关文档

- `doc/cn/ALL_FILTERING_MECHANISMS.md`
- `doc/cn/FILTERING.md`

