package main

import (
	"time"
)

// Document 文档
type Document struct {
	ID         int64     `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	TotalPages *int      `json:"total_pages" db:"total_pages"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (*Document) TableName() string {
	return "documents"
}

// PageIndexNode PageIndex 节点（扁平化存储）
type PageIndexNode struct {
	ID        int64     `json:"id" db:"id"`
	DocID     *int64    `json:"doc_id" db:"doc_id"`
	NodeID    string    `json:"node_id" db:"node_id"`
	ParentID  string    `json:"parent_id" db:"parent_id"`
	Title     string    `json:"title" db:"title"`
	StartPage *int      `json:"start_page" db:"start_page"`
	EndPage   *int      `json:"end_page" db:"end_page"`
	Summary   string    `json:"summary" db:"summary"`
	Content   string    `json:"content" db:"content"`
	Level     *int      `json:"level" db:"level"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (*PageIndexNode) TableName() string {
	return "page_index_nodes"
}

// PageIndexJSON PageIndex 生成的原始 JSON 结构
type PageIndexJSON struct {
	Title      string          `json:"title"`
	NodeID     string          `json:"node_id"`
	StartIndex *int            `json:"start_index"`
	EndIndex   *int            `json:"end_index"`
	Summary    string          `json:"summary"`
	Nodes      []PageIndexJSON `json:"nodes"`
}

// ImportRequest 导入请求
type ImportRequest struct {
	DocumentName string        `json:"document_name" binding:"required"`
	TotalPages   *int          `json:"total_pages"`
	Structure    PageIndexJSON `json:"structure" binding:"required"`
}

// SearchByTitleRequest 标题搜索请求
type SearchByTitleRequest struct {
	DocID   *int64 `json:"doc_id" binding:"required"`
	Keyword string `json:"keyword" binding:"required"`
}

// SearchByPageRequest 页码搜索请求
type SearchByPageRequest struct {
	DocID *int64 `json:"doc_id" binding:"required"`
	Page  *int   `json:"page" binding:"required"`
}

// NodeResponse 节点响应
type NodeResponse struct {
	Node     *PageIndexNode   `json:"node"`
	Children []*PageIndexNode `json:"children,omitempty"`
}
