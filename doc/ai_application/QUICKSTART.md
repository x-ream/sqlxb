# AI 应用快速入门

## 📋 5 分钟快速体验

本指南帮助您在 5 分钟内搭建一个简单的 RAG 应用。

## 🚀 前置条件

- Go 1.21+
- Docker (用于运行 Qdrant)
- OpenAI API Key

## 📦 第一步：启动 Qdrant

```bash
docker run -d \
  --name qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  qdrant/qdrant:latest
```

验证 Qdrant 是否运行：
```bash
curl http://localhost:6333/health
# 应该返回: {"title":"qdrant - vector search engine","version":"x.x.x"}
```

## 💻 第二步：编写 RAG 应用

创建 `main.go`:

```go
package main

import (
    "fmt"
    "github.com/x-ream/xb"
)

// 1. 定义文档结构
type Document struct {
    ID        int64     `json:"id"`
    Content   string    `json:"content"`
    Embedding []float32 `json:"embedding"`
}

// 2. 模拟 Embedding 函数（实际应调用 OpenAI API）
func mockEmbed(text string) []float32 {
    // 返回 1536 维的随机向量（OpenAI text-embedding-ada-002 的维度）
    vec := make([]float32, 1536)
    for i := range vec {
        vec[i] = float32(len(text)) / 1000.0 // 简化演示
    }
    return vec
}

func main() {
    // 3. 准备文档
    documents := []string{
        "sqlxb 是一个 AI-First 的 Go ORM 库",
        "sqlxb 支持 PostgreSQL 和 Qdrant",
        "sqlxb 提供类型安全的查询构建器",
    }

    fmt.Println("=== 索引文档 ===")
    
    // 4. 索引文档（这里只是生成查询，实际需要执行）
    for i, doc := range documents {
        embedding := mockEmbed(doc)
        
        // 注意：sqlxb 主要用于查询，插入建议直接用 SQL 或 ORM
        // 这里展示如何准备数据
        docData := Document{
            ID:        int64(i + 1),
            Content:   doc,
            Embedding: embedding,
        }
        
        fmt.Printf("文档 %d: %s\n", i+1, doc)
        fmt.Printf("Embedding 维度: %d\n\n", len(embedding))
        
        // 实际插入到 Qdrant 的逻辑在这里
        _ = docData
    }

    // 5. 查询示例
    query := "sqlxb 支持什么数据库？"
    queryVector := mockEmbed(query)

    fmt.Println("=== 执行向量检索 ===")
    fmt.Printf("查询: %s\n\n", query)

    // 6. 构建向量检索查询
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", queryVector, 5).
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(0.7)
        }).
        Build()

    qdrantJSON, _ := built.ToQdrantJSON()

    fmt.Println("Qdrant 查询:")
    fmt.Printf("%s\n", qdrantJSON)
}
```

运行：
```bash
go mod init demo
go get github.com/x-ream/xb
go run main.go
```

## 🎯 第三步：完整 RAG 应用

创建 `rag.go`，集成真实的 OpenAI Embedding:

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    openai "github.com/sashabaranov/go-openai"
    "github.com/x-ream/xb"
)

func main() {
    // 1. 初始化 OpenAI 客户端
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        fmt.Println("请设置 OPENAI_API_KEY 环境变量")
        return
    }
    
    client := openai.NewClient(apiKey)
    
    // 2. 文档列表
    documents := []string{
        "sqlxb 是一个现代化的 Go ORM 库，专为 AI 应用设计",
        "sqlxb 支持 PostgreSQL 和 Qdrant 两种数据库后端",
        "sqlxb 提供类型安全的查询构建，避免 SQL 注入",
        "sqlxb 的向量检索功能支持相似度搜索和混合查询",
    }

    // 3. 生成 Embeddings
    fmt.Println("正在生成文档 Embeddings...")
    embeddings, err := generateEmbeddings(client, documents)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    // 4. 用户查询
    query := "sqlxb 支持哪些数据库？"
    fmt.Printf("\n用户查询: %s\n", query)

    // 5. 生成查询向量
    queryEmbedding, err := generateEmbedding(client, query)
    if err != nil {
        fmt.Printf("错误: %v\n", err)
        return
    }

    // 6. 构建向量检索查询
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", queryEmbedding, 3).
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(0.75)
        }).
        Build()

    qdrantJSON, _ := built.ToQdrantJSON()

    fmt.Println("\n生成的 Qdrant 查询:")
    fmt.Printf("%s\n", qdrantJSON)

    // 7. 模拟检索结果
    fmt.Println("\n最相关的文档:")
    for i, doc := range documents[:3] {
        fmt.Printf("%d. %s\n", i+1, doc)
    }
}

func generateEmbedding(client *openai.Client, text string) ([]float32, error) {
    resp, err := client.CreateEmbeddings(
        context.Background(),
        openai.EmbeddingRequest{
            Input: []string{text},
            Model: openai.AdaEmbeddingV2,
        },
    )
    if err != nil {
        return nil, err
    }

    return resp.Data[0].Embedding, nil
}

func generateEmbeddings(client *openai.Client, texts []string) ([][]float32, error) {
    resp, err := client.CreateEmbeddings(
        context.Background(),
        openai.EmbeddingRequest{
            Input: texts,
            Model: openai.AdaEmbeddingV2,
        },
    )
    if err != nil {
        return nil, err
    }

    embeddings := make([][]float32, len(resp.Data))
    for i, data := range resp.Data {
        embeddings[i] = data.Embedding
    }

    return embeddings, nil
}
```

运行：
```bash
export OPENAI_API_KEY="your-api-key"
go get github.com/sashabaranov/go-openai
go run rag.go
```

## 📊 预期输出

```
正在生成文档 Embeddings...

用户查询: sqlxb 支持哪些数据库？

生成的 Qdrant 查询:
map[collection_name:documents vector:[0.01 0.02 ...] limit:3 score_threshold:0.75]

最相关的文档:
1. sqlxb 支持 PostgreSQL 和 Qdrant 两种数据库后端
2. sqlxb 是一个现代化的 Go ORM 库，专为 AI 应用设计
3. sqlxb 的向量检索功能支持相似度搜索和混合查询
```

## 🎓 下一步

现在您已经掌握了基础用法，可以继续学习：

1. **[RAG_BEST_PRACTICES.md](./RAG_BEST_PRACTICES.md)** - 学习生产级 RAG 应用的最佳实践
2. **[AGENT_TOOLKIT.md](./AGENT_TOOLKIT.md)** - 将 sqlxb 集成到 AI Agent 系统
3. **[LANGCHAIN_INTEGRATION.md](./LANGCHAIN_INTEGRATION.md)** - Python LangChain 集成
4. **[HYBRID_SEARCH.md](./HYBRID_SEARCH.md)** - 混合检索策略
5. **[PERFORMANCE.md](./PERFORMANCE.md)** - 性能优化指南

## ❓ 常见问题

### Q: 如何连接真实的 Qdrant？

```go
import qdrant "github.com/qdrant/go-client/qdrant"

client, err := qdrant.NewClient(&qdrant.Config{
    Host: "localhost",
    Port: 6334,
})
```

### Q: Embedding 向量维度不匹配怎么办？

确保您的 Qdrant Collection 创建时使用了正确的维度：

```bash
curl -X PUT 'http://localhost:6333/collections/documents' \
  -H 'Content-Type: application/json' \
  -d '{
    "vectors": {
      "size": 1536,
      "distance": "Cosine"
    }
  }'
```

### Q: 如何批量索引大量文档？

```go
// 分批处理，每批 100 个文档
batchSize := 100
for i := 0; i < len(documents); i += batchSize {
    end := min(i+batchSize, len(documents))
    batch := documents[i:end]
    
    // 批量生成 embeddings
    embeddings, _ := generateEmbeddings(client, batch)
    
    // 批量插入
    // ... 插入逻辑
}
```

## 🎉 完成！

恭喜！您已经完成了 sqlxb AI 应用的快速入门。

如有问题，请查看：
- [FAQ.md](./FAQ.md) - 常见问题
- [GitHub Issues](https://github.com/x-ream/xb/issues)
- [GitHub Discussions](https://github.com/x-ream/xb/discussions)

---

**开始构建您的 AI 应用吧！** 🚀

