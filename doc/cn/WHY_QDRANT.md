# 为什么选择 Qdrant？

本文档是 `xb/doc/WHY_QDRANT.md` 的中文版本。它解释了为什么 Qdrant 成为第一个官方支持的向量后端。

---

## 选择 Qdrant 的原因

1. **功能速度** – 频繁发布，包含 Recommend、Discover、Scroll、payload 选择器。
2. **运营成熟度** – 水平扩展、快照、云 + 自托管选项。
3. **JSON 优先设计** – 易于与 xb 的 `Custom.Generate` 集成。
4. **开源友好** – 宽松的许可证和强大的社区。

---

## 比较快照

| 引擎 | 说明 |
|------|------|
| Qdrant | 功能平衡，可预测的 API |
| Milvus | 规模大但操作更复杂 |
| Pinecone | 仅托管；封闭 API |
| Weaviate | 也是 JSON 优先但模式管理更重 |

---

## 路线图对齐

- Qdrant 的高级 API 清晰地映射到 xb 的 DSL。
- 多样性辅助方法镜像 Qdrant 的 payload 选择器。
- 未来功能（重排序、多向量）可以通过 `Custom` 添加。

---

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/QDRANT_ADVANCED_API.md`
- `doc/cn/VECTOR_GUIDE.md`

