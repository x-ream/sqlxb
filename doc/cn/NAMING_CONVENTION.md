# 命名约定

本文档是 `xb/doc/NAMING_CONVENTION.md` 的中文版本。它标准化了项目中构建器、结构体、表和文件的命名方式。

---

## 包和文件

- 在 Go 中使用小写加下划线（`builder_select.go`、`qdrant_custom.go`）。
- 保持包名简短（`xb`、`interceptor`、`qdrant`）。
- 镜像测试文件（`builder_select_test.go`）靠近它们的实现。

---

## 结构体和 DTO

- 导出的结构体：`CamelCase`（`Builder`、`Built`、`QdrantCustom`）。
- 请求/响应 DTO 包含后缀（`ListUserRequest`、`SearchResult`）。
- 如需要，对只读视图对象使用 `RO` 或 `VO` 后缀。

---

## 构建器方法

- 操作使用动词风格（`Select`、`Sort`、`Limit`、`VectorSearch`）。
- DSL 辅助方法使用短名词（`Or`、`And`、`Meta`、`Cond`）。
- 保持链顺序一致以帮助可读性。

---

## 数据库工件

- 表：`snake_case`（`t_user`、`recent_orders`）。
- 列：`snake_case`。
- CTE 别名：简短但描述性（`recent_orders ro`、`team_hierarchy th`）。

---

## 相关文档

- `doc/cn/BUILDER_BEST_PRACTICES.md`
- `doc/cn/TESTING_STRATEGY.md`

