package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchHandler 搜索处理器
func SearchHandler(client *QdrantClient) gin.HandlerFunc {
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

		minScore := 0.0
		if req.MinScore != nil {
			minScore = *req.MinScore
		}

		docs, err := client.Search(
			req.QueryVector,
			req.DocType,
			req.Language,
			minScore,
			limit,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": docs,
			"total":   len(docs),
		})
	}
}

// RecommendHandler 推荐处理器
func RecommendHandler(client *QdrantClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RecommendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 设置默认值
		limit := 10
		if req.Limit != nil && *req.Limit > 0 {
			limit = *req.Limit
		}

		docs, err := client.Recommend(req.Positive, req.Negative, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": docs,
			"total":   len(docs),
		})
	}
}
