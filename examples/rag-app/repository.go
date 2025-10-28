package main

import (
	"github.com/fndome/xb"
	"github.com/jmoiron/sqlx"
)

// ChunkRepository 文档分块仓库
type ChunkRepository struct {
	db *sqlx.DB
}

func NewChunkRepository(db *sqlx.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

// Create 创建文档分块
func (r *ChunkRepository) Create(chunk *DocumentChunk) error {
	sql, args := xb.Of(&DocumentChunk{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("doc_id", chunk.DocID).
				Set("chunk_id", chunk.ChunkID).
				Set("content", chunk.Content).
				Set("embedding", chunk.Embedding).
				Set("doc_type", chunk.DocType).
				Set("language", chunk.Language).
				Set("metadata", chunk.Metadata)
		}).
		Build().
		SqlOfInsert()

	_, err := r.db.Exec(sql, args...)
	return err
}

// VectorSearch 向量搜索
func (r *ChunkRepository) VectorSearch(queryVector []float32, docType, language string, limit int) ([]*DocumentChunk, error) {
	sql, args := xb.Of(&DocumentChunk{}).
		VectorSearch("embedding", queryVector, limit).
		Eq("doc_type", docType).
		Eq("language", language).
		Build().
		SqlOfVectorSearch()

	var chunks []*DocumentChunk
	err := r.db.Select(&chunks, sql, args...)
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// HybridSearch 混合搜索（关键词 + 向量）
func (r *ChunkRepository) HybridSearch(queryVector []float32, keyword, docType, language string, limit int) ([]*DocumentChunk, error) {
	// 先做向量检索（over-fetch）
	vectorLimit := limit * 3

	sql, args := xb.Of(&DocumentChunk{}).
		VectorSearch("embedding", queryVector, vectorLimit).
		Like("content", keyword).
		Eq("doc_type", docType).
		Eq("language", language).
		Build().
		SqlOfVectorSearch()

	var chunks []*DocumentChunk
	err := r.db.Select(&chunks, sql, args...)
	if err != nil {
		return nil, err
	}

	// 在应用层做重排序（BM25 + Vector Score）
	// 这里简化处理，实际应用中可以使用更复杂的重排序算法
	if len(chunks) > limit {
		chunks = chunks[:limit]
	}

	return chunks, nil
}
