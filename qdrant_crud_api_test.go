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
// Qdrant INSERT 测试（使用真实 API）
// ============================================================================

// TestQdrantAPI_Insert 测试 Qdrant Insert API
func TestQdrantAPI_Insert(t *testing.T) {
	// ⭐ 使用 Insert(func(ib *InsertBuilder)) - 与 SQL 一致的 API
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantCustom()).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 123).
				Set("vector", []float32{0.1, 0.2, 0.3, 0.4}).
				Set("language", "golang").
				Set("content", "func main() {...}")
		}).
		Build()

	jsonStr, err := built.JsonOfInsert()
	if err != nil {
		t.Fatalf("JsonOfInsert failed: %v", err)
	}

	t.Logf("=== Qdrant Insert API ===\n%s", jsonStr)

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

	// 验证 ID
	if int(req.Points[0].ID.(float64)) != 123 {
		t.Errorf("Expected ID 123, got %v", req.Points[0].ID)
	}

	// 验证 payload
	if req.Points[0].Payload["language"] != "golang" {
		t.Errorf("Expected language=golang, got %v", req.Points[0].Payload["language"])
	}

	t.Logf("✅ Qdrant Insert API works")
}

// ============================================================================
// Qdrant UPDATE 测试（使用真实 API）
// ============================================================================

// TestQdrantAPI_Update 测试 Qdrant Update API
func TestQdrantAPI_Update(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantCustom()).
		Eq("id", 123).
		Update(func(ub *UpdateBuilder) {
			ub.Set("language", "rust").
				Set("version", "1.75")
		}).
		Build()

	jsonStr, err := built.JsonOfUpdate()
	if err != nil {
		t.Fatalf("JsonOfUpdate failed: %v", err)
	}

	t.Logf("=== Qdrant Update API ===\n%s", jsonStr)

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

	if req.Payload["language"] != "rust" {
		t.Errorf("Expected language=rust, got %v", req.Payload["language"])
	}

	t.Logf("✅ Qdrant Update API works")
}

// TestQdrantAPI_UpdateByFilter 测试根据过滤器更新
func TestQdrantAPI_UpdateByFilter(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantCustom()).
		Eq("language", "golang").
		Gt("quality_score", 0.8).
		Update(func(ub *UpdateBuilder) {
			ub.Set("verified", true)
		}).
		Build()

	jsonStr, err := built.JsonOfUpdate()
	if err != nil {
		t.Fatalf("JsonOfUpdate failed: %v", err)
	}

	t.Logf("=== Qdrant Update by Filter ===\n%s", jsonStr)

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
		t.Errorf("Expected at least 2 conditions, got %d", len(req.Filter.Must))
	}

	t.Logf("✅ Qdrant Update by Filter works: %d conditions", len(req.Filter.Must))
}

// ============================================================================
// Qdrant DELETE 测试（使用真实 API）
// ============================================================================

// TestQdrantAPI_Delete 测试 Qdrant Delete API
func TestQdrantAPI_Delete(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantCustom()).
		Eq("id", 456).
		Build()

	jsonStr, err := built.JsonOfDelete()
	if err != nil {
		t.Fatalf("JsonOfDelete failed: %v", err)
	}

	t.Logf("=== Qdrant Delete API ===\n%s", jsonStr)

	var req struct {
		Points []interface{} `json:"points,omitempty"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point ID, got %d", len(req.Points))
	}

	t.Logf("✅ Qdrant Delete API works")
}

// ============================================================================
// 完整 CRUD 工作流测试（使用真实 API）
// ============================================================================

// TestQdrantAPI_FullCRUD 测试完整的 CRUD 工作流
func TestQdrantAPI_FullCRUD(t *testing.T) {
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

	insertBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", point["id"]).
				Set("vector", point["vector"]).
				Set("language", point["payload"].(map[string]interface{})["language"]).
				Set("status", point["payload"].(map[string]interface{})["status"])
		}).
		Build()

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
	updateBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 999).
		Update(func(ub *UpdateBuilder) {
			ub.Set("status", "archived")
		}).
		Build()

	updateJSON, err := updateBuilt.JsonOfUpdate()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	t.Logf("Update JSON:\n%s", updateJSON)

	// 4. DELETE
	t.Log("\n=== 4. DELETE ===")
	deleteBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 999).
		Build()

	deleteJSON, err := deleteBuilt.JsonOfDelete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Logf("Delete JSON:\n%s", deleteJSON)

	t.Log("\n✅ 完整 CRUD 工作流测试成功（使用真实 API）！")
}

// ============================================================================
// 架构验证：Custom 接口的统一性
// ============================================================================

// TestQdrantAPI_CustomInterfaceUnification 验证 Custom 接口的统一性
func TestQdrantAPI_CustomInterfaceUnification(t *testing.T) {
	qdrant := NewQdrantCustom()

	t.Log("=== 验证：同一个 Custom，处理所有操作 ===")

	// INSERT
	insertBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 1).
				Set("vector", []float32{0.1, 0.2}).
				Set("test", "value")
		}).
		Build()

	_, err := insertBuilt.JsonOfInsert()
	if err != nil {
		t.Errorf("INSERT failed: %v", err)
	}

	// UPDATE
	updateBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 1).
		Update(func(ub *UpdateBuilder) {
			ub.Set("test", "updated")
		}).
		Build()

	_, err = updateBuilt.JsonOfUpdate()
	if err != nil {
		t.Errorf("UPDATE failed: %v", err)
	}

	// DELETE
	deleteBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 1).
		Build()

	_, err = deleteBuilt.JsonOfDelete()
	if err != nil {
		t.Errorf("DELETE failed: %v", err)
	}

	// SELECT
	selectBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		VectorSearch("embedding", Vector{0.1, 0.2}, 10).
		Build()

	_, err = selectBuilt.JsonOfSelect()
	if err != nil {
		t.Errorf("SELECT failed: %v", err)
	}

	t.Log("✅ 同一个 Custom 实例处理所有 CRUD 操作")
	t.Log("✅ Custom 接口设计验证成功")
}
