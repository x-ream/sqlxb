package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateDocumentHandler 创建文档
func CreateDocumentHandler(service *RAGService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateDocRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 分块处理
		chunks := splitDocument(req.Content)

		// 向量化每个分块
		for i, chunkContent := range chunks {
			embedding, err := service.embedder.Embed(c.Request.Context(), chunkContent)
			if err != nil {
				log.Printf("Embedding failed: %v", err)
				continue
			}

			chunk := &DocumentChunk{
				ChunkID:   &i,
				Content:   chunkContent,
				Embedding: embedding,
				DocType:   req.DocType,
				Language:  req.Language,
				Metadata:  `{"title": "` + req.Title + `"}`,
			}

			if err := service.repo.Create(chunk); err != nil {
				log.Printf("Create chunk failed: %v", err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Document created",
			"chunks":  len(chunks),
		})
	}
}

// RAGQueryHandler RAG 查询处理器
func RAGQueryHandler(service *RAGService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RAGQueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 执行 RAG 查询
		resp, err := service.Query(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// splitDocument 简单的文档分块（实际应用中应使用更复杂的分块策略）
func splitDocument(content string) []string {
	// 按段落分块
	paragraphs := strings.Split(content, "\n\n")

	chunks := make([]string, 0)
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if len(p) > 50 { // 只保留有足够内容的段落
			chunks = append(chunks, p)
		}
	}

	return chunks
}
