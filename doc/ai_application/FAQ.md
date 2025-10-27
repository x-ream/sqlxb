# 常见问题解答 (FAQ)

## 📋 目录

- [基础问题](#基础问题)
- [向量检索](#向量检索)
- [RAG 应用](#rag-应用)
- [性能优化](#性能优化)
- [集成问题](#集成问题)
- [故障排查](#故障排查)

## 基础问题

### Q1: sqlxb 与传统 ORM 有什么区别？

**A**: sqlxb 的核心区别：

1. **AI-First 设计**
   - 函数式 API，易于 AI 理解和生成
   - 原生支持向量数据库（Qdrant）
   - 自动过滤机制（9 层）

2. **类型安全**
   - 编译时类型检查
   - 避免 SQL 注入
   - IDE 自动补全

3. **向量检索能力**
   ```go
   // sqlxb
   sqlxb.Of(&Doc{}).VectorSearch("embedding", vector).Build()
   
   // 传统 ORM
   // ❌ 不支持向量检索
   ```

### Q2: sqlxb 适合什么场景？

**A**: 推荐使用场景：

✅ **推荐**:
- RAG 应用（文档问答、知识库）
- AI Agent 系统
- 向量相似度搜索
- 混合检索（标量 + 向量）

⚠️ **谨慎使用**:
- 纯关系型 CRUD（使用 GORM 等即可）
- 超复杂的 JOIN 查询
- 遗留系统迁移

### Q3: sqlxb 的学习曲线如何？

**A**: 

- **如果熟悉 Go**: 10-30 分钟上手
- **如果熟悉 ORM**: 30 分钟-1 小时
- **如果是 Go 新手**: 2-4 小时

推荐学习路径：
1. [QUICKSTART.md](./QUICKSTART.md) - 5 分钟体验
2. [RAG_BEST_PRACTICES.md](./RAG_BEST_PRACTICES.md) - 深入学习
3. 示例项目 - 实战练习

## 向量检索

### Q4: 支持哪些向量数据库？

**A**: 当前支持：

1. **Qdrant** (推荐)
   - 高性能
   - 功能完整
   - 易于部署

2. **PostgreSQL + pgvector**
   - 关系型 + 向量
   - 数据一致性好
   - 适合小规模

**未来计划**:
- Milvus (考虑中)
- Weaviate (考虑中)

### Q5: 如何选择向量维度？

**A**: 常见选择：

| 模型 | 维度 | 适用场景 |
|-----|------|---------|
| OpenAI Ada-002 | 1536 | 通用，英文优秀 |
| OpenAI text-embedding-3-small | 1536 | 更快更便宜 |
| OpenAI text-embedding-3-large | 3072 | 最高质量 |
| Cohere embed-multilingual | 768 | 多语言支持 |
| 本地模型 (BERT) | 768 | 私有部署 |

**推荐**:
- 英文为主: OpenAI Ada-002
- 多语言: Cohere multilingual
- 私有部署: Sentence-BERT

### Q6: 向量检索的精度如何？

**A**: 典型指标：

```
Recall@10: 0.90-0.95  (召回率)
Precision@10: 0.85-0.92 (准确率)
Latency (P95): < 50ms  (延迟)
```

影响因素：
- Embedding 模型质量
- 文档分块策略
- Top-K 设置
- 过滤条件

### Q7: 如何提高向量检索的相关性？

**A**: 5 个技巧：

1. **优化分块**
   ```go
   // 保持语义完整性
   chunks := ChunkBySentence(doc, 3-5) // 3-5 句话一块
   ```

2. **混合检索**
   ```go
   // 结合标量过滤
   sqlxb.Of(&Doc{}).
       VectorSearch("embedding", vec).
       Eq("category", "tech").  // 过滤类别
       Gte("score", 0.7)        // 质量过滤
   ```

3. **重排序**
   ```go
   // 使用 Cross-Encoder 精排
   reranked := RerankWithCrossEncoder(results, query)
   ```

4. **多样性控制**
   ```go
   // MMR 算法
   diverse := ApplyMMR(results, lambda=0.7, topK=10)
   ```

5. **查询扩展**
   ```go
   // 生成相关查询
   queries := []string{query, synonym1, synonym2}
   // 合并结果
   ```

## RAG 应用

### Q8: RAG 应用的典型延迟是多少？

**A**: 端到端延迟分解：

```
总延迟 (P95): 250-500ms

组成:
├─ Embedding:    100ms (40%)
├─ 向量检索:      50ms (20%)
├─ 重排序:        30ms (12%)
└─ LLM 生成:     70ms (28%)
```

优化目标:
- < 200ms: 优秀
- 200-500ms: 良好
- > 500ms: 需要优化

### Q9: 如何处理大文档？

**A**: 三种策略：

**策略 1: 层级分块**
```go
// 文档 → 章节 → 段落 → 句子
chunks := HierarchicalChunk(document)
```

**策略 2: 滑动窗口**
```go
// 固定大小 + 重叠
chunks := ChunkWithOverlap(doc, size=500, overlap=50)
```

**策略 3: 语义分块**
```go
// 基于语义边界
chunks := SemanticChunk(doc)  // 自动识别段落
```

**推荐**: 混合使用，根据文档类型选择。

### Q10: 如何避免 RAG 幻觉（Hallucination）？

**A**: 5 个方法：

1. **提高检索质量**
   ```go
   // 设置较高的相似度阈值
   qx.ScoreThreshold(0.75)  // 而非 0.5
   ```

2. **明确指示**
   ```
   Prompt: "仅根据以下文档回答，如果文档中没有相关信息，请明确说明。"
   ```

3. **引用来源**
   ```go
   // 返回文档 ID 和分数
   response.Sources = []Source{{ID: "doc1", Score: 0.92}}
   ```

4. **多文档验证**
   ```go
   // 要求至少 2 个文档支持
   if len(relevantDocs) < 2 {
       return "信息不足，无法回答"
   }
   ```

5. **后处理检查**
   ```go
   // 验证答案是否在文档中出现
   if !ContainsInDocs(answer, docs) {
       return "答案可能不准确"
   }
   ```

## 性能优化

### Q11: 如何提高查询性能？

**A**: 5 个优化点：

1. **索引优化**
   ```go
   // Qdrant HNSW 参数调优
   HNSWConfig{M: 32, EfConstruct: 256}
   ```

2. **批量操作**
   ```go
   // 批量 embedding
   embeddings := BatchEmbed(texts, batchSize=32)
   ```

3. **缓存**
   ```go
   // 缓存热门查询
   cache.Set(queryHash, results, ttl=15*time.Minute)
   ```

4. **并行查询**
   ```go
   // 多个查询并行执行
   results := ParallelSearch(queries)
   ```

5. **连接池**
   ```go
   // 复用连接
   pool := NewConnectionPool(minConns=10, maxConns=100)
   ```

### Q12: 内存占用如何优化？

**A**: 

1. **量化 (Quantization)**
   ```go
   // 使用 int8 量化，减少 75% 内存
   QuantizationConfig{Type: "int8"}
   ```

2. **降维**
   ```go
   // 1536 → 768 维
   reducedVector := PCA(vector, targetDim=768)
   ```

3. **流式处理**
   ```go
   // 不要一次加载所有文档
   for chunk := range documentStream {
       process(chunk)
   }
   ```

### Q13: 支持多少文档规模？

**A**: 性能参考：

| 规模 | 延迟 (P95) | 内存 | 推荐配置 |
|-----|-----------|------|---------|
| < 10万 | < 20ms | 2GB | 单节点 |
| 10-100万 | < 50ms | 8GB | 单节点 + 量化 |
| 100-1000万 | < 100ms | 32GB | 集群 (3节点) |
| > 1000万 | < 200ms | 128GB+ | 集群 (5+节点) |

## 集成问题

### Q14: 如何与 LangChain 集成？

**A**: 详见 [LANGCHAIN_INTEGRATION.md](./LANGCHAIN_INTEGRATION.md)

快速示例：
```python
from langchain.vectorstores import SqlxbVectorStore

vector_store = SqlxbVectorStore(
    backend_url="http://localhost:8080",
    embedding=OpenAIEmbeddings()
)
```

### Q15: 支持哪些编程语言？

**A**: 

**原生支持**:
- ✅ Go (一等公民)

**通过 HTTP API**:
- ✅ Python (LangChain, LlamaIndex)
- ✅ .NET (Semantic Kernel)
- ✅ JavaScript/TypeScript (计划中)
- ✅ Java (计划中)

### Q16: 如何在云平台部署？

**A**: 

**Kubernetes**:
```yaml
# 参考 examples/k8s/deployment.yaml
```

**Docker Compose**:
```yaml
# 参考 examples/docker-compose/rag-stack.yml
```

**Serverless**:
- AWS Lambda: ✅ 支持
- Google Cloud Run: ✅ 支持
- Azure Functions: ✅ 支持

## 故障排查

### Q17: 向量检索返回空结果

**排查步骤**:

1. 检查向量维度
   ```go
   fmt.Println(len(queryVector))  // 应该匹配 collection
   ```

2. 降低相似度阈值
   ```go
   qx.ScoreThreshold(0.5)  // 临时降低
   ```

3. 验证数据存在
   ```bash
   curl http://localhost:6333/collections/docs/points/count
   ```

4. 检查过滤条件
   ```go
   // 移除所有过滤，只测试向量检索
   sqlxb.Of(&Doc{}).VectorSearch("embedding", vec)
   ```

### Q18: Embedding API 调用失败

**常见原因**:

1. **API Key 错误**
   ```bash
   export OPENAI_API_KEY="sk-..."
   ```

2. **速率限制**
   ```go
   // 添加重试逻辑
   for i := 0; i < 3; i++ {
       if emb, err := callAPI(); err == nil {
           return emb
       }
       time.Sleep(time.Second * time.Duration(math.Pow(2, float64(i))))
   }
   ```

3. **文本过长**
   ```go
   // OpenAI 限制 8191 tokens
   if len(text) > 30000 {
       text = text[:30000]  // 截断
   }
   ```

### Q19: 查询很慢怎么办？

**诊断工具**:

```go
import "time"

start := time.Now()
results, _ := executeQuery()
fmt.Printf("查询耗时: %v\n", time.Since(start))

// 分阶段测量
embeddingTime := measureEmbedding()
searchTime := measureSearch()
rerankTime := measureRerank()
```

**常见瓶颈**:
1. Embedding 生成 → 使用缓存
2. 向量检索 → 优化索引
3. 网络延迟 → 本地部署

### Q20: 如何调试 sqlxb 查询？

**方法 1: 打印 SQL**
```go
sql, args, _ := sqlxb.Of(&User{}).Eq("status", "active").Build()
fmt.Printf("SQL: %s\nArgs: %+v\n", sql, args)
```

**方法 2: 使用 Interceptor**
```go
sqlxb.RegisterInterceptor(func(sql string, args []interface{}) {
    log.Printf("[SQLXB] %s | %+v", sql, args)
})
```

**方法 3: Qdrant 查询日志**
```go
built := builder.Build()
json, _ := built.ToQdrantJSON()
fmt.Printf("Qdrant Query: %s\n", json)
```

## 🙋 还有问题？

- 查看 [GitHub Issues](https://github.com/x-ream/sqlxb/issues)
- 加入 [GitHub Discussions](https://github.com/x-ream/sqlxb/discussions)
- 阅读完整文档 [README.md](./README.md)

---

**持续更新中...** 如果您的问题未被列出，欢迎提交 Issue！

