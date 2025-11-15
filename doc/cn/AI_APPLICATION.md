# AI 应用指南（中文）

本页整合 `xb/doc/ai_application/` 旧文档的核心内容，聚焦三种场景：RAG、Agent 工具、以及依赖 xb 的分析仪表盘。

---

## 1. RAG 流程

### 1.1 数据处理清单

1. **切片**：按 512-1024 token 分段。
2. **嵌入**：保存向量、原文、元数据（`tenant_id`, `lang`, `tags`）。
3. **索引**：按租户或业务拆分 Qdrant collection。
4. **同步**：SQL 与向量库通过 CDC 或异步任务保持一致。

### 1.2 查询示例

```go
queryVec := embedder.Encode(prompt)

resultsJSON, _ := xb.Of(&DocVector{}).
    Custom(qdrantCustom).
    Eq("tenant_id", tenantID).
    VectorSearch("embedding", queryVec, 8).
    Build().
    JsonOfSelect()
```

将 `resultsJSON` 传给 LangChain/LlamaIndex/Semantic Kernel 等框架即可，因为 xb 输出原生 Qdrant JSON。

---

## 2. Agent 工具

- **工具模式**：把常用 builder 预设封装为 agent tool（如 `recommend_feed`）。
- **守卫**：通过 `Meta(func)` 注入 agent/session ID，在拦截器中执行配额或审计。
- **流式场景**：针对 Scroll，使用 `ScrollID` 分批返回结果。

示例：

```go
func RecommendFeedTool(input RecommendInput) (string, error) {
    json, err := xb.Of(&FeedVector{}).
        Custom(qdrantCustom).
        Eq("tenant_id", input.Tenant).
        VectorSearch("embedding", input.Vector, input.Limit).
        Build().
        JsonOfSelect()
    if err != nil {
        return "", err
    }
    return string(json), nil
}
```

---

## 3. 分析仪表盘

即便前端只消费 SQL，也可以借助 xb 统一记录：

- `Meta(func)` 中写入 `TraceID`、`UserID`、`Model`。
- 注册 `AfterBuild` 拦截器，将 SQL/JSON 摘要送到观测系统。
- 同时持久化 `SqlOfSelect()` 与 `JsonOfSelect()`，方便关联。

---

## 4. 集成提示

| 工具链 | 建议 |
|--------|------|
| LangChain | 将 xb builder 封成 `Runnable`，直接返回 `JsonOfSelect` |
| LlamaIndex | 实现自定义 retriever，使用 xb 生成的 JSON |
| Semantic Kernel | 通过 `IQueryFunction` 将 xb 调用暴露给技能 |

---

## 5. 延伸阅读

- `QUICKSTART.md`：基础 API
- `QDRANT_GUIDE.md`：高级向量调用
- `VECTOR_GUIDE.md`：嵌入与多样性实践
- `FILTERING.md`：自动过滤机制

如果你有重复可复用的 AI 流程，欢迎在此文件追加案例或提交 PR。

