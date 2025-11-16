# 向量数据库抽象总结

本文档是 `xb/doc/VECTOR_DB_ABSTRACTION_SUMMARY.md` 的中文版本。它解释了 xb 如何在不过度牺牲功能的情况下抽象向量数据库。

---

## 目标

- 保持 SQL 和向量工作流的流畅构建器相同。
- 通过 `Custom` 暴露向量特定控制，而不是污染基础 API。
- 使用户能够运行混合管道（SQL → 向量 → SQL）。

---

## 抽象层

| 层 | 职责 |
|----|------|
| Builder | 收集过滤器、排序、向量操作 |
| Built | 用于 SQL/JSON 生成的不可变快照 |
| Custom 适配器 | 渲染 JSON 负载 |
| Transport | 调用者的 HTTP/GRPC 客户端 |

---

## 保留在适配器中的内容

- Payload 模式
- 身份验证/传输
- 高级旋钮（多样性、多向量、payload 投影）
- 后端特定限制

---

## 相关文档

- `doc/cn/CUSTOM_INTERFACE.md`
- `doc/cn/QDRANT_GUIDE.md`

