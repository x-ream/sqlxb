package main

import (
	"time"

	"github.com/fndome/xb"
)

// Document 文档模型
type Document struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	DocType   string    `json:"doc_type"`
	Language  string    `json:"language"`
	Embedding xb.Vector `json:"embedding"`
	CreatedAt time.Time `json:"created_at"`
}

func (*Document) TableName() string {
	return "documents"
}

// CreateDocRequest 创建文档请求
type CreateDocRequest struct {
	Title     string    `json:"title" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	DocType   string    `json:"doc_type"`
	Language  string    `json:"language"`
	Embedding []float32 `json:"embedding" binding:"required"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	QueryVector []float32 `json:"query_vector" binding:"required"`
	DocType     string    `json:"doc_type"`
	Language    string    `json:"language"`
	MinScore    *float64  `json:"min_score"`
	Limit       *int      `json:"limit"`
}

// RecommendRequest 推荐请求
type RecommendRequest struct {
	Positive []int64 `json:"positive" binding:"required"`
	Negative []int64 `json:"negative"`
	Limit    *int    `json:"limit"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results []*Document `json:"results"`
	Total   int         `json:"total"`
}
