package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 初始化数据库
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost/codebase?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建 Repository
	repo := NewCodeRepository(db)

	// 创建 HTTP 服务
	r := gin.Default()

	// 注册路由
	api := r.Group("/api")
	{
		api.POST("/code", CreateCodeHandler(repo))
		api.GET("/search", SearchHandler(repo))
		api.POST("/hybrid-search", HybridSearchHandler(repo))
		api.GET("/code/:id", GetCodeHandler(repo))
	}

	// 启动服务
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
