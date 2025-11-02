// Copyright 2025 me.fndo.xb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xb

import (
	"encoding/json"
	"testing"
)

// ============================================================================
// Qdrant INSERT 测试
// ============================================================================

// TestQdrantInsert_SinglePoint 测试插入单个点
func TestQdrantInsert_SinglePoint(t *testing.T) {
	point := map[string]interface{}{
		"id":     123,
		"vector": []float32{0.1, 0.2, 0.3, 0.4},
		"payload": map[string]interface{}{
			"language": "golang",
			"content":  "func main() {...}",
		},
	}

	builder := X().Custom(NewQdrantCustom())
	builder.inserts = &[]Bb{
		{Value: point},
	}
	built := builder.Build()

	jsonStr, err := built.JsonOfInsert()
	if err != nil {
		t.Fatalf("JsonOfInsert failed: %v", err)
	}

	t.Logf("=== Qdrant Insert (单点) ===\n%s", jsonStr)

	// 验证 JSON
	var req struct {
		Points []QdrantPoint `json:"points"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point, got %d", len(req.Points))
	}

	// JSON unmarshal 会把数字变成 float64
	if int(req.Points[0].ID.(float64)) != 123 {
		t.Errorf("Expected ID 123, got %v", req.Points[0].ID)
	}

	t.Logf("✅ Qdrant Insert JSON 生成成功")
}

// TestQdrantInsert_MultiplePoints 测试插入多个点
func TestQdrantInsert_MultiplePoints(t *testing.T) {
	points := []map[string]interface{}{
		{
			"id":     1,
			"vector": []float32{0.1, 0.2, 0.3},
			"payload": map[string]interface{}{
				"language": "golang",
			},
		},
		{
			"id":     2,
			"vector": []float32{0.4, 0.5, 0.6},
			"payload": map[string]interface{}{
				"language": "python",
			},
		},
		{
			"id":     3,
			"vector": []float32{0.7, 0.8, 0.9},
			"payload": map[string]interface{}{
				"language": "rust",
			},
		},
	}

	builder := X().Custom(NewQdrantCustom())
	bbs := []Bb{}
	for _, p := range points {
		bbs = append(bbs, Bb{Value: p})
	}
	builder.inserts = &bbs
	built := builder.Build()

	jsonStr, err := built.JsonOfInsert()
	if err != nil {
		t.Fatalf("JsonOfInsert failed: %v", err)
	}

	t.Logf("=== Qdrant Insert (多点) ===\n%s", jsonStr)

	// 验证 JSON
	var req struct {
		Points []QdrantPoint `json:"points"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 3 {
		t.Errorf("Expected 3 points, got %d", len(req.Points))
	}

	t.Logf("✅ 批量插入 %d 个点", len(req.Points))
}

// ============================================================================
// Qdrant UPDATE 测试
// ============================================================================

// TestQdrantUpdate_ByID 测试根据 ID 更新
func TestQdrantUpdate_ByID(t *testing.T) {
	builder := X().
		Custom(NewQdrantCustom()).
		Eq("id", 123)

	builder.updates = &[]Bb{
		{Key: "language", Value: "rust"},
		{Key: "version", Value: "1.75"},
	}

	built := builder.Build()

	jsonStr, err := built.JsonOfUpdate()
	if err != nil {
		t.Fatalf("JsonOfUpdate failed: %v", err)
	}

	t.Logf("=== Qdrant Update (By ID) ===\n%s", jsonStr)

	// 验证 JSON
	var req struct {
		Points  []interface{}          `json:"points,omitempty"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point ID, got %d", len(req.Points))
	}

	if int(req.Points[0].(float64)) != 123 {
		t.Errorf("Expected point ID 123, got %v", req.Points[0])
	}

	if req.Payload["language"] != "rust" {
		t.Errorf("Expected language=rust, got %v", req.Payload["language"])
	}

	t.Logf("✅ Qdrant Update (By ID) 成功")
}

// TestQdrantUpdate_ByFilter 测试根据过滤器更新
func TestQdrantUpdate_ByFilter(t *testing.T) {
	builder := X().
		Custom(NewQdrantCustom()).
		Eq("language", "golang").
		Gt("quality_score", 0.8)

	builder.updates = &[]Bb{
		{Key: "verified", Value: true},
	}

	built := builder.Build()

	jsonStr, err := built.JsonOfUpdate()
	if err != nil {
		t.Fatalf("JsonOfUpdate failed: %v", err)
	}

	t.Logf("=== Qdrant Update (By Filter) ===\n%s", jsonStr)

	// 验证 JSON
	var req struct {
		Filter  *QdrantFilter          `json:"filter,omitempty"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	}

	if len(req.Filter.Must) < 2 {
		t.Errorf("Expected at least 2 filter conditions, got %d", len(req.Filter.Must))
	}

	t.Logf("✅ Qdrant Update (By Filter) 成功，条件数：%d", len(req.Filter.Must))
}

// ============================================================================
// Qdrant DELETE 测试
// ============================================================================

// TestQdrantDelete_ByID 测试根据 ID 删除
func TestQdrantDelete_ByID(t *testing.T) {
	builder := X().
		Custom(NewQdrantCustom()).
		Eq("id", 456)

	built := builder.Build()
	built.Delete = true

	jsonStr, err := built.JsonOfDelete()
	if err != nil {
		t.Fatalf("JsonOfDelete failed: %v", err)
	}

	t.Logf("=== Qdrant Delete (By ID) ===\n%s", jsonStr)

	// 验证 JSON
	var req struct {
		Points []interface{} `json:"points,omitempty"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point ID, got %d", len(req.Points))
	}

	if int(req.Points[0].(float64)) != 456 {
		t.Errorf("Expected point ID 456, got %v", req.Points[0])
	}

	t.Logf("✅ Qdrant Delete (By ID) 成功")
}

// TestQdrantDelete_ByFilter 测试根据过滤器删除
func TestQdrantDelete_ByFilter(t *testing.T) {
	builder := X().
		Custom(NewQdrantCustom()).
		Eq("language", "outdated").
		Lt("quality_score", 0.3)

	built := builder.Build()
	built.Delete = true

	jsonStr, err := built.JsonOfDelete()
	if err != nil {
		t.Fatalf("JsonOfDelete failed: %v", err)
	}

	t.Logf("=== Qdrant Delete (By Filter) ===\n%s", jsonStr)

	// 验证 JSON
	var req struct {
		Filter *QdrantFilter `json:"filter,omitempty"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	}

	if len(req.Filter.Must) < 2 {
		t.Errorf("Expected at least 2 filter conditions, got %d", len(req.Filter.Must))
	}

	t.Logf("✅ Qdrant Delete (By Filter) 成功，条件数：%d", len(req.Filter.Must))
}

// ============================================================================
// 完整工作流测试
// ============================================================================

// TestQdrant_FullCRUD 测试完整的 CRUD 操作
func TestQdrant_FullCRUD(t *testing.T) {
	qdrant := NewQdrantCustom()

	// 1. INSERT
	t.Log("\n=== 1. INSERT ===")
	point := map[string]interface{}{
		"id":     999,
		"vector": []float32{0.9, 0.9, 0.9},
		"payload": map[string]interface{}{
			"language": "golang",
			"status":   "active",
		},
	}

	insertBuilder := X().Custom(qdrant)
	insertBuilder.inserts = &[]Bb{{Value: point}}
	insertBuilt := insertBuilder.Build()

	insertJSON, err := insertBuilt.JsonOfInsert()
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	t.Logf("Insert JSON:\n%s", insertJSON)

	// 2. SELECT (Search)
	t.Log("\n=== 2. SELECT (Search) ===")
	queryVector := Vector{0.9, 0.9, 0.9}
	searchBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("status", "active").
		VectorSearch("embedding", queryVector, 10).
		Build()

	searchJSON, err := searchBuilt.JsonOfSelect()
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	t.Logf("Search JSON (truncated):\n%s", searchJSON[:200]+"...")

	// 3. UPDATE
	t.Log("\n=== 3. UPDATE ===")
	updateBuilder := X().Custom(qdrant).Eq("id", 999)
	updateBuilder.updates = &[]Bb{{Key: "status", Value: "archived"}}
	updateBuilt := updateBuilder.Build()

	updateJSON, err := updateBuilt.JsonOfUpdate()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	t.Logf("Update JSON:\n%s", updateJSON)

	// 4. DELETE
	t.Log("\n=== 4. DELETE ===")
	deleteBuilder := X().Custom(qdrant).Eq("id", 999)
	deleteBuilt := deleteBuilder.Build()
	deleteBuilt.Delete = true

	deleteJSON, err := deleteBuilt.JsonOfDelete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Logf("Delete JSON:\n%s", deleteJSON)

	t.Log("\n✅ 完整 CRUD 工作流测试成功！")
}

// ============================================================================
// 架构验证测试
// ============================================================================

// TestCustomInterface_QdrantAllOperations 验证 Custom 接口对所有操作的支持
func TestCustomInterface_QdrantAllOperations(t *testing.T) {
	t.Log("=== 验证 Custom 接口架构 ===")

	qdrant := NewQdrantCustom()

	// 测试 1: INSERT
	insertBuilder := X().Custom(qdrant)
	insertBuilder.inserts = &[]Bb{{Value: map[string]interface{}{
		"id":     1,
		"vector": []float32{0.1, 0.2},
	}}}
	insertBuilt := insertBuilder.Build()

	_, err := insertBuilt.JsonOfInsert()
	if err != nil {
		t.Errorf("INSERT failed: %v", err)
	}

	// 测试 2: UPDATE
	updateBuilder := X().Custom(qdrant).Eq("id", 1)
	updateBuilder.updates = &[]Bb{{Key: "test", Value: "value"}}
	updateBuilt := updateBuilder.Build()

	_, err = updateBuilt.JsonOfUpdate()
	if err != nil {
		t.Errorf("UPDATE failed: %v", err)
	}

	// 测试 3: DELETE
	deleteBuilder := X().Custom(qdrant).Eq("id", 1)
	deleteBuilt := deleteBuilder.Build()
	deleteBuilt.Delete = true

	_, err = deleteBuilt.JsonOfDelete()
	if err != nil {
		t.Errorf("DELETE failed: %v", err)
	}

	// 测试 4: SELECT
	selectBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		VectorSearch("embedding", Vector{0.1, 0.2}, 10).
		Build()

	_, err = selectBuilt.JsonOfSelect()
	if err != nil {
		t.Errorf("SELECT failed: %v", err)
	}

	t.Log("✅ Custom 接口支持所有 CRUD 操作")
	t.Log("✅ 一个 Generate() 方法处理 4 种操作")
	t.Log("✅ 架构验证成功！")
}
