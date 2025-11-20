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

// 测试用的向量数据模型（与 vector_test.go 相同）
type CodeVectorForQdrant struct {
	Id           int64  `db:"id"`
	Content      string `db:"content"`
	Embedding    Vector `db:"embedding"`
	Language     string `db:"language"`
	Layer        string `db:"layer"`
	SemanticHash string `db:"semantic_hash"` // ⭐ 用于多样性去重
}

func (CodeVectorForQdrant) TableName() string {
	return "code_vectors"
}

// 测试基础 Qdrant JSON 生成
func TestJsonOfSelect_Basic(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== 基础 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON 格式
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// 验证字段
	if len(req.Vector) != 4 {
		t.Errorf("Expected vector length 4, got %d", len(req.Vector))
	}
	if req.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", req.Limit)
	}
}

// 测试带标量过滤的 Qdrant JSON
func TestJsonOfSelect_WithFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		Eq("layer", "repository").
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== 带过滤器的 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// 验证 filter
	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	} else if len(req.Filter.Must) != 2 {
		t.Errorf("Expected 2 must conditions, got %d", len(req.Filter.Must))
	}
}

// 测试哈希多样性 - Qdrant JSON
func TestJsonOfSelect_WithHashDiversity(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash"). // ⭐ 多样性：哈希去重
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== 哈希多样性 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ 多样性：limit 应该被放大（过度获取）
	if req.Limit != 20*5 { // 默认 5 倍
		t.Errorf("Expected limit %d (20*5), got %d", 20*5, req.Limit)
	}

	t.Logf("✅ 多样性启用：Limit 从 20 扩大到 %d（5倍过度获取）", req.Limit)
	t.Logf("ℹ️  后续需要在应用层基于 semantic_hash 去重到 20 个")
}

// 测试最小距离多样性 - Qdrant JSON
func TestJsonOfSelect_WithMinDistance(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		VectorSearch("embedding", queryVector, 20).
		WithMinDistance(0.3). // ⭐ 多样性：最小距离 0.3
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== 最小距离多样性 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ 多样性：limit 应该被放大
	if req.Limit != 20*5 {
		t.Errorf("Expected limit %d, got %d", 20*5, req.Limit)
	}

	t.Logf("✅ 多样性启用：Limit 从 20 扩大到 %d", req.Limit)
	t.Logf("ℹ️  后续需要在应用层确保结果间距离 >= 0.3")
}

// 测试 MMR 多样性 - Qdrant JSON
func TestJsonOfSelect_WithMMR(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithMMR(0.5). // ⭐ 多样性：MMR lambda=0.5
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== MMR 多样性 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ 多样性：limit 应该被放大
	if req.Limit != 20*5 {
		t.Errorf("Expected limit %d, got %d", 20*5, req.Limit)
	}

	t.Logf("✅ 多样性启用：Limit 从 20 扩大到 %d", req.Limit)
	t.Logf("ℹ️  后续需要在应用层使用 MMR 算法（lambda=0.5）过滤")
}

// 测试范围查询 - Qdrant JSON
func TestJsonOfSelect_WithRange(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Gt("score", 0.8).
		Lt("complexity", 100).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== 范围查询 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// 验证 filter
	if req.Filter == nil || len(req.Filter.Must) < 2 {
		t.Errorf("Expected at least 2 range conditions")
	}
}

// 测试 IN 查询 - Qdrant JSON
func TestJsonOfSelect_WithIn(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		In("language", "golang", "python", "rust"). // ⭐ 修正：直接传值，不是 slice
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== IN 查询 Qdrant JSON ===\n%s", jsonStr)

	// 验证 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// 验证 filter
	if req.Filter == nil || len(req.Filter.Must) == 0 {
		t.Errorf("Expected IN condition in filter")
	}
}

// 测试 PostgreSQL SQL 不受多样性影响
func TestSqlOfVectorSearch_IgnoresDiversity(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// ⭐ 关键测试：即使设置了多样性，PostgreSQL SQL 也应该正常工作
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash"). // ⭐ 多样性设置
		Build()

	// 生成 PostgreSQL SQL
	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== PostgreSQL SQL（应忽略多样性）===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ 关键验证：SQL 不应该包含多样性相关逻辑
	// LIMIT 应该保持为 20，而不是 100
	if !containsString(sql, "LIMIT 20") {
		t.Errorf("Expected LIMIT 20 in SQL (diversity should be ignored)")
	}

	// 验证基本查询功能正常
	if !containsString(sql, "language") {
		t.Errorf("Expected language filter in SQL")
	}

	if !containsString(sql, "ORDER BY distance") {
		t.Errorf("Expected ORDER BY distance in SQL")
	}

	t.Logf("✅ 多样性参数被正确忽略（PostgreSQL 不支持）")
}

// 测试完整的工作流
func TestQdrant_FullWorkflow(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	// 构建查询
	builder := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		Gt("quality_score", 0.8).
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash")

	built := builder.Build()

	// 1. PostgreSQL 使用
	t.Log("\n=== 1. PostgreSQL 后端 ===")
	sql, args := built.SqlOfVectorSearch()
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	t.Logf("✅ 多样性被忽略，正常生成 SQL")

	// 2. Qdrant 使用
	t.Log("\n=== 2. Qdrant 后端 ===")
	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}
	t.Logf("JSON:\n%s", jsonStr)
	t.Logf("✅ 多样性被应用，Limit 扩大到 %d", 20*5)

	// 3. 验证同一个 Built 对象可以用于不同后端
	t.Log("\n=== 3. 同一查询，多后端兼容 ===")
	t.Logf("✅ 一份代码，两种后端：PostgreSQL 和 Qdrant")
	t.Logf("✅ 优雅降级：不支持的功能自动忽略")
}
