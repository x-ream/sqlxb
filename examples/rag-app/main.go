package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 初始化数据库
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost/rag_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建服务
	repo := NewChunkRepository(db)
	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}
	ragService := NewRAGService(repo, embedder, llm)

	// 创建 HTTP 服务
	r := gin.Default()

	// 注册路由
	api := r.Group("/api")
	{
		api.POST("/documents", CreateDocumentHandler(ragService))
		api.POST("/rag/query", RAGQueryHandler(ragService))
	}

	// 启动服务
	log.Println("RAG Server starting on :8080")
	log.Println("Endpoints:")
	log.Println("  POST /api/documents - 上传文档")
	log.Println("  POST /api/rag/query - RAG 查询")
	
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

