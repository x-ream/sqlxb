package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateCodeHandler 创建代码片段
func CreateCodeHandler(repo *CodeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		code := &CodeSnippet{
			FilePath:  req.FilePath,
			Language:  req.Language,
			Content:   req.Content,
			Embedding: req.Embedding,
		}

		if err := repo.Create(code); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, code)
	}
}

// GetCodeHandler 获取代码片段
func GetCodeHandler(repo *CodeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		code, err := repo.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		c.JSON(http.StatusOK, code)
	}
}

// SearchHandler 关键词搜索
func SearchHandler(repo *CodeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyword := c.Query("query")
		language := c.Query("language")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		rows, _ := strconv.Atoi(c.DefaultQuery("rows", "10"))

		codes, total, err := repo.KeywordSearch(keyword, language, page, rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": codes,
			"total":   total,
			"page":    page,
			"rows":    rows,
		})
	}
}

// HybridSearchHandler 混合搜索
func HybridSearchHandler(repo *CodeRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 设置默认值
		limit := 10
		if req.Limit != nil && *req.Limit > 0 {
			limit = *req.Limit
		}

		codes, err := repo.HybridSearch(req.QueryVector, req.Language, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": codes,
			"total":   len(codes),
		})
	}
}
