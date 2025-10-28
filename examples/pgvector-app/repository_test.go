package main

import (
	"testing"

	"github.com/fndome/xb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// 注意：这些测试需要实际的 PostgreSQL + pgvector 环境
// 可以通过 Docker 快速启动：
// docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password ankane/pgvector

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres://postgres:password@localhost/test_db?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
		return nil
	}

	// 创建测试表
	_, err = db.Exec(`
		CREATE EXTENSION IF NOT EXISTS vector;
		DROP TABLE IF EXISTS code_snippets;
		CREATE TABLE code_snippets (
			id BIGSERIAL PRIMARY KEY,
			file_path VARCHAR(500),
			language VARCHAR(50),
			content TEXT,
			embedding vector(768),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewCodeRepository(db)

	// 创建测试数据
	embedding := make(xb.Vector, 768)
	for i := range embedding {
		embedding[i] = 0.1
	}

	code := &CodeSnippet{
		FilePath:  "user_service.go",
		Language:  "golang",
		Content:   "func GetUser(id int64) (*User, error) { ... }",
		Embedding: embedding,
	}

	err := repo.Create(code)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if code.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	t.Logf("Created code snippet with ID: %d", code.ID)
}

func TestVectorSearch(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewCodeRepository(db)

	// 插入测试数据
	embedding := make(xb.Vector, 768)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001
	}

	code := &CodeSnippet{
		FilePath:  "test.go",
		Language:  "golang",
		Content:   "test content",
		Embedding: embedding,
	}
	repo.Create(code)

	// 测试向量搜索
	queryVector := make([]float32, 768)
	for i := range queryVector {
		queryVector[i] = float32(i) * 0.001
	}

	results, err := repo.VectorSearch(queryVector, 10)
	if err != nil {
		t.Fatalf("VectorSearch failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 result")
	}

	t.Logf("Found %d results", len(results))
}

func TestHybridSearch(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewCodeRepository(db)

	// 插入测试数据
	embedding := make(xb.Vector, 768)
	code := &CodeSnippet{
		FilePath:  "service.go",
		Language:  "golang",
		Content:   "user service implementation",
		Embedding: embedding,
	}
	repo.Create(code)

	// 测试混合搜索
	queryVector := make([]float32, 768)
	results, err := repo.HybridSearch(queryVector, "golang", 10)
	if err != nil {
		t.Fatalf("HybridSearch failed: %v", err)
	}

	// 验证过滤条件
	for _, r := range results {
		if r.Language != "golang" {
			t.Errorf("Expected language=golang, got %s", r.Language)
		}
	}

	t.Logf("Hybrid search found %d golang results", len(results))
}
