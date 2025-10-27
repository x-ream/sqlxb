# sqlxb 完整应用示例

本目录包含 sqlxb 在实际应用中的完整示例代码。

## 📚 示例列表

### 1. [PostgreSQL + pgvector 应用](./pgvector-app/)

**场景**: 代码语义搜索

**技术栈**:
- PostgreSQL + pgvector
- sqlx
- Gin

**功能**:
- 代码片段存储
- 向量搜索
- 混合检索
- 分页查询

**测试**:
- repository_test.go (3个测试)
- 包含集成测试示例

---

### 2. [Qdrant 集成应用](./qdrant-app/)

**场景**: 文档检索系统

**技术栈**:
- Qdrant
- Gin
- go-client

**功能**:
- 文档向量化
- 语义搜索
- 推荐系统（Recommend API）
- 探索查询（Discover API）

**测试**:
- qdrant_client_test.go (4个测试)
- model_test.go (2个测试)
- JSON 生成验证

---

### 3. [RAG 检索应用](./rag-app/)

**场景**: RAG (Retrieval Augmented Generation)

**技术栈**:
- PostgreSQL + pgvector
- sqlx
- Gin
- LLM API

**功能**:
- 文档分块和向量化
- 混合检索
- 重排序
- LLM 集成

**测试**:
- repository_test.go (3个测试)
- rag_service_test.go (3个测试)
- 包含 Mock 服务示例

---

## 🚀 快速开始

### 运行示例

```bash
# 1. 选择示例
cd pgvector-app   # 或 qdrant-app, rag-app

# 2. 安装依赖
go mod tidy

# 3. 运行
go run *.go
```

### 运行测试

```bash
# 运行单元测试（不需要数据库）
go test -v

# 运行集成测试（需要数据库）
# PostgreSQL 示例需要：
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password ankane/pgvector

# Qdrant 示例需要：
docker run -d -p 6333:6333 qdrant/qdrant

# 然后运行测试
go test -v
```

---

## 📖 学习路径

### 如果您是...

#### 初学者 👶
1. 阅读 [pgvector-app](./pgvector-app/) - 最简单
2. 理解基础向量检索流程

#### 进阶开发者 🧑‍💻
1. 阅读 [qdrant-app](./qdrant-app/) - 中等难度
2. 学习 Qdrant 高级 API

#### 架构师 🏗️
1. 阅读 [rag-app](./rag-app/) - 完整 RAG 架构
2. 理解生产级 RAG 应用设计

---

## 🔗 相关文档

- [sqlxb README](../README.md)
- [Vector Database Quick Start](../doc/VECTOR_QUICKSTART.md)
- [Builder Best Practices](../doc/BUILDER_BEST_PRACTICES.md)
- [AI Application Docs](../doc/ai_application/README.md)

---

**版本**: v0.10.3  
**最后更新**: 2025-02-27

