# 向量快速开始

本文档是 `xb/doc/VECTOR_QUICKSTART.md` 的中文版本。它专注于使用 xb + Qdrant/pgvector 启动向量项目。

---

## 设置检查清单

1. 定义 DTO（`struct CodeVector { ... }`）。
2. 创建嵌入（OpenAI、本地模型等）。
3. 在你选择的数据库中存储向量 + 元数据。
4. 使用 xb 构建 SQL 和向量查询。

---

## 示例工作流

```go
queryVec := embedder.Encode(prompt)

json, _ := xb.Of(&CodeVector{}).
    Custom(xb.NewQdrantCustom()).
    Eq("tenant_id", tenant).
    VectorSearch("embedding", queryVec, 10).
    Build().
    JsonOfSelect()
```

对于 pgvector：

```go
sql, args, _ := xb.Of(&CodeVector{}).
    Eq("tenant_id", tenant).
    X("embedding <#> ?", queryVec). // 自定义操作符
    Limit(10).
    Build().
    SqlOfSelect()
```

---

## 技巧

- 预先归一化向量。
- 存储时间戳和持久性元数据以支持重新索引。
- 添加回归测试以确保向量 + 标量过滤器正确组合。

---

## 相关文档

- `doc/cn/VECTOR_GUIDE.md`
- `doc/cn/QDRANT_GUIDE.md`

