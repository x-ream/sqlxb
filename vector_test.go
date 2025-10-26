// Copyright 2020 io.xream.sqlxb
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
package sqlxb

import (
	"testing"
	"time"
)

// 测试用的向量数据模型
type CodeVector struct {
	Id        int64     `db:"id"`
	Content   string    `db:"content"`
	Embedding Vector    `db:"embedding"`
	Language  string    `db:"language"`
	Layer     string    `db:"layer"`
	CreatedAt time.Time `db:"created_at"`
}

func (CodeVector) TableName() string {
	return "code_vectors"
}

// 测试基础向量检索
func TestVectorSearch_Basic(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	sql, args := Of(&CodeVector{}).
		VectorSearch("embedding", queryVector, 10).
		Build().
		SqlOfVectorSearch()

	expectedSQL := "SELECT *, embedding <-> ? AS distance FROM code_vectors ORDER BY distance LIMIT 10"

	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s\nGot: %s", expectedSQL, sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

// 测试向量检索 + 标量过滤
func TestVectorSearch_WithScalarFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	sql, args := Of(&CodeVector{}).
		Eq("language", "golang").
		Eq("layer", "repository").
		VectorSearch("embedding", queryVector, 5).
		Build().
		SqlOfVectorSearch()

	// SQL 应该包含 WHERE 条件
	if !containsString(sql, "WHERE") {
		t.Errorf("Expected WHERE clause in SQL: %s", sql)
	}

	if !containsString(sql, "language") {
		t.Errorf("Expected language filter in SQL: %s", sql)
	}

	if !containsString(sql, "ORDER BY distance") {
		t.Errorf("Expected ORDER BY distance in SQL: %s", sql)
	}

	if !containsString(sql, "LIMIT 5") {
		t.Errorf("Expected LIMIT 5 in SQL: %s", sql)
	}

	// args: queryVector, "golang", "repository"
	if len(args) < 3 {
		t.Errorf("Expected at least 3 args, got %d", len(args))
	}
}

// 测试 L2 距离
func TestVectorSearch_L2Distance(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	sql, _ := Of(&CodeVector{}).
		VectorSearch("embedding", queryVector, 10).
		VectorDistance(L2Distance).
		Build().
		SqlOfVectorSearch()

	// SQL 应该使用 <#> 运算符
	if !containsString(sql, "<#>") {
		t.Errorf("Expected <#> (L2 distance) in SQL: %s", sql)
	}
}

// 测试向量距离过滤
func TestVectorDistanceFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	sql, args := Of(&CodeVector{}).
		Eq("language", "golang").
		VectorDistanceFilter("embedding", queryVector, "<", 0.3).
		Build().
		SqlOfVectorSearch()

	// SQL 应该包含距离过滤条件
	if !containsString(sql, "<-> ?") {
		t.Errorf("Expected distance filter in SQL: %s", sql)
	}

	if !containsString(sql, "< ?") {
		t.Errorf("Expected threshold comparison in SQL: %s", sql)
	}

	// args: "golang", queryVector, 0.3
	// 注意：VectorDistanceFilter 不会在顶部添加向量，只在 WHERE 中
	if len(args) < 3 {
		t.Errorf("Expected at least 3 args, got %d", len(args))
	}
}

// 测试自动忽略 nil
func TestVectorSearch_AutoIgnoreNil(t *testing.T) {
	queryVector := Vector{0.1, 0.2}

	// language 为空字符串，应该被忽略
	sql, args := Of(&CodeVector{}).
		Eq("language", "").
		Eq("layer", "repository").
		VectorSearch("embedding", queryVector, 10).
		Build().
		SqlOfVectorSearch()

	// SQL 不应该包含 language（因为是空字符串）
	// 但应该包含 layer
	if containsString(sql, "language") {
		t.Errorf("Empty language should be ignored, but found in SQL: %s", sql)
	}

	if !containsString(sql, "layer") {
		t.Errorf("Expected layer filter in SQL: %s", sql)
	}

	// args: queryVector, "repository"
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

// 测试向量类型的距离计算
func TestVector_Distance(t *testing.T) {
	vec1 := Vector{1.0, 0.0, 0.0}
	vec2 := Vector{0.0, 1.0, 0.0}

	// 余弦距离
	cosDist := vec1.Distance(vec2, CosineDistance)
	if cosDist != 1.0 {
		t.Errorf("Expected cosine distance 1.0, got %f", cosDist)
	}

	// L2 距离
	l2Dist := vec1.Distance(vec2, L2Distance)
	expected := float32(1.414213) // sqrt(2)
	if abs(l2Dist-expected) > 0.001 {
		t.Errorf("Expected L2 distance ~1.414, got %f", l2Dist)
	}
}

// 测试向量归一化
func TestVector_Normalize(t *testing.T) {
	vec := Vector{3.0, 4.0} // 长度为 5
	normalized := vec.Normalize()

	// 归一化后长度应该为 1
	expected := Vector{0.6, 0.8}

	if abs(normalized[0]-expected[0]) > 0.001 {
		t.Errorf("Expected normalized[0] = 0.6, got %f", normalized[0])
	}

	if abs(normalized[1]-expected[1]) > 0.001 {
		t.Errorf("Expected normalized[1] = 0.8, got %f", normalized[1])
	}
}

// 辅助函数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

