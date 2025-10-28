package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 初始化数据库
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost/pageindex_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建服务
	repo := NewDocumentRepository(db)
	importer := NewPageIndexImporter(repo)

	// 创建 HTTP 服务
	r := gin.Default()

	// 注册路由
	api := r.Group("/api")
	{
		// 导入
		api.POST("/import", ImportHandler(importer))

		// 搜索
		api.GET("/search/title", SearchByTitleHandler(repo))
		api.GET("/search/page", SearchByPageHandler(repo))
		api.GET("/search/level", SearchByLevelHandler(repo))

		// 节点详情
		api.GET("/nodes/:node_id/children", GetNodeWithChildrenHandler(importer))
	}

	// 启动服务
	log.Println("PageIndex Server starting on :8080")
	log.Println("Endpoints:")
	log.Println("  POST /api/import - 导入 PageIndex JSON")
	log.Println("  GET  /api/search/title?doc_id=1&keyword=xxx - 标题搜索")
	log.Println("  GET  /api/search/page?doc_id=1&page=25 - 页码搜索")
	log.Println("  GET  /api/search/level?doc_id=1&level=2 - 层级搜索")
	log.Println("  GET  /api/nodes/:node_id/children?doc_id=1 - 节点详情")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

