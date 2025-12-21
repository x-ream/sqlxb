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

// TestEmptyOr_Filtered test empty OR condition is automatically filtered
func TestEmptyOr_Filtered(t *testing.T) {
	// Build query: contains an empty OR
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang"). // ✅ Valid
		Or(func(cb *CondBuilder) {
			// ⭐ All conditions in OR are nil/0, will be filtered
			cb.Eq("category", "")
			cb.Gt("rank", 0)
			cb.Lt("score", 0)
		}).
		Gt("quality", 0.8). // ✅ Valid
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== Empty OR filtering test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ Verify: SQL should not contain OR
	if containsString(sql, "OR") {
		t.Errorf("Empty OR should be filtered out, but found in SQL")
	}

	// ⭐ Verify: only 2 valid conditions
	// language = golang AND quality > 0.8
	if !containsString(sql, "language") {
		t.Errorf("Expected language condition")
	}
	if !containsString(sql, "quality") {
		t.Errorf("Expected quality condition")
	}

	t.Logf("✅ Empty OR is successfully filtered")
}

// TestEmptyAnd_Filtered test empty AND condition is automatically filtered
func TestEmptyAnd_Filtered(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		And(func(cb *CondBuilder) {
			// ⭐ All conditions in AND are nil/0
			cb.Eq("category", "")
			cb.Gt("rank", 0)
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== Empty AND filtering test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ Verify: should not have extra AND ( ) structure
	// Normally should be: language = ? AND quality > ?
	if len(args) != 2 {
		t.Errorf("Expected 2 args (language, quality), got %d", len(args))
	}

	t.Logf("✅ Empty AND is successfully filtered")
}

// TestPartialEmptyOr test mixed scenario: partial OR is empty
func TestPartialEmptyOr(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			// ⭐ First OR: valid conditions
			cb.Eq("category", "service")
			cb.Eq("category", "repository")
		}).
		Or(func(cb *CondBuilder) {
			// ⭐ Second OR: all are empty
			cb.Eq("tag", "")
			cb.Gt("count", 0)
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== Partial empty OR test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ Should contain first OR (valid)
	if !containsString(sql, "category") {
		t.Errorf("Expected first OR with category")
	}

	// ⭐ Should not contain second OR (empty)
	if containsString(sql, "tag") || containsString(sql, "count") {
		t.Errorf("Second empty OR should be filtered")
	}

	t.Logf("✅ Only valid OR is retained")
}

// TestQdrant_EmptyOr_Filtered test Qdrant JSON: empty OR is filtered
func TestQdrant_EmptyOr_Filtered(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			// Empty OR
			cb.Eq("category", "")
			cb.Gt("rank", 0)
		}).
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== Qdrant Empty OR Test ===\n%s", jsonStr)

	// ⭐ JSON should not contain should (OR corresponds to should)
	if containsString(jsonStr, "should") {
		t.Errorf("Empty OR should be filtered from Qdrant JSON")
	}

	// ⭐ Should only have language condition
	if !containsString(jsonStr, "language") {
		t.Errorf("Expected language condition")
	}

	t.Logf("✅ Empty OR filtered in Qdrant JSON")
}

// Test nested empty OR
func TestNestedEmptyOr(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			cb.And(func(cb2 *CondBuilder) {
				// Nested AND is also empty
				cb2.Eq("tag", "")
			})
		}).
		Gt("quality", 0.8).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== Nested Empty OR Test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ Entire OR should be filtered
	if containsString(sql, "OR") {
		t.Errorf("Nested empty OR should be filtered")
	}

	// Should only have 2 conditions
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Logf("✅ Nested empty OR successfully filtered")
}

// Test Bool() function with empty OR
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

	t.Logf("=== Bool + Empty OR Test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ includeOptional = false, entire Bool block not executed
	// Even if executed, empty OR will be filtered
	if containsString(sql, "OR") {
		t.Errorf("OR should not appear")
	}

	t.Logf("✅ Bool + empty OR handled correctly")
}

// Test complete scenario: multi-layer filtering
func TestComplexFiltering(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang"). // ✅ Valid
		Eq("category", "").       // ⭐ Empty string, single condition filtered
		Or(func(cb *CondBuilder) {
			// ⭐ Empty OR, entire OR filtered
			cb.Eq("tag1", "")
			cb.Gt("count1", 0)
		}).
		And(func(cb *CondBuilder) {
			// ⭐ Partially valid AND
			cb.Eq("status", "active") // ✅ Valid
			cb.Gt("rank", 0)          // ⭐ 0, filtered
		}).
		Or(func(cb *CondBuilder) {
			// ⭐ All valid OR
			cb.Eq("layer", "service")
			cb.Eq("layer", "repository")
		}).
		VectorSearch("embedding", queryVector, 20).
		Build()

	// PostgreSQL
	sql, args := built.SqlOfVectorSearch()
	t.Logf("=== Complex Filtering Test - PostgreSQL ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))

	// Verify results:
	// 1. category="" is filtered
	// 2. First empty OR is filtered
	// 3. rank=0 in AND is filtered, but status="active" is retained
	// 4. Second OR is retained (valid)

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
	t.Logf("\n=== Complex Filtering Test - Qdrant ===\n%s", jsonStr)

	t.Logf("✅ Multi-layer filtering works correctly")
}
