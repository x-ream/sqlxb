# Qdrant 集成完整应用示例

这是一个使用 xb + Qdrant 构建的完整文档检索应用。

## 📋 功能

- 文档向量化存储
- 语义搜索
- 推荐系统（Recommend API）
- 探索查询（Discover API）
- 高级过滤

## 🚀 快速开始

### 1. 安装依赖

```bash
go get github.com/fndome/xb
go get github.com/qdrant/go-client
go get github.com/gin-gonic/gin
```

### 2. 启动 Qdrant

```bash
docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

### 3. 运行应用

```bash
cd examples/qdrant-app
go run *.go
```

### 4. 测试 API

```bash
# 插入文档
curl -X POST http://localhost:8080/api/document \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go并发编程",
    "content": "goroutine和channel的使用...",
    "doc_type": "article",
    "embedding": [0.1, 0.2, ..., 0.768]
  }'

# 向量搜索
curl "http://localhost:8080/api/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query_vector": [0.1, 0.2, ..., 0.768],
    "doc_type": "article",
    "limit": 10
  }'

# 推荐查询
curl "http://localhost:8080/api/recommend" \
  -H "Content-Type: application/json" \
  -d '{
    "positive": [123, 456],
    "negative": [789],
    "limit": 10
  }'
```

## 📁 项目结构

```
qdrant-app/
├── README.md
├── main.go            # 主程序
├── model.go           # 数据模型
├── qdrant_client.go   # Qdrant 客户端
├── handler.go         # HTTP 处理器
└── go.mod
```

## 📚 相关文档

- [QdrantX Usage](../../doc/QDRANT_X_USAGE.md)
- [Qdrant Advanced API](../../doc/QDRANT_ADVANCED_API.md)
- [Builder Best Practices](../../doc/BUILDER_BEST_PRACTICES.md)

