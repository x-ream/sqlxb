# 向量数据库接口设计

本文档是 `xb/doc/VECTOR_DB_INTERFACE_DESIGN.md` 的中文版本。它详细说明了向向量适配器公开的公共接口。

---

## 关键结构体

- `Builder` – 流畅 DSL 入口点。
- `Built` – 由 SQL/JSON 渲染器使用的不可变快照。
- `Custom` – 适配器实现的接口。
- `QdrantCustom`、`MilvusCustom`（示例）– 围绕接口的包装器。

---

## 设计原则

| 原则 | 描述 |
|------|------|
| 最小 API | 保持公共表面小巧；依赖构建器闭包 |
| 类型安全 | 尽可能使用和类型 / 枚举（在 Zig/V 中） |
| 确定性输出 | 相同的输入树 → 相同的 SQL/JSON |
| 可扩展性 | 元数据、验证、拦截器的钩子 |

---

## 必需方法

- `VectorSearch(field string, vector xb.Vector, limit int)`
- `VectorSearchMulti(...)`
- `VectorSearchByIDs(...)`
- `Meta(func(*interceptor.Metadata))`

适配器在生成请求时应尊重这些字段。

---

## 相关文档

- `doc/cn/VECTOR_DB_ABSTRACTION_SUMMARY.md`
- `doc/cn/CUSTOM_INTERFACE.md`

