package main

import (
	"time"

	"github.com/fndome/xb"
)

// DocumentChunk 文档分块
type DocumentChunk struct {
	ID        int64     `json:"id" db:"id"`
	DocID     *int64    `json:"doc_id" db:"doc_id"`
	ChunkID   *int      `json:"chunk_id" db:"chunk_id"`
	Content   string    `json:"content" db:"content"`
	Embedding xb.Vector `json:"embedding" db:"embedding"`
	DocType   string    `json:"doc_type" db:"doc_type"`
	Language  string    `json:"language" db:"language"`
	Metadata  string    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (*DocumentChunk) TableName() string {
	return "document_chunks"
}

// CreateDocRequest 创建文档请求
type CreateDocRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	DocType  string `json:"doc_type"`
	Language string `json:"language"`
}

// RAGQueryRequest RAG 查询请求
type RAGQueryRequest struct {
	Question string `json:"question" binding:"required"`
	DocType  string `json:"doc_type"`
	Language string `json:"language"`
	TopK     *int   `json:"top_k"`
}

// RAGQueryResponse RAG 查询响应
type RAGQueryResponse struct {
	Answer   string                 `json:"answer"`
	Sources  []*DocumentChunk       `json:"sources"`
	Metadata map[string]interface{} `json:"metadata"`
}
