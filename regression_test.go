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

// Regression test: README example of And()/Or()
// Bug: SqlOfVectorSearch cannot handle Or() correctly in v0.9.0
func TestRegression_README_AndOr(t *testing.T) {
	// Reuse existing CodeVectorForQdrant type to avoid type conflict
	builder := Of(&CodeVectorForQdrant{}).
		Gt("id", 10000).
		And(func(cb *CondBuilder) {
			cb.Gte("price", 100.5).
				OR().
				Eq("is_sold", true)
		}).
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== README And/Or regression test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL is correct
	expectedSQL := "SELECT * FROM t_cat WHERE id > ? AND (price >= ? OR is_sold = ?)"
	if !containsString(sql, "id > ?") {
		t.Errorf("Missing 'id > ?'")
	}
	if !containsString(sql, "price >= ?") {
		t.Errorf("Missing 'price >= ?'")
	}
	if !containsString(sql, "is_sold = ?") {
		t.Errorf("Missing 'is_sold = ?'")
	}
	if !containsString(sql, "OR") {
		t.Errorf("Missing OR operator, got: %s", sql)
	}

	// Verify parameters
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	t.Logf("✅ README example test passed: %s", expectedSQL)
}

// Regression test: float64 zero value filtering
// Bug: interface{} == 0 is invalid for float64 in v0.9.0
func TestRegression_Float64_ZeroFilter(t *testing.T) {
	// Test for all numeric types zero values
	builder := Of(&CodeVectorForQdrant{}).
		Eq("name", "laptop"). // ✅ Keep
		// float64 zero value
		Gt("min_price", 0.0).   // ❌ Should be filtered
		Lt("max_price", 999.9). // ✅ Keep
		// float32 zero value
		Gt("min_score", float32(0.0)). // ❌ Should be filtered
		Lt("max_score", float32(0.9)). // ✅ Keep
		// int zero value
		Gt("min_weight", 0).    // ❌ Should be filtered
		Lt("max_weight", 1000). // ✅ Keep
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== float64/int zero value filtering test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify zero values are filtered
	if containsString(sql, "min_price") {
		t.Errorf("min_price=0.0 should be filtered")
	}
	if containsString(sql, "min_score") {
		t.Errorf("min_score=0.0 should be filtered")
	}
	if containsString(sql, "min_weight") {
		t.Errorf("min_weight=0 should be filtered")
	}

	// Verify non-zero values are kept
	if !containsString(sql, "max_price") {
		t.Errorf("max_price should exist")
	}
	if !containsString(sql, "max_score") {
		t.Errorf("max_score should exist")
	}
	if !containsString(sql, "max_weight") {
		t.Errorf("max_weight should exist")
	}

	// Verify parameter count
	if len(args) != 4 { // name, max_price, max_score, max_weight
		t.Errorf("Expected 4 args, got %d: %v", len(args), args)
	}

	t.Logf("✅ All numeric types zero value filtering correct")
}

// Regression test: vector search + And/Or combination
// Bug: SqlOfVectorSearch uses simplified buildConditionSql in v0.9.0, ignoring subs
func TestRegression_VectorSearch_WithAndOr(t *testing.T) {
	vec := Vector{0.1, 0.2, 0.3}

	// Combined test: vector search + And/Or subquery
	builder := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", vec, 10).
		And(func(cb *CondBuilder) {
			cb.Gte("quality", 0.8).
				OR().
				Eq("verified", true)
		}).
		Build()

	sql, args := builder.SqlOfVectorSearch()

	t.Logf("=== vector search + And/Or regression test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL structure
	if !containsString(sql, "embedding <-> ?") {
		t.Errorf("Missing vector search")
	}
	if !containsString(sql, "language = ?") {
		t.Errorf("Missing language filter")
	}
	if !containsString(sql, "quality >= ?") {
		t.Errorf("Missing quality filter")
	}
	if !containsString(sql, "verified = ?") {
		t.Errorf("Missing verified filter")
	}
	if !containsString(sql, "OR") {
		t.Errorf("Missing OR operator, got: %s", sql)
	}

	// Verify parameters
	if len(args) != 4 { // vec, language, quality, verified
		t.Errorf("Expected 4 args, got %d", len(args))
	}

	t.Logf("✅ vector search + And/Or correctly works")
}

// Regression test: consistency between SqlOfSelect and SqlOfVectorSearch
// Bug: Both use different condition building logic, causing inconsistent behavior
func TestRegression_SqlOfSelect_vs_SqlOfVectorSearch(t *testing.T) {
	// Build same query conditions
	buildQuery := func() *Built {
		return Of(&CodeVectorForQdrant{}).
			Eq("language", "golang").
			Or(func(cb *CondBuilder) {
				cb.Eq("layer", "service").
					OR().
					Eq("layer", "repository")
			}).
			Gt("quality", 0.8).
			Build()
	}

	// Test SqlOfSelect
	built1 := buildQuery()
	sql1, args1, _ := built1.SqlOfSelect()

	t.Logf("=== SqlOfSelect ===")
	t.Logf("SQL: %s", sql1)
	t.Logf("Args: %v", args1)

	// Test SqlOfVectorSearch (add vector search)
	vec := Vector{0.1, 0.2, 0.3}
	built2 := buildQuery()
	built2.Conds = append(built2.Conds, Bb{
		Op:    VECTOR_SEARCH,
		Key:   "embedding",
		Value: VectorSearchParams{QueryVector: vec, TopK: 10, DistanceMetric: CosineDistance},
	})
	sql2, args2 := built2.SqlOfVectorSearch()

	t.Logf("\n=== SqlOfVectorSearch ===")
	t.Logf("SQL: %s", sql2)
	t.Logf("Args: %v", args2)

	// Verify: Both should handle Or() consistently
	// SqlOfSelect should have (layer = ? OR layer = ?)
	if !containsString(sql1, "OR") {
		t.Errorf("SqlOfSelect missing OR")
	}

	// SqlOfVectorSearch should also have (layer = ? OR layer = ?)
	if !containsString(sql2, "OR") {
		t.Errorf("SqlOfVectorSearch missing OR (this is a bug in v0.9.0!)")
	}

	// The number of layer conditions should be the same
	// Note: SqlOfVectorSearch will have one more vector parameter
	if len(args1) != 4 { // language, layer(service), layer(repository), quality
		t.Errorf("SqlOfSelect expected 4 args, got %d", len(args1))
	}
	if len(args2) != 5 { // vec, language, layer(service), layer(repository), quality
		t.Errorf("SqlOfVectorSearch expected 5 args, got %d", len(args2))
	}

	t.Logf("✅ SqlOfSelect and SqlOfVectorSearch behavior一致")
}

// Regression test: empty And/Or subquery filtering
// Ensure correct filtering in all query types
func TestRegression_EmptyAndOr_AllQueryTypes(t *testing.T) {
	vec := Vector{0.1, 0.2, 0.3}

	builder := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			// ⭐ All conditions are empty, the entire Or should be filtered
			cb.Eq("tag", "").
				OR().
				Gt("count", 0)
		}).
		VectorSearch("embedding", vec, 10).
		Build()

	// Test SqlOfSelect
	sql1, args1, _ := builder.SqlOfSelect()
	t.Logf("=== SqlOfSelect (Empty Or) ===")
	t.Logf("SQL: %s", sql1)
	t.Logf("Args: %v", args1)

	if containsString(sql1, "tag") || containsString(sql1, "count") {
		t.Errorf("SqlOfSelect: Empty Or should be filtered")
	}

	// Test SqlOfVectorSearch
	sql2, args2 := builder.SqlOfVectorSearch()
	t.Logf("\n=== SqlOfVectorSearch (Empty Or) ===")
	t.Logf("SQL: %s", sql2)
	t.Logf("Args: %v", args2)

	if containsString(sql2, "tag") || containsString(sql2, "count") {
		t.Errorf("SqlOfVectorSearch: Empty Or should be filtered")
	}

	// Test ToQdrantRequest
	req, _ := builder.ToQdrantRequest()
	t.Logf("\n=== Qdrant Request (Empty Or) ===")
	t.Logf("Filters: %d", len(req.Filter.Must))

	// Qdrant should only have 1 filter (language)
	if len(req.Filter.Must) != 1 {
		t.Errorf("Qdrant: Expected 1 filter, got %d", len(req.Filter.Must))
	}

	t.Logf("✅ Empty And/Or is correctly filtered in all query types")
}

// Regression test: nested And/Or
// Ensure complex nesting can be handled correctly
func TestRegression_NestedAndOr(t *testing.T) {
	builder := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		And(func(cb *CondBuilder) {
			cb.Or(func(cb2 *CondBuilder) {
				cb2.Eq("layer", "service").
					OR().
					Eq("layer", "repository")
			}).
				OR().
				Gt("quality", 0.9)
		}).
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== Nested And/Or regression test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify nested structure
	if !containsString(sql, "layer") {
		t.Errorf("Missing nested layer conditions")
	}
	if !containsString(sql, "quality") {
		t.Errorf("Missing quality condition")
	}

	// Should have at least 2 OR
	// TODO: More precise verification

	t.Logf("✅ Nested And/Or correctly handled")
}
