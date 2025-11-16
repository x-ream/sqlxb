# Custom 接口哲学

本文档是 `xb/doc/CUSTOM_INTERFACE_PHILOSOPHY.md` 的中文版本。它描述了为什么 xb 选择 `Custom` 接口而不是硬编码数十种方言。

---

## 原则

1. **最小表面** – 一个方法（`Generate`）处理每个操作（SQL 或 JSON）。
2. **用户所有权** – 团队可以实现小众数据库，而无需等待核心发布。
3. **可预测行为** – 无论后端如何，都使用相同的构建器 API。
4. **可组合性** – 使用单个流畅链组合 SQL + 向量工作流。

---

## Dialect vs Custom

| Dialect 枚举 | Custom 接口 |
|-------------|------------|
| 框架捆绑每个数据库 | 框架保持小巧；用户扩展它 |
| 新数据库需要核心 PR | 用户在自己的仓库中发布适配器 |
| 难以保持同步 | 适配器所有者按自己的节奏迭代 |

---

## 设计规则

- 保持适配器无状态或易于克隆。
- 返回类型必须是驱动程序友好的（`string`、`[]byte` 或薄结构体）。
- 提供构建器（`NewQdrantCustom()`、`NewMilvusCustom()`）而不是暴露原始结构体。
- 记录高级旋钮，以便用户可以推理权衡。

---

## 推荐阅读

- `doc/cn/CUSTOM_INTERFACE.md`
- `doc/cn/CUSTOM_QUICKSTART.md`
- `doc/cn/CUSTOM_VECTOR_DB_GUIDE.md`

