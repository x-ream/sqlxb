# RAG 检索应用完整示例

这是一个使用 xb 构建的完整 RAG (Retrieval Augmented Generation) 应用，展示如何将文档检索与 LLM 结合。

## 📋 功能

- 文档分块和向量化
- 语义检索
- 混合检索（关键词 + 向量）
- 重排序和多样性
- LLM 集成

## 🏗️ 架构

```
用户查询 → 向量化 → xb 检索 → 重排序 → LLM 生成 → 回答
            ↓           ↓          ↓
         Embedding   PostgreSQL  Application
                     或 Qdrant    Layer
```

## 🚀 快速开始

### 1. 安装依赖

```bash
go get github.com/fndome/xb
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go get github.com/gin-gonic/gin
```

### 2. 创建数据库

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE document_chunks (
    id BIGSERIAL PRIMARY KEY,
    doc_id BIGINT,
    chunk_id INT,
    content TEXT,
    embedding vector(768),
    doc_type VARCHAR(50),
    language VARCHAR(10),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ON document_chunks USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX ON document_chunks (doc_type);
CREATE INDEX ON document_chunks (language);
```

### 3. 运行应用

```bash
cd examples/rag-app
go run *.go
```

### 4. 测试 API

```bash
# 上传文档
curl -X POST http://localhost:8080/api/documents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go语言并发编程",
    "content": "Goroutine和Channel是Go语言并发编程的核心...",
    "doc_type": "article",
    "language": "zh"
  }'

# RAG 查询
curl -X POST http://localhost:8080/api/rag/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "如何在Go中使用Channel？",
    "doc_type": "article",
    "top_k": 5
  }'
```

## 📁 项目结构

```
rag-app/
├── README.md
├── main.go            # 主程序
├── model.go           # 数据模型
├── repository.go      # 数据访问层
├── rag_service.go     # RAG 服务层
├── handler.go         # HTTP 处理器
└── go.mod
```

## 📚 相关文档

- [RAG Best Practices](../../doc/ai_application/RAG_BEST_PRACTICES.md)
- [Hybrid Search](../../doc/ai_application/HYBRID_SEARCH.md)
- [Vector Diversity](../../doc/VECTOR_DIVERSITY_QDRANT.md)

