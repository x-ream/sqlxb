# sqlxb AI 应用生态

## 📋 概述

本目录包含 sqlxb 在 AI 应用场景下的集成指南、最佳实践和示例代码。sqlxb 的 AI-First 设计使其成为 RAG、AI Agent 和向量检索应用的理想选择。

## 🎯 核心特性

### 1. AI-Friendly API 设计
- **函数式 API**: 易于 AI 理解和生成
- **编译时类型安全**: 减少运行时错误
- **自动过滤机制**: 智能处理 nil/0 值
- **可预测的行为**: 相同输入总是产生相同输出

### 2. 向量数据库原生支持
- **PostgreSQL pgvector**: 关系型 + 向量能力
- **Qdrant 深度集成**: 高性能向量检索
- **混合检索**: 标量过滤 + 向量搜索
- **多种距离度量**: Cosine, L2, Inner Product

### 3. RAG 场景优化
- **Chunk 存储模式**: 高效的文档分块存储
- **元数据管理**: 灵活的过滤和检索
- **重排序支持**: MMR, Hash, Distance 多样性策略

## 📚 文档导航

### 快速开始
- [AI Agent 工具链](./AGENT_TOOLKIT.md) - JSON Schema、OpenAPI 规范生成
- [RAG 最佳实践](./RAG_BEST_PRACTICES.md) - 文档检索应用构建指南

### 生态集成
- [LangChain 集成](./LANGCHAIN_INTEGRATION.md) - Python LangChain 集成示例
- [LlamaIndex 集成](./LLAMAINDEX_INTEGRATION.md) - Python LlamaIndex 集成示例
- [Semantic Kernel 集成](./SEMANTIC_KERNEL_INTEGRATION.md) - .NET Semantic Kernel 集成

### 高级主题
- [自然语言查询转换](./NL2SQL.md) - 实验性 NL → SQL 转换
- [混合检索策略](./HYBRID_SEARCH.md) - 标量 + 向量混合检索
- [性能优化指南](./PERFORMANCE.md) - AI 应用性能调优

## 🚀 快速示例

### 基础 RAG 查询

```go
package main

import (
    "github.com/x-ream/sqlxb"
)

type DocumentChunk struct {
    ID        int64     `json:"id"`
    Content   string    `json:"content"`
    Embedding []float32 `json:"embedding"`
    Metadata  string    `json:"metadata"` // JSON
    DocID     *int64    `json:"doc_id"`   // 原文档ID（非主键，可为 nil）
}

func SearchSimilarChunks(queryVector []float32, limit int) (string, []interface{}, error) {
    return sqlxb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector).
        Limit(limit).
        Build()
}
```

### Qdrant 混合检索

```go
func HybridSearch(queryVector []float32, docType string, minScore float64) (string, error) {
    built := sqlxb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector, 20).  // 返回 20 条
        Eq("doc_type", docType).                      // 标量过滤
        Ne("status", "deleted").                      // 排除已删除
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(float32(minScore))
        }).
        Build()

    return built.ToQdrantJSON()
}
```

### LangChain 向量存储

```python
from langchain.vectorstores import Qdrant
from langchain.embeddings import OpenAIEmbeddings

# 使用 sqlxb 生成的 Qdrant 查询
vector_store = Qdrant(
    client=qdrant_client,
    collection_name="documents",
    embeddings=OpenAIEmbeddings(),
)

results = vector_store.similarity_search_with_score(
    query="如何使用 sqlxb?",
    k=5,
    filter={
        "must": [
            {"key": "doc_type", "match": {"value": "tutorial"}}
        ]
    }
)
```

## 🏗️ 典型应用架构

### RAG 知识库系统

```
┌─────────────────────────────────────────────────────────┐
│                      用户查询                            │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                 LLM Embedding API                        │
│              (OpenAI, Cohere, 本地模型)                  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                  sqlxb 查询构建                          │
│   • VectorSearch() - 向量相似度                          │
│   • Eq/Ne/In - 元数据过滤                                │
│   • WithScoreThreshold - 相关性阈值                      │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Qdrant / PostgreSQL+pgvector                │
│              (向量数据库层)                              │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   重排序 & 后处理                        │
│   • MMR 多样性算法                                       │
│   • 元数据增强                                          │
│   • 上下文合并                                          │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                  LLM 生成回答                            │
│              (GPT-4, Claude, Llama)                      │
└─────────────────────────────────────────────────────────┘
```

### AI Agent 工具链

```
┌─────────────────────────────────────────────────────────┐
│                   AI Agent (GPT-4)                       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│              Function Calling / Tool Use                 │
│   • search_knowledge_base()                              │
│   • filter_by_metadata()                                 │
│   • get_similar_documents()                              │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│           sqlxb 动态查询构建 (JSON Schema)               │
│   • 参数验证                                            │
│   • 查询优化                                            │
│   • 安全检查                                            │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    数据库执行                           │
└─────────────────────────────────────────────────────────┘
```

## 🎯 适用场景

### ✅ 推荐使用场景

1. **RAG 应用**
   - 企业知识库
   - 客户服务机器人
   - 文档问答系统
   - 代码搜索引擎

2. **AI Agent 系统**
   - 自主数据分析 Agent
   - 多源信息聚合 Agent
   - 智能推荐系统
   - 个性化内容生成

3. **混合检索应用**
   - 电商商品搜索（文本+图像）
   - 多模态内容检索
   - 语义化日志分析
   - 智能告警系统

4. **向量数据库应用**
   - 相似内容推荐
   - 图像/音频检索
   - 用户行为分析
   - 异常检测

### ⚠️ 不适用场景

- 纯关系型 OLTP（使用标准 ORM 即可）
- 超大规模分析查询（考虑专用 OLAP）
- 实时流处理（考虑 Flink/Spark）

## 🧪 示例项目

### 1. RAG 知识库 (Go + Qdrant)
- **目录**: `examples/rag-knowledge-base/`
- **技术栈**: Go, sqlxb, Qdrant, OpenAI
- **功能**: 文档上传、分块、向量化、检索

### 2. AI Agent 工具 (Python + LangChain)
- **目录**: `examples/langchain-agent/`
- **技术栈**: Python, LangChain, sqlxb (via API)
- **功能**: 自然语言查询、多工具调用、结果合成

### 3. 混合检索 API (Go + PostgreSQL)
- **目录**: `examples/hybrid-search-api/`
- **技术栈**: Go, sqlxb, PostgreSQL+pgvector
- **功能**: REST API、标量+向量检索、结果排序

## 📊 性能参考

### 向量检索性能 (Qdrant)

| 数据规模 | Top-K | 响应时间 (P95) | 吞吐量 (QPS) |
|---------|-------|----------------|-------------|
| 100K    | 10    | 15ms           | 2000        |
| 1M      | 10    | 30ms           | 1000        |
| 10M     | 10    | 80ms           | 400         |

### 混合检索性能 (PostgreSQL+pgvector)

| 数据规模 | 过滤条件 | 响应时间 (P95) | 吞吐量 (QPS) |
|---------|---------|----------------|-------------|
| 100K    | 2个      | 50ms           | 500         |
| 1M      | 2个      | 120ms          | 200         |
| 10M     | 2个      | 400ms          | 80          |

*测试环境: 4C8G, SSD, 本地部署*

## 🔧 开发工具

### JSON Schema 生成器
```bash
go run ./tools/gen-json-schema -type DocumentChunk -output schema.json
```

### OpenAPI 规范生成器
```bash
go run ./tools/gen-openapi -package main -output openapi.yaml
```

### 自然语言查询测试
```bash
go run ./tools/nl2sql -query "查找最近7天创建的活跃用户"
```

## 🤝 贡献指南

我们欢迎社区贡献新的 AI 应用集成和示例！

### 贡献类型
1. **新框架集成**: Haystack, AutoGen, CrewAI 等
2. **示例项目**: 完整可运行的 AI 应用
3. **最佳实践**: 性能优化、架构模式
4. **文档翻译**: 英文、日文等

### 提交流程
1. Fork 项目
2. 创建特性分支
3. 添加文档/代码
4. 提交 Pull Request

详见: [CONTRIBUTING.md](../CONTRIBUTING.md)

## 📞 支持与反馈

- **GitHub Issues**: 问题报告和功能建议
- **GitHub Discussions**: AI 应用讨论和经验分享
- **示例问题**: 提供完整的可复现代码

## 📄 许可证

与 sqlxb 主项目相同，采用 MIT 许可证。

---

**让 AI 和人类都能轻松构建高性能数据库应用！** 🚀

