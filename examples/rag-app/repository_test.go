package main

import (
	"testing"

	"github.com/fndome/xb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// 注意：这些测试需要实际的 PostgreSQL + pgvector 环境
// docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password ankane/pgvector

func setupRAGTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres://postgres:password@localhost/rag_test?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
		return nil
	}

	_, err = db.Exec(`
		CREATE EXTENSION IF NOT EXISTS vector;
		DROP TABLE IF EXISTS document_chunks;
		CREATE TABLE document_chunks (
			id BIGSERIAL PRIMARY KEY,
			doc_id BIGINT,
			chunk_id INT,
			content TEXT,
			embedding vector(768),
			doc_type VARCHAR(50),
			language VARCHAR(10),
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX ON document_chunks USING ivfflat (embedding vector_cosine_ops);
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func TestCreateChunk(t *testing.T) {
	db := setupRAGTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewChunkRepository(db)

	// 创建测试分块
	embedding := make(xb.Vector, 768)
	for i := range embedding {
		embedding[i] = 0.1
	}

	chunk := &DocumentChunk{
		ChunkID:   xb.Int(0),
		Content:   "Go 语言并发编程...",
		Embedding: embedding,
		DocType:   "article",
		Language:  "zh",
		Metadata:  `{"title": "Go并发"}`,
	}

	err := repo.Create(chunk)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	t.Logf("Created chunk successfully")
}

func TestVectorSearchWithFiltering(t *testing.T) {
	db := setupRAGTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewChunkRepository(db)

	// 插入测试数据
	embedding := make(xb.Vector, 768)
	chunk := &DocumentChunk{
		ChunkID:   xb.Int(0),
		Content:   "test content",
		Embedding: embedding,
		DocType:   "article",
		Language:  "zh",
		Metadata:  `{}`,
	}
	repo.Create(chunk)

	// 测试向量搜索 + 过滤
	queryVector := make([]float32, 768)
	results, err := repo.VectorSearch(queryVector, "article", "zh", 10)
	if err != nil {
		t.Fatalf("VectorSearch failed: %v", err)
	}

	// 验证过滤条件
	for _, r := range results {
		if r.DocType != "article" {
			t.Errorf("Expected doc_type=article, got %s", r.DocType)
		}
		if r.Language != "zh" {
			t.Errorf("Expected language=zh, got %s", r.Language)
		}
	}

	t.Logf("VectorSearch with filtering: found %d results", len(results))
}

func TestHybridSearchAutoFiltering(t *testing.T) {
	db := setupRAGTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewChunkRepository(db)

	queryVector := make([]float32, 768)

	// 测试自动过滤（空字符串）
	var emptyKeyword string
	var emptyDocType string
	var emptyLanguage string

	results, err := repo.HybridSearch(queryVector, emptyKeyword, emptyDocType, emptyLanguage, 10)
	if err != nil {
		t.Fatalf("HybridSearch failed: %v", err)
	}

	// 应该成功执行，空字符串被自动过滤
	t.Logf("HybridSearch with auto-filtering: found %d results", len(results))
}
