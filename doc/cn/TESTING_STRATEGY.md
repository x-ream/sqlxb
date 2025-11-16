# 测试策略

本文档是 `xb/doc/TESTING_STRATEGY.md` 的中文版本。它概述了项目如何防止回归。

---

## 先前的事件

- **向量 vs SQL 分离** – 向量构建器缺少 `And`/`Or` 测试，导致回归。
- **数值零处理** – 浮点数未被覆盖，因此 `Gt("score", 0.0)` 行为不正确。

---

## 当前计划

1. 每个功能都附带 SQL 和向量路径的单元测试。
2. 快照测试保护 `JsonOfSelect()` 负载。
3. 回归套件通过 `go test ./...` 运行并捕获覆盖率数字。
4. 发布检查清单需要手动验证演示项目。

---

## 技巧

- 通过 `built.Raw()` 测试构建器 AST，同时渲染 SQL/JSON。
- 覆盖边界值（`0`、`nil`、空切片）。
- 为每个高级 Qdrant API（Recommend/Discover/Scroll）添加固定装置。

---

## 相关文档

- `doc/cn/ALL_FILTERING_MECHANISMS.md`
- `doc/cn/QDRANT_ADVANCED_API.md`

