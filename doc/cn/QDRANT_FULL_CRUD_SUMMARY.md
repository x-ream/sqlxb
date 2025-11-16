# Qdrant 完整 CRUD 总结

本文档是 `xb/doc/QDRANT_FULL_CRUD_SUMMARY.md` 的中文版本。它概述了 xb 如何在 Qdrant 的 CRUD API 中镜像 SQL 语义。

---

## 操作

| 操作 | 构建器入口 | 说明 |
|------|-----------|------|
| Insert | `Insert(func(*InsertBuilder))` + `JsonOfInsert()` | 支持批量插入 |
| Update | `Update(func(*UpdateBuilder))` + `JsonOfUpdate()` | 需要过滤器/ids |
| Delete | `Build()` + `JsonOfDelete()` | 尊重与 SQL 相同的过滤器 DSL |
| Select | `Build()` + `JsonOfSelect()` | 统一的 Search/Recommend/Discover/Scroll |

---

## 常见模式

- 对单条记录操作使用 `Eq("id", id)`。
- `InRequired` 防止意外的批量更新/删除。
- 元数据（`Meta(func)`）被传播，因此审计日志知道谁运行了操作。

---

## 功能对等矩阵

| SQL 概念 | Qdrant 映射 |
|---------|------------|
| WHERE 子句 | `filter.must` / `filter.should` |
| ORDER BY score | 向量搜索固有；显式排序可选 |
| LIMIT/OFFSET | `limit`、`offset`、`scroll_id` |
| With/CTE | 不适用；将逻辑保留在应用层 |

---

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/CUSTOM_INTERFACE.md`

