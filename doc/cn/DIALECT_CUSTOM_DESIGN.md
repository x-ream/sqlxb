# Dialect vs Custom 设计

本文档是 `xb/doc/DIALECT_CUSTOM_DESIGN.md` 的中文版本。它比较了传统 ORM 方言开关与 xb 的 `Custom` 架构。

---

## Dialect 模型（传统 ORM）

- 列出每个数据库的中央枚举（`MySQL`、`PostgreSQL`、`Oracle`、...）。
- 框架在内部实现特定于供应商的 SQL 生成。
- 添加新数据库需要接触核心仓库。
- 随着更多方言堆积，发布节奏变慢。

---

## Custom 模型（xb）

- 单一 `Custom` 接口；适配器存在于用户空间。
- 核心保持小巧和稳定。
- 团队无需等待上游审查即可发布适配器。
- 适用于 SQL、JSON、GRPC、HTTP、CLI 或任何其他执行层。

---

## 何时选择每个

| 场景 | 推荐模型 |
|------|---------|
| 具有标准语法的商品 SQL | 内置 SQL 生成器 |
| 专用引擎（ClickHouse、Qdrant） | Custom |
| 专有内部 API | Custom |
| 需要即时实验 | Custom |

---

## 适配器作者指南

1. 提供类型化的构建器/配置。
2. 记录限制（不支持的子句、最大向量数等）。
3. 将测试保持在适配器附近。
4. 在向其他团队公开 API 时遵循语义版本控制。

有关底层设计动机，请参阅 `doc/cn/CUSTOM_INTERFACE_PHILOSOPHY.md`。

