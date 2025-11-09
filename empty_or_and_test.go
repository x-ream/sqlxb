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
)

// 测试空 OR 条件被自动过滤
func TestEmptyOr_Filtered(t *testing.T) {
	// 构建查询：包含一个空的 OR
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang"). // ✅ 有效
		Or(func(cb *CondBuilder) {
			// ⭐ OR 中的所有条件都是 nil/0，会被过滤
			cb.Eq("category", "")
			cb.Gt("rank", 0)
			cb.Lt("score", 0)
		}).
		Gt("quality", 0.8). // ✅ 有效
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== 空 OR 过滤测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ 验证：SQL 不应该包含 OR
	if containsString(sql, "OR") {
		t.Errorf("Empty OR should be filtered out, but found in SQL")
	}

	// ⭐ 验证：只有 2 个有效条件
	// language = golang AND quality > 0.8
	if !containsString(sql, "language") {
		t.Errorf("Expected language condition")
	}
	if !containsString(sql, "quality") {
		t.Errorf("Expected quality condition")
	}

	t.Logf("✅ 空 OR 被成功过滤")
}

// 测试空 AND 条件被自动过滤
func TestEmptyAnd_Filtered(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		And(func(cb *CondBuilder) {
			// ⭐ AND 中的所有条件都是 nil/0
			cb.Eq("category", "")
			cb.Gt("rank", 0)
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== 空 AND 过滤测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ 验证：不应该有额外的 AND () 结构
	// 正常应该是：language = ? AND quality > ?
	if len(args) != 2 {
		t.Errorf("Expected 2 args (language, quality), got %d", len(args))
	}

	t.Logf("✅ 空 AND 被成功过滤")
}

// 测试混合场景：部分 OR 为空
func TestPartialEmptyOr(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			// ⭐ 第一个 OR：有效条件
			cb.Eq("category", "service")
			cb.Eq("category", "repository")
		}).
		Or(func(cb *CondBuilder) {
			// ⭐ 第二个 OR：全部为空
			cb.Eq("tag", "")
			cb.Gt("count", 0)
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== 部分空 OR 测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ 应该包含第一个 OR（有效）
	if !containsString(sql, "category") {
		t.Errorf("Expected first OR with category")
	}

	// ⭐ 不应该包含第二个 OR（空）
	if containsString(sql, "tag") || containsString(sql, "count") {
		t.Errorf("Second empty OR should be filtered")
	}

	t.Logf("✅ 只有有效的 OR 被保留")
}

// 测试 Qdrant JSON：空 OR 被过滤
func TestQdrant_EmptyOr_Filtered(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantCustom()).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			// 空 OR
			cb.Eq("category", "")
			cb.Gt("rank", 0)
		}).
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== Qdrant 空 OR 测试 ===\n%s", jsonStr)

	// ⭐ JSON 不应该包含 should（OR 对应 should）
	if containsString(jsonStr, "should") {
		t.Errorf("Empty OR should be filtered from Qdrant JSON")
	}

	// ⭐ 应该只有 language 条件
	if !containsString(jsonStr, "language") {
		t.Errorf("Expected language condition")
	}

	t.Logf("✅ Qdrant JSON 中空 OR 被过滤")
}

// 测试嵌套空 OR
func TestNestedEmptyOr(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			cb.And(func(cb2 *CondBuilder) {
				// 嵌套的 AND 也是空的
				cb2.Eq("tag", "")
			})
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== 嵌套空 OR 测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ 整个 OR 应该被过滤
	if containsString(sql, "OR") {
		t.Errorf("Nested empty OR should be filtered")
	}

	// 应该只有 2 个条件
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Logf("✅ 嵌套空 OR 被成功过滤")
}

// 测试 Bool() 函数配合空 OR
func TestBoolWithEmptyOr(t *testing.T) {
	includeOptional := false

	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Bool(func() bool { return includeOptional }, func(cb *CondBuilder) {
			cb.Or(func(cb2 *CondBuilder) {
				cb2.Eq("category", "")
			})
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== Bool + 空 OR 测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ includeOptional = false，整个 Bool 块不执行
	// 即使执行，空 OR 也会被过滤
	if containsString(sql, "OR") {
		t.Errorf("OR should not appear")
	}

	t.Logf("✅ Bool + 空 OR 正确处理")
}

// 测试完整场景：多层过滤
func TestComplexFiltering(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantCustom()).
		Eq("language", "golang"). // ✅ 有效
		Eq("category", "").       // ⭐ 空字符串，单个条件过滤
		Or(func(cb *CondBuilder) {
			// ⭐ 空 OR，整个 OR 过滤
			cb.Eq("tag1", "")
			cb.Gt("count1", 0)
		}).
		And(func(cb *CondBuilder) {
			// ⭐ 部分有效的 AND
			cb.Eq("status", "active") // ✅ 有效
			cb.Gt("rank", 0)          // ⭐ 0，被过滤
		}).
		Or(func(cb *CondBuilder) {
			// ⭐ 全部有效的 OR
			cb.Eq("layer", "service")
			cb.Eq("layer", "repository")
		}).
		VectorSearch("embedding", queryVector, 20).
		Build()

	// PostgreSQL
	sql, args := built.SqlOfVectorSearch()
	t.Logf("=== 复杂过滤测试 - PostgreSQL ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))

	// 验证结果：
	// 1. category="" 被过滤
	// 2. 第一个空 OR 被过滤
	// 3. AND 中 rank=0 被过滤，但 status="active" 保留
	// 4. 第二个 OR 保留（有效）

	if containsString(sql, "category") {
		t.Errorf("category='' should be filtered")
	}

	if !containsString(sql, "status") {
		t.Errorf("status condition should exist")
	}

	if !containsString(sql, "layer") {
		t.Errorf("layer OR should exist")
	}

	// Qdrant
	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}
	t.Logf("\n=== 复杂过滤测试 - Qdrant ===\n%s", jsonStr)

	t.Logf("✅ 多层过滤正确工作")
}
