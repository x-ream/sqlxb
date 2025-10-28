package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ImportHandler 导入 PageIndex JSON
func ImportHandler(importer *PageIndexImporter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ImportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := importer.Import(req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Import successful"})
	}
}

// SearchByTitleHandler 按标题搜索
func SearchByTitleHandler(repo *DocumentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		docID, _ := strconv.ParseInt(c.Query("doc_id"), 10, 64)
		keyword := c.Query("keyword")

		nodes, err := repo.FindNodesByTitle(docID, keyword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": nodes,
			"total":   len(nodes),
		})
	}
}

// SearchByPageHandler 按页码搜索
func SearchByPageHandler(repo *DocumentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		docID, _ := strconv.ParseInt(c.Query("doc_id"), 10, 64)
		page, _ := strconv.Atoi(c.Query("page"))

		nodes, err := repo.FindNodesByPage(docID, page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": nodes,
			"total":   len(nodes),
		})
	}
}

// SearchByLevelHandler 按层级搜索
func SearchByLevelHandler(repo *DocumentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		docID, _ := strconv.ParseInt(c.Query("doc_id"), 10, 64)
		level, _ := strconv.Atoi(c.Query("level"))

		nodes, err := repo.FindNodesByLevel(docID, level)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": nodes,
			"total":   len(nodes),
		})
	}
}

// GetNodeWithChildrenHandler 获取节点及其子节点
func GetNodeWithChildrenHandler(importer *PageIndexImporter) gin.HandlerFunc {
	return func(c *gin.Context) {
		docID, _ := strconv.ParseInt(c.Query("doc_id"), 10, 64)
		nodeID := c.Param("node_id")

		resp, err := importer.BuildHierarchy(docID, nodeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

