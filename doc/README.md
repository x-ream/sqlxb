# xb Documentation

## 📚 文档目录

### 📅 项目规划

- **[v1.0.0 路线图](./ROADMAP_v1.0.0.md)** - 从 v0.9.1 到 v1.0.0 的完整规划（目标: 2026年6月）
- **[v1.0.0 检查清单](./V1_CHECKLIST.md)** - v1.0.0 发布的核心任务和进度追踪

---

### 🚀 快速开始

- **[VECTOR_QUICKSTART.md](./VECTOR_QUICKSTART.md)** - 向量数据库快速开始指南
- **[VECTOR_DIVERSITY_QDRANT.md](./VECTOR_DIVERSITY_QDRANT.md)** - 向量多样性和 Qdrant 使用指南

---

### 🤖 AI 应用生态

- **[AI 应用文档入口](./ai_application/README.md)** - AI/RAG/Agent 应用集成完整指南
- **[PageIndex 集成指南](./PAGEINDEX_INTEGRATION.md)** - 结构化文档检索（Vectify AI PageIndex + xb）
- **[使用场景决策指南](./USE_CASE_GUIDE_ZH.md)** - 向量数据库 vs PageIndex vs 传统 SQL 选型指南

**核心文档**:
- **[AI Agent 工具链](./ai_application/AGENT_TOOLKIT.md)** - JSON Schema、OpenAPI 集成
- **[RAG 最佳实践](./ai_application/RAG_BEST_PRACTICES.md)** - 文档检索应用指南
- **[LangChain 集成](./ai_application/LANGCHAIN_INTEGRATION.md)** - Python LangChain
- **[LlamaIndex 集成](./ai_application/LLAMAINDEX_INTEGRATION.md)** - Python LlamaIndex
- **[Semantic Kernel 集成](./ai_application/SEMANTIC_KERNEL_INTEGRATION.md)** - .NET SK
- **[混合检索策略](./ai_application/HYBRID_SEARCH.md)** - 标量 + 向量混合
- **[NL2SQL](./ai_application/NL2SQL.md)** - 自然语言查询转换
- **[性能优化](./ai_application/PERFORMANCE.md)** - AI 应用性能调优
- **[FAQ](./ai_application/FAQ.md)** - 常见问题

---

### 📖 核心文档

#### 最佳实践

- **[BUILDER_BEST_PRACTICES.md](./BUILDER_BEST_PRACTICES.md)** - Builder 使用最佳实践
- **[COMMON_ERRORS.md](./COMMON_ERRORS.md)** - 常见错误和解决方法

#### 完整应用示例

- **[Examples](../examples/README.md)** - 完整应用示例代码
  - [PostgreSQL + pgvector 应用](../examples/pgvector-app/)
  - [Qdrant 集成应用](../examples/qdrant-app/)
  - [RAG 检索应用](../examples/rag-app/)
  - [PageIndex 文档结构化检索](../examples/pageindex-app/)

#### API 设计

- **[QDRANT_NIL_FILTER_AND_JOIN.md](./QDRANT_NIL_FILTER_AND_JOIN.md)** - nil/0 过滤和 JOIN 方案

#### Qdrant 专题

- **[WHY_QDRANT.md](./WHY_QDRANT.md)** - 为什么选择 Qdrant
- **[QDRANT_X_USAGE.md](./QDRANT_X_USAGE.md)** - QdrantX 使用指南
- **[QDRANT_ADVANCED_API.md](./QDRANT_ADVANCED_API.md)** - Recommend, Discover, Scroll API (v0.10.0)
- **[QDRANT_API_SYNC_STRATEGY.md](./QDRANT_API_SYNC_STRATEGY.md)** - Qdrant API 同步策略

#### 高级主题

- **[FROM_BUILDER_OPTIMIZATION_EXPLAINED.md](./FROM_BUILDER_OPTIMIZATION_EXPLAINED.md)** - FROM 构建器优化详解

#### 扩展指南

- **[CUSTOM_VECTOR_DB_GUIDE.md](./CUSTOM_VECTOR_DB_GUIDE.md)** - 自定义向量数据库支持指南
- **[CUSTOM_JOINS_GUIDE.md](./CUSTOM_JOINS_GUIDE.md)** - 自定义 JOIN 扩展指南

---

### 🧪 测试与质量

- **[TESTING_STRATEGY.md](./TESTING_STRATEGY.md)** - 测试策略与改进计划（v0.9.1 新增）
- **[ALL_FILTERING_MECHANISMS.md](./ALL_FILTERING_MECHANISMS.md)** - 完整的自动过滤机制（9层）
- **[EMPTY_OR_AND_FILTERING.md](./EMPTY_OR_AND_FILTERING.md)** - 空 OR/AND 过滤详解

---

### 🤖 AI-First 开发

- **[CONTRIBUTORS.md](./CONTRIBUTORS.md)** - 贡献者和 AI-First 协作模型
- **[AI_MAINTAINABILITY_ANALYSIS.md](./AI_MAINTAINABILITY_ANALYSIS.md)** - AI 可维护性分析
- **[MAINTENANCE_STRATEGY.md](./MAINTENANCE_STRATEGY.md)** - 维护策略（80/15/5 模型）

---

### 📋 发布文档

- **[RELEASE_NOTES_v0.9.0.md](./RELEASE_NOTES_v0.9.0.md)** - v0.9.0 发布说明
- **[RELEASE_v0.9.0_GUIDE.md](./RELEASE_v0.9.0_GUIDE.md)** - v0.9.0 发布指南

---

### 🛠️ 其他

- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - 贡献指南
- **[GITHUB_ISSUE_TEMPLATE.md](./GITHUB_ISSUE_TEMPLATE.md)** - GitHub Issue 模板

---

## 🎯 推荐阅读路径

### 路径 1: 快速了解（10 分钟）

1. [VECTOR_QUICKSTART.md](./VECTOR_QUICKSTART.md) - 5 分钟
2. [VECTOR_DIVERSITY_QDRANT.md](./VECTOR_DIVERSITY_QDRANT.md) - 5 分钟

---

### 路径 2: 深入理解（30 分钟）

1. [VECTOR_DIVERSITY_QDRANT.md](./VECTOR_DIVERSITY_QDRANT.md) - 10 分钟
2. [ALL_FILTERING_MECHANISMS.md](./ALL_FILTERING_MECHANISMS.md) - 10 分钟
3. [WHY_QDRANT.md](./WHY_QDRANT.md) - 10 分钟

---

### 路径 3: 完整掌握（90 分钟）

1. 阅读路径 2 的所有文档
2. [AI_MAINTAINABILITY_ANALYSIS.md](./AI_MAINTAINABILITY_ANALYSIS.md) - 15 分钟
3. [FROM_BUILDER_OPTIMIZATION_EXPLAINED.md](./FROM_BUILDER_OPTIMIZATION_EXPLAINED.md) - 15 分钟
4. [CUSTOM_VECTOR_DB_GUIDE.md](./CUSTOM_VECTOR_DB_GUIDE.md) - 15 分钟（可选）
5. [CUSTOM_JOINS_GUIDE.md](./CUSTOM_JOINS_GUIDE.md) - 15 分钟（可选）

---

## 📊 文档统计

- **总文档数**: 41
- **项目规划**: 3 (ROADMAP, V1_CHECKLIST, TESTING_STRATEGY)
- **用户指南**: 3 (VECTOR_QUICKSTART, VECTOR_DIVERSITY_QDRANT, USE_CASE_GUIDE_ZH) ← v0.10.4 新增
- **最佳实践**: 2 (BUILDER_BEST_PRACTICES, COMMON_ERRORS) ← v0.10.3 新增
- **完整示例**: 4 (pgvector-app, qdrant-app, rag-app, pageindex-app) ← v0.10.4 新增
- **AI 应用生态**: 13 (ai_application 目录 + PAGEINDEX_INTEGRATION) ← v0.10.4 新增
  - Agent 工具链、RAG 最佳实践、LangChain/LlamaIndex/SK 集成
  - 混合检索、NL2SQL、性能优化、FAQ 等
- **API 设计**: 1 (QDRANT_NIL_FILTER_AND_JOIN)
- **Qdrant 专题**: 4 (WHY_QDRANT, QDRANT_X_USAGE, QDRANT_ADVANCED_API, QDRANT_API_SYNC_STRATEGY)
- **扩展指南**: 2 (CUSTOM_VECTOR_DB_GUIDE, CUSTOM_JOINS_GUIDE)
- **测试与质量**: 3 (TESTING_STRATEGY, ALL_FILTERING, EMPTY_OR_AND)
- **AI-First 开发**: 3 (CONTRIBUTORS, AI_MAINTAINABILITY, MAINTENANCE_STRATEGY)
- **发布文档**: 2 (RELEASE_NOTES_v0.9.0, RELEASE_v0.9.0_GUIDE)
- **其他**: 3 (README, CONTRIBUTING, FROM_BUILDER_OPTIMIZATION)

---

## 🔗 快速链接

- **[返回主页](../README.md)**
- **[GitHub Repository](https://github.com/fndome/xb)**
- **[pkg.go.dev](https://pkg.go.dev/github.com/fndome/xb)**

