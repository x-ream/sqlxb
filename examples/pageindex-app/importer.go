package main

import (
	"fmt"
)

// PageIndexImporter PageIndex JSON 导入器
type PageIndexImporter struct {
	repo *DocumentRepository
}

func NewPageIndexImporter(repo *DocumentRepository) *PageIndexImporter {
	return &PageIndexImporter{repo: repo}
}

// Import 导入 PageIndex 生成的 JSON 结构
func (imp *PageIndexImporter) Import(req ImportRequest) error {
	// 1. 创建文档记录
	doc := &Document{
		Name:       req.DocumentName,
		TotalPages: req.TotalPages,
	}

	if err := imp.repo.CreateDocument(doc); err != nil {
		return fmt.Errorf("create document failed: %w", err)
	}

	// 2. 递归导入节点
	return imp.importNode(doc.ID, req.Structure, "", 0)
}

// importNode 递归导入节点
func (imp *PageIndexImporter) importNode(docID int64, jsonNode PageIndexJSON, parentID string, level int) error {
	// 创建当前节点
	node := &PageIndexNode{
		DocID:     &docID,
		NodeID:    jsonNode.NodeID,
		ParentID:  parentID,
		Title:     jsonNode.Title,
		StartPage: jsonNode.StartIndex,
		EndPage:   jsonNode.EndIndex,
		Summary:   jsonNode.Summary,
		Level:     &level,
	}

	if err := imp.repo.CreateNode(node); err != nil {
		return fmt.Errorf("create node %s failed: %w", jsonNode.NodeID, err)
	}

	// 递归处理子节点
	for _, childJSON := range jsonNode.Nodes {
		if err := imp.importNode(docID, childJSON, jsonNode.NodeID, level+1); err != nil {
			return err
		}
	}

	return nil
}

// BuildHierarchy 构建层级结构（用于响应）
func (imp *PageIndexImporter) BuildHierarchy(docID int64, nodeID string) (*NodeResponse, error) {
	// 获取当前节点
	node, err := imp.repo.FindNodeByID(docID, nodeID)
	if err != nil {
		return nil, err
	}

	// 获取子节点
	children, err := imp.repo.FindChildNodes(docID, nodeID)
	if err != nil {
		return nil, err
	}

	return &NodeResponse{
		Node:     node,
		Children: children,
	}, nil
}

