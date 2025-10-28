# PostgreSQL + pgvector 完整应用示例

这是一个使用 xb + PostgreSQL + pgvector 构建的完整代码搜索应用。

## 📋 功能

- 代码向量化存储
- 语义搜索
- 混合检索（关键词 + 向量）
- 分页查询

## 🚀 快速开始

### 1. 安装依赖

```bash
go get github.com/fndome/xb
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
```

### 2. 创建数据库

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE code_snippets (
    id BIGSERIAL PRIMARY KEY,
    file_path VARCHAR(500),
    language VARCHAR(50),
    content TEXT,
    embedding vector(768),  -- OpenAI ada-002 维度
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ON code_snippets USING ivfflat (embedding vector_cosine_ops);
```

### 3. 运行应用

```bash
cd examples/pgvector-app
go run main.go
```

### 4. 测试 API

```bash
# 插入代码片段
curl -X POST http://localhost:8080/api/code \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "user_service.go",
    "language": "golang",
    "content": "func GetUser(id int64) (*User, error) { ... }",
    "embedding": [0.1, 0.2, ..., 0.768]
  }'

# 搜索相似代码
curl "http://localhost:8080/api/search?query=user%20service&limit=10"

# 混合搜索
curl "http://localhost:8080/api/hybrid-search" \
  -H "Content-Type: application/json" \
  -d '{
    "query_vector": [0.1, 0.2, ..., 0.768],
    "language": "golang",
    "limit": 10
  }'
```

## 📁 项目结构

```
pgvector-app/
├── README.md
├── main.go            # 主程序
├── model.go           # 数据模型
├── repository.go      # 数据访问层
├── handler.go         # HTTP 处理器
└── go.mod
```

## 🔍 核心代码

见同目录下的 Go 文件。

## 📚 相关文档

- [xb README](../../README.md)
- [Vector Database Quick Start](../../doc/VECTOR_QUICKSTART.md)
- [Builder Best Practices](../../doc/BUILDER_BEST_PRACTICES.md)

