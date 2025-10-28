package main

import (
	"testing"

	"github.com/fndome/xb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres://postgres:password@localhost/pageindex_test?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
		return nil
	}

	// 创建测试表
	_, err = db.Exec(`
		DROP TABLE IF EXISTS page_index_nodes;
		DROP TABLE IF EXISTS documents;
		
		CREATE TABLE documents (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(500),
			total_pages INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE page_index_nodes (
			id BIGSERIAL PRIMARY KEY,
			doc_id BIGINT REFERENCES documents(id),
			node_id VARCHAR(50),
			parent_id VARCHAR(50),
			title TEXT,
			start_page INT,
			end_page INT,
			summary TEXT,
			content TEXT,
			level INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return db
}

func TestCreateDocument(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewDocumentRepository(db)

	doc := &Document{
		Name:       "Annual Report 2024",
		TotalPages: xb.Int(100),
	}

	err := repo.CreateDocument(doc)
	if err != nil {
		t.Fatalf("CreateDocument failed: %v", err)
	}

	if doc.ID == 0 {
		t.Error("Expected doc ID to be set")
	}

	t.Logf("Created document with ID: %d", doc.ID)
}

func TestFindNodesByTitle(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewDocumentRepository(db)

	// 创建测试数据
	doc := &Document{Name: "Test Doc", TotalPages: xb.Int(50)}
	repo.CreateDocument(doc)

	node := &PageIndexNode{
		DocID:     &doc.ID,
		NodeID:    "0001",
		ParentID:  "",
		Title:     "Financial Stability",
		StartPage: xb.Int(10),
		EndPage:   xb.Int(20),
		Summary:   "Summary of financial stability",
		Level:     xb.Int(1),
	}
	repo.CreateNode(node)

	// 测试标题搜索
	results, err := repo.FindNodesByTitle(doc.ID, "Financial")
	if err != nil {
		t.Fatalf("FindNodesByTitle failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 result")
	}

	if results[0].Title != "Financial Stability" {
		t.Errorf("Expected title 'Financial Stability', got '%s'", results[0].Title)
	}

	t.Logf("Found %d nodes with title containing 'Financial'", len(results))
}

func TestFindNodesByPage(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewDocumentRepository(db)

	doc := &Document{Name: "Test Doc", TotalPages: xb.Int(50)}
	repo.CreateDocument(doc)

	// 创建节点：页码范围 10-20
	node := &PageIndexNode{
		DocID:     &doc.ID,
		NodeID:    "0001",
		Title:     "Chapter 1",
		StartPage: xb.Int(10),
		EndPage:   xb.Int(20),
		Level:     xb.Int(1),
	}
	repo.CreateNode(node)

	// 测试：查询第 15 页（应该找到）
	results, err := repo.FindNodesByPage(doc.ID, 15)
	if err != nil {
		t.Fatalf("FindNodesByPage failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to find node containing page 15")
	}

	// 测试：查询第 5 页（不应该找到）
	results, err = repo.FindNodesByPage(doc.ID, 5)
	if err != nil {
		t.Fatalf("FindNodesByPage failed: %v", err)
	}

	if len(results) != 0 {
		t.Error("Expected no nodes containing page 5")
	}

	t.Log("Page range query works correctly")
}

func TestFindChildNodes(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewDocumentRepository(db)

	doc := &Document{Name: "Test Doc", TotalPages: xb.Int(50)}
	repo.CreateDocument(doc)

	// 创建父节点
	parent := &PageIndexNode{
		DocID:  &doc.ID,
		NodeID: "0001",
		Title:  "Chapter 1",
		Level:  xb.Int(1),
	}
	repo.CreateNode(parent)

	// 创建子节点
	child1 := &PageIndexNode{
		DocID:    &doc.ID,
		NodeID:   "0002",
		ParentID: "0001",
		Title:    "Section 1.1",
		Level:    xb.Int(2),
	}
	repo.CreateNode(child1)

	child2 := &PageIndexNode{
		DocID:    &doc.ID,
		NodeID:   "0003",
		ParentID: "0001",
		Title:    "Section 1.2",
		Level:    xb.Int(2),
	}
	repo.CreateNode(child2)

	// 查询子节点
	children, err := repo.FindChildNodes(doc.ID, "0001")
	if err != nil {
		t.Fatalf("FindChildNodes failed: %v", err)
	}

	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}

	t.Logf("Found %d child nodes", len(children))
}

func TestAutoFiltering(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewDocumentRepository(db)

	doc := &Document{Name: "Test Doc", TotalPages: xb.Int(50)}
	repo.CreateDocument(doc)

	// 测试自动过滤（空字符串）
	var emptyKeyword string
	results, err := repo.FindNodesByTitle(doc.ID, emptyKeyword)
	if err != nil {
		t.Fatalf("FindNodesByTitle failed: %v", err)
	}

	// 应该成功执行，空字符串被自动过滤
	// 返回所有该文档的节点
	t.Logf("Auto-filtering works: empty keyword returned %d results", len(results))
}
