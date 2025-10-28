package main

import (
	"time"

	"github.com/fndome/xb"
)

// CodeSnippet 代码片段模型
type CodeSnippet struct {
	ID        int64     `json:"id" db:"id"`
	FilePath  string    `json:"file_path" db:"file_path"`
	Language  string    `json:"language" db:"language"`
	Content   string    `json:"content" db:"content"`
	Embedding xb.Vector `json:"embedding" db:"embedding"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (*CodeSnippet) TableName() string {
	return "code_snippets"
}

// CreateCodeRequest 创建请求
type CreateCodeRequest struct {
	FilePath  string    `json:"file_path" binding:"required"`
	Language  string    `json:"language" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	Embedding []float32 `json:"embedding" binding:"required"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	QueryVector []float32 `json:"query_vector" binding:"required"`
	Language    string    `json:"language"`
	Limit       *int      `json:"limit"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Results []*CodeSnippet `json:"results"`
	Total   int            `json:"total"`
}
