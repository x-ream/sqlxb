package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化 Qdrant 客户端
	qdrant := NewQdrantClient("http://localhost:6333", "documents")

	// 创建 HTTP 服务
	r := gin.Default()

	// 注册路由
	api := r.Group("/api")
	{
		api.POST("/search", SearchHandler(qdrant))
		api.POST("/recommend", RecommendHandler(qdrant))
	}

	// 启动服务
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

