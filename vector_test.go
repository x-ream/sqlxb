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

	t.Logf("=== SELECT 向量检索测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	if len(args) > 0 {
		t.Logf("Args[0] type: %T", args[0])
		t.Logf("Args[0] value: %v", args[0])

		// 检查查询参数类型
		switch args[0].(type) {
		case Vector:
			t.Logf("✅ 查询参数是 Vector 类型，driver.Valuer 会被调用")
		case string:
			t.Logf("⚠️ 查询参数是 string 类型")
		default:
			t.Logf("❓ 查询参数是未知类型: %T", args[0])
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

// 测试向量检索 + 标量过滤
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

	sql, args := Of(&CodeVector{}).
		VectorSearch("embedding", queryVector, 10).
		VectorDistance(L2Distance).
		Build().
		SqlOfVectorSearch()

	t.Logf("SQL: %s", sql)
	t.Logf("Distance Metric: L2Distance (<#>)")
	t.Logf("Args: %d", len(args))

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

	t.Logf("SQL: %s", sql)
	t.Logf("Threshold: < 0.3")
	t.Logf("Args: %d", len(args))

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

	t.Logf("SQL: %s", sql)
	t.Logf("Note: Empty language auto-ignored")
	t.Logf("Args: %d", len(args))

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
	t.Logf("Cosine Distance: %.4f", cosDist)
	if cosDist != 1.0 {
		t.Errorf("Expected cosine distance 1.0, got %f", cosDist)
	}

	// L2 距离
	l2Dist := vec1.Distance(vec2, L2Distance)
	t.Logf("L2 Distance: %.4f", l2Dist)
	expected := float32(1.414213) // sqrt(2)
	if abs(l2Dist-expected) > 0.001 {
		t.Errorf("Expected L2 distance ~1.414, got %f", l2Dist)
	}
}

// 测试向量归一化
func TestVector_Normalize(t *testing.T) {
	vec := Vector{3.0, 4.0} // 长度为 5
	normalized := vec.Normalize()

	t.Logf("Original: %v", vec)
	t.Logf("Normalized: %v", normalized)

	// 归一化后长度应该为 1
	expected := Vector{0.6, 0.8}

	if abs(normalized[0]-expected[0]) > 0.001 {
		t.Errorf("Expected normalized[0] = 0.6, got %f", normalized[0])
	}

	if abs(normalized[1]-expected[1]) > 0.001 {
		t.Errorf("Expected normalized[1] = 0.8, got %f", normalized[1])
	}
}

// 测试向量插入
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

	t.Logf("=== INSERT 测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	for i, arg := range args {
		t.Logf("Args[%d] type: %T", i, arg)
		t.Logf("Args[%d] value: %v", i, arg)
	}

	// 验证 SQL 包含所有字段
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

// 测试向量更新
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

	t.Logf("=== UPDATE 测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	for i, arg := range args {
		t.Logf("Args[%d] type: %T", i, arg)
		t.Logf("Args[%d] value: %v", i, arg)
	}

	// 验证 SQL
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

// 测试向量类型在 Set() 中的处理
func TestVector_SetBehavior(t *testing.T) {
	vec := Vector{1.0, 2.0, 3.0}

	// 测试 InsertBuilder.Set()
	sql, args := Of(&CodeVector{}).
		Insert(func(ib *InsertBuilder) {
			ib.Set("embedding", vec)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== Vector Set() 行为测试 ===")
	t.Logf("原始 Vector: %v (类型: %T)", vec, vec)
	t.Logf("SQL: %s", sql)

	if len(args) > 0 {
		t.Logf("Set() 后 args[0] 类型: %T", args[0])
		t.Logf("Set() 后 args[0] 值: %v", args[0])

		// 关键检查：args[0] 是 Vector 还是 string？
		switch args[0].(type) {
		case Vector:
			t.Logf("✅ args[0] 是 Vector 类型，driver.Valuer 会被调用")
		case string:
			t.Logf("⚠️ args[0] 是 string 类型，已被 JSON Marshal")
			t.Logf("⚠️ driver.Valuer 不会被调用")
		case []float32:
			t.Logf("✅ args[0] 是 []float32 类型")
		default:
			t.Logf("❓ args[0] 是未知类型: %T", args[0])
		}
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
