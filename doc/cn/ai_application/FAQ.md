# AI 应用 FAQ

本文档是 `xb/doc/ai_application/FAQ.md` 的中文版本。

---

## 常见问题

**问：我可以在一个构建器中混合 SQL 和向量查询吗？**  
答：可以——向量功能只是向 `Built` 添加额外的元数据，不会破坏 SQL 生成。

**问：我需要 Qdrant 才能使用 xb 吗？**  
答：不需要，xb 可以与普通 SQL 一起工作。Qdrant 是第一个官方向量适配器。

**问：如何调试被跳过的过滤器？**  
答：检查 `built.Conds` 或 `built.SqlOfSelect()` 以查看哪些子句在自动过滤后保留。

---

## 支持

- 打开带有重现步骤的 GitHub issues。
- 加入讨论以提出新的 AI 集成。

