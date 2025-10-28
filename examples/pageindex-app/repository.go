package main

import (
	"github.com/fndome/xb"
	"github.com/jmoiron/sqlx"
)

// DocumentRepository 文档仓库
type DocumentRepository struct {
	db *sqlx.DB
}

func NewDocumentRepository(db *sqlx.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// CreateDocument 创建文档
func (r *DocumentRepository) CreateDocument(doc *Document) error {
	sql, args := xb.Of(&Document{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("name", doc.Name).
				Set("total_pages", doc.TotalPages)
		}).
		Build().
		SqlOfInsert()

	result, err := r.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	doc.ID = id
	return nil
}

// CreateNode 创建节点
func (r *DocumentRepository) CreateNode(node *PageIndexNode) error {
	sql, args := xb.Of(&PageIndexNode{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("doc_id", node.DocID).
				Set("node_id", node.NodeID).
				Set("parent_id", node.ParentID).
				Set("title", node.Title).
				Set("start_page", node.StartPage).
				Set("end_page", node.EndPage).
				Set("summary", node.Summary).
				Set("content", node.Content).
				Set("level", node.Level)
		}).
		Build().
		SqlOfInsert()

	_, err := r.db.Exec(sql, args...)
	return err
}

// FindNodesByTitle 按标题搜索节点
func (r *DocumentRepository) FindNodesByTitle(docID int64, keyword string) ([]*PageIndexNode, error) {
	sql, args, _ := xb.Of(&PageIndexNode{}).
		Eq("doc_id", docID).
		Like("title", keyword). // 自动添加 %
		Sort("level", xb.ASC).
		Sort("start_page", xb.ASC).
		Build().
		SqlOfSelect()

	var nodes []*PageIndexNode
	err := r.db.Select(&nodes, sql, args...)
	return nodes, err
}

// FindNodesByPage 查询包含某页的节点
func (r *DocumentRepository) FindNodesByPage(docID int64, page int) ([]*PageIndexNode, error) {
	sql, args, _ := xb.Of(&PageIndexNode{}).
		Eq("doc_id", docID).
		Lte("start_page", page). // start_page <= page
		Gte("end_page", page).   // end_page >= page
		Sort("level", xb.ASC).
		Build().
		SqlOfSelect()

	var nodes []*PageIndexNode
	err := r.db.Select(&nodes, sql, args...)
	return nodes, err
}

// FindNodesByLevel 按层级查询节点
func (r *DocumentRepository) FindNodesByLevel(docID int64, level int) ([]*PageIndexNode, error) {
	sql, args, _ := xb.Of(&PageIndexNode{}).
		Eq("doc_id", docID).
		Eq("level", level).
		Sort("node_id", xb.ASC).
		Build().
		SqlOfSelect()

	var nodes []*PageIndexNode
	err := r.db.Select(&nodes, sql, args...)
	return nodes, err
}

// FindChildNodes 查询子节点
func (r *DocumentRepository) FindChildNodes(docID int64, parentNodeID string) ([]*PageIndexNode, error) {
	sql, args, _ := xb.Of(&PageIndexNode{}).
		Eq("doc_id", docID).
		Eq("parent_id", parentNodeID).
		Sort("start_page", xb.ASC).
		Build().
		SqlOfSelect()

	var nodes []*PageIndexNode
	err := r.db.Select(&nodes, sql, args...)
	return nodes, err
}

// FindNodeByID 根据 node_id 查询节点
func (r *DocumentRepository) FindNodeByID(docID int64, nodeID string) (*PageIndexNode, error) {
	sql, args, _ := xb.Of(&PageIndexNode{}).
		Eq("doc_id", docID).
		Eq("node_id", nodeID).
		Build().
		SqlOfSelect()

	var node PageIndexNode
	err := r.db.Get(&node, sql, args...)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// FindPageRange 查询页码范围内的节点
func (r *DocumentRepository) FindPageRange(docID int64, startPage, endPage int) ([]*PageIndexNode, error) {
	sql, args, _ := xb.Of(&PageIndexNode{}).
		Eq("doc_id", docID).
		Gte("start_page", startPage).
		Lte("end_page", endPage).
		Sort("start_page", xb.ASC).
		Build().
		SqlOfSelect()

	var nodes []*PageIndexNode
	err := r.db.Select(&nodes, sql, args...)
	return nodes, err
}
