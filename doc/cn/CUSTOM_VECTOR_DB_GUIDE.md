# Custom 向量数据库指南

本文档基于 `xb/doc/CUSTOM_VECTOR_DB_GUIDE.md`。它总结了使用 `Custom` 接口构建向量数据库适配器的最佳实践。

---

## 1. 适配器蓝图

| 关注点 | 建议 |
|--------|------|
| Payload 模式 | 定义镜像数据库 JSON 的 Go 结构体以避免 map 操作 |
| 过滤器编码 | 尽可能重用 `built.FilterJSON()` |
| 向量序列化 | 接受 `[]float32` 和预归一化向量 |
| 分页 | 根据后端支持 `limit`、`offset`、`scroll_id` |

---

## 2. 预设构造函数

- `NewQdrantBuilder()` – 官方参考
- `NewMilvusCustom()` – 包含服务器端点 + GRPC 选项
- `NewWeaviateCustom()` – 处理类名 + where 过滤器

暴露针对特定工作负载调整的变体，如 `HighPrecision`、`HighRecall` 或 `HighSpeed`。

---

## 3. 映射构建器功能

| 构建器功能 | 向量数据库映射 |
|-----------|---------------|
| `VectorSearch` | `search`/`recommend` API |
| `Eq`/`In` | filter `must` 子句 |
| `Meta` | 请求级元数据（trace、tenant） |
| `Limit`、`Offset` | `limit`、`offset`、`scroll` |

记录不支持的组合，以便调用者知道何时回退到原始 API。

---

## 4. 运营技巧

- 在执行前验证集合/命名空间是否存在。
- 逐字显示后端错误；它有助于错误配置的负载。
- 添加可选的日志记录钩子，以便在不接触应用程序代码的情况下调试负载。

---

## 5. 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/VECTOR_GUIDE.md`
- `doc/cn/AI_APPLICATION.md`

