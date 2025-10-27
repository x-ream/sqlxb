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

		if req.Limit <= 0 {
			req.Limit = 10
		}

		docs, err := client.Search(
			req.QueryVector,
			req.DocType,
			req.Language,
			req.MinScore,
			req.Limit,
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

		if req.Limit <= 0 {
			req.Limit = 10
		}

		docs, err := client.Recommend(req.Positive, req.Negative, req.Limit)
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

