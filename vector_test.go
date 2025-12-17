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
	"testing"
	"time"
)

// Test vector data model
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

// Test basic vector search
func TestVectorSearch_Basic(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	sql, args := Of(&CodeVector{}).
		VectorSearch("embedding", queryVector, 10).
		Build().
		SqlOfVectorSearch()

	t.Logf("=== SELECT vector search test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	if len(args) > 0 {
		t.Logf("Args[0] type: %T", args[0])
		t.Logf("Args[0] value: %v", args[0])

		// Check query parameter type
		switch args[0].(type) {
		case Vector:
			t.Logf("✅ Query parameter is Vector type, driver.Valuer will be called")
		case string:
			t.Logf("⚠️ Query parameter is string type")
		default:
			t.Logf("❓ Query parameter is unknown type: %T", args[0])
		}
	}

	expectedSQL := "SELECT *, embedding <-> ? AS distance FROM code_vectors ORDER BY distance LIMIT 10"

	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s\nGot: %s", expectedSQL, sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

// Test vector search + scalar filter
func TestVectorSearch_WithScalarFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	sql, args := Of(&CodeVector{}).
		Eq("language", "golang").
		Eq("layer", "repository").
		VectorSearch("embedding", queryVector, 5).
		Build().
		SqlOfVectorSearch()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %d", len(args))

	// SQL should contain WHERE condition
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

// Test L2 distance
func TestVectorSearch_L2Distance(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	sql, args := Of(&CodeVector{}).
		VectorSearch("embedding", queryVector, 10).
		VectorDistance(L2Distance).
		Build().
		SqlOfVectorSearch()

	t.Logf("SQL: %s", sql)
	t.Logf("Distance Metric: L2Distance (<#>)")
	t.Logf("Args: %d", len(args))

	// SQL should use <#> operator
	if !containsString(sql, "<#>") {
		t.Errorf("Expected <#> (L2 distance) in SQL: %s", sql)
	}
}

// Test vector distance filtering
func TestVectorDistanceFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	sql, args := Of(&CodeVector{}).
		Eq("language", "golang").
		VectorDistanceFilter("embedding", queryVector, "<", 0.3).
		Build().
		SqlOfVectorSearch()

	t.Logf("SQL: %s", sql)
	t.Logf("Threshold: < 0.3")
	t.Logf("Args: %d", len(args))

	// SQL should contain distance filter condition
	if !containsString(sql, "<-> ?") {
		t.Errorf("Expected distance filter in SQL: %s", sql)
	}

	if !containsString(sql, "< ?") {
		t.Errorf("Expected threshold comparison in SQL: %s", sql)
	}

	// args: "golang", queryVector, 0.3
	// Note: VectorDistanceFilter does not add vector at the top, only in WHERE
	if len(args) < 3 {
		t.Errorf("Expected at least 3 args, got %d", len(args))
	}
}

// Test auto ignore nil
func TestVectorSearch_AutoIgnoreNil(t *testing.T) {
	queryVector := Vector{0.1, 0.2}

	// language is an empty string, should be ignored
	sql, args := Of(&CodeVector{}).
		Eq("language", "").
		Eq("layer", "repository").
		VectorSearch("embedding", queryVector, 10).
		Build().
		SqlOfVectorSearch()

	t.Logf("SQL: %s", sql)
	t.Logf("Note: Empty language is auto-ignored")
	t.Logf("Args: %d", len(args))

	// SQL should not contain language (because it's an empty string)
	// 但应该包含 layer
	if containsString(sql, "language") {
		t.Errorf("Empty language should be ignored, but found in SQL: %s", sql)
	}

	if !containsString(sql, "layer") {
		t.Errorf("Expected layer filter in SQL: %s", sql)
	}

	// Args: queryVector, "repository"
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
}

// Test vector distance calculation
func TestVector_Distance(t *testing.T) {
	vec1 := Vector{1.0, 0.0, 0.0}
	vec2 := Vector{0.0, 1.0, 0.0}

	// Cosine distance
	cosDist := vec1.Distance(vec2, CosineDistance)
	t.Logf("Cosine Distance: %.4f", cosDist)
	if cosDist != 1.0 {
		t.Errorf("Expected cosine distance 1.0, got %f", cosDist)
	}

	// L2 distance
	l2Dist := vec1.Distance(vec2, L2Distance)
	t.Logf("L2 Distance: %.4f", l2Dist)
	expected := float32(1.414213) // sqrt(2)
	if abs(l2Dist-expected) > 0.001 {
		t.Errorf("Expected L2 distance ~1.414, got %f", l2Dist)
	}
}

// Test vector normalization
func TestVector_Normalize(t *testing.T) {
	vec := Vector{3.0, 4.0} // Length is 5
	normalized := vec.Normalize()

	t.Logf("Original: %v", vec)
	t.Logf("Normalized: %v", normalized)

	// After normalization, length should be 1
	expected := Vector{0.6, 0.8}

	if abs(normalized[0]-expected[0]) > 0.001 {
		t.Errorf("Expected normalized[0] = 0.6, got %f", normalized[0])
	}

	if abs(normalized[1]-expected[1]) > 0.001 {
		t.Errorf("Expected normalized[1] = 0.8, got %f", normalized[1])
	}
}

// Test vector insert
func TestVector_Insert(t *testing.T) {
	code := &CodeVector{
		Content:   "func main() { fmt.Println(\"Hello\") }",
		Embedding: Vector{0.1, 0.2, 0.3, 0.4},
		Language:  "golang",
		Layer:     "main",
	}

	sql, args := Of(code).
		Insert(func(ib *InsertBuilder) {
			ib.Set("content", code.Content).
				Set("embedding", code.Embedding).
				Set("language", code.Language).
				Set("layer", code.Layer)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== INSERT test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	for i, arg := range args {
		t.Logf("Args[%d] type: %T", i, arg)
		t.Logf("Args[%d] value: %v", i, arg)
	}

	// Verify SQL contains all fields
	if !containsString(sql, "content") {
		t.Errorf("Expected content field in SQL: %s", sql)
	}
	if !containsString(sql, "embedding") {
		t.Errorf("Expected embedding field in SQL: %s", sql)
	}
	if !containsString(sql, "language") {
		t.Errorf("Expected language field in SQL: %s", sql)
	}
}

// Test vector update
func TestVector_Update(t *testing.T) {
	newEmbedding := Vector{0.5, 0.6, 0.7, 0.8}

	sql, args := Of(&CodeVector{}).
		Update(func(ub *UpdateBuilder) {
			ub.Set("embedding", newEmbedding).
				Set("language", "golang")
		}).
		Eq("id", 123).
		Build().
		SqlOfUpdate()

	t.Logf("=== UPDATE test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	for i, arg := range args {
		t.Logf("Args[%d] type: %T", i, arg)
		t.Logf("Args[%d] value: %v", i, arg)
	}

	// Verify SQL
	if !containsString(sql, "UPDATE") {
		t.Errorf("Expected UPDATE in SQL: %s", sql)
	}
	if !containsString(sql, "embedding") {
		t.Errorf("Expected embedding field in SQL: %s", sql)
	}
	if !containsString(sql, "WHERE") {
		t.Errorf("Expected WHERE clause in SQL: %s", sql)
	}
}

// Test vector type handling in Set()
func TestVector_SetBehavior(t *testing.T) {
	vec := Vector{1.0, 2.0, 3.0}

	// Test InsertBuilder.Set()
	sql, args := Of(&CodeVector{}).
		Insert(func(ib *InsertBuilder) {
			ib.Set("embedding", vec)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== Vector Set() behavior test ===")
	t.Logf("Original Vector: %v (type: %T)", vec, vec)
	t.Logf("SQL: %s", sql)

	if len(args) > 0 {
		t.Logf("After Set(), args[0] type: %T", args[0])
		t.Logf("After Set(), args[0] value: %v", args[0])

		// Key check: args[0] is Vector or string?
		switch args[0].(type) {
		case Vector:
			t.Logf("✅ args[0] is Vector type, driver.Valuer will be called")
		case string:
			t.Logf("⚠️ args[0] is string type, has been JSON Marshal")
			t.Logf("⚠️ driver.Valuer will not be called")
		case []float32:
			t.Logf("✅ args[0] is []float32 type")
		default:
			t.Logf("❓ args[0] is unknown type: %T", args[0])
		}
	}
}

// Helper functions
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
