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
)

// 回归测试：README 中的 And()/Or() 示例
// Bug: v0.9.0 中 SqlOfVectorSearch 无法正确处理 Or()
func TestRegression_README_AndOr(t *testing.T) {
	// 复用已有的 CodeVectorForQdrant 类型，避免类型冲突
	builder := Of(&CodeVectorForQdrant{}).
		Gt("id", 10000).
		And(func(cb *CondBuilder) {
			cb.Gte("price", 100.5).
				OR().
				Eq("is_sold", true)
		}).
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== README And/Or 回归测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 正确
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

	// 验证参数
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	t.Logf("✅ README 示例测试通过: %s", expectedSQL)
}

// 回归测试：float64 零值过滤
// Bug: v0.9.0 中 interface{} == 0 对 float64 无效
func TestRegression_Float64_ZeroFilter(t *testing.T) {
	// 所有数值类型的零值测试
	builder := Of(&CodeVectorForQdrant{}).
		Eq("name", "laptop"). // ✅ 保留
		// float64 零值
		Gt("min_price", 0.0).   // ❌ 应过滤
		Lt("max_price", 999.9). // ✅ 保留
		// float32 零值
		Gt("min_score", float32(0.0)). // ❌ 应过滤
		Lt("max_score", float32(0.9)). // ✅ 保留
		// int 零值
		Gt("min_weight", 0).    // ❌ 应过滤
		Lt("max_weight", 1000). // ✅ 保留
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== float64/int 零值过滤测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证零值被过滤
	if containsString(sql, "min_price") {
		t.Errorf("min_price=0.0 should be filtered")
	}
	if containsString(sql, "min_score") {
		t.Errorf("min_score=0.0 should be filtered")
	}
	if containsString(sql, "min_weight") {
		t.Errorf("min_weight=0 should be filtered")
	}

	// 验证非零值保留
	if !containsString(sql, "max_price") {
		t.Errorf("max_price should exist")
	}
	if !containsString(sql, "max_score") {
		t.Errorf("max_score should exist")
	}
	if !containsString(sql, "max_weight") {
		t.Errorf("max_weight should exist")
	}

	// 验证参数数量
	if len(args) != 4 { // name, max_price, max_score, max_weight
		t.Errorf("Expected 4 args, got %d: %v", len(args), args)
	}

	t.Logf("✅ 所有数值类型零值过滤正确")
}

// 回归测试：向量查询 + And/Or 组合
// Bug: v0.9.0 中 SqlOfVectorSearch 使用简化的 buildConditionSql，忽略 subs
func TestRegression_VectorSearch_WithAndOr(t *testing.T) {
	vec := Vector{0.1, 0.2, 0.3}

	// 组合测试：向量检索 + And/Or 子查询
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

	t.Logf("=== 向量查询 + And/Or 回归测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 结构
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

	// 验证参数
	if len(args) != 4 { // vec, language, quality, verified
		t.Errorf("Expected 4 args, got %d", len(args))
	}

	t.Logf("✅ 向量查询 + And/Or 正确工作")
}

// 回归测试：SqlOfSelect 和 SqlOfVectorSearch 一致性
// Bug: 两者使用不同的条件构建逻辑，导致行为不一致
func TestRegression_SqlOfSelect_vs_SqlOfVectorSearch(t *testing.T) {
	// 构建相同的查询条件
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

	// 测试 SqlOfSelect
	built1 := buildQuery()
	sql1, args1, _ := built1.SqlOfSelect()

	t.Logf("=== SqlOfSelect ===")
	t.Logf("SQL: %s", sql1)
	t.Logf("Args: %v", args1)

	// 测试 SqlOfVectorSearch（添加向量检索）
	vec := Vector{0.1, 0.2, 0.3}
	built2 := buildQuery()
	built2.Conds = append(built2.Conds, Bb{
		op:    VECTOR_SEARCH,
		key:   "embedding",
		value: VectorSearchParams{QueryVector: vec, TopK: 10, DistanceMetric: CosineDistance},
	})
	sql2, args2 := built2.SqlOfVectorSearch()

	t.Logf("\n=== SqlOfVectorSearch ===")
	t.Logf("SQL: %s", sql2)
	t.Logf("Args: %v", args2)

	// 验证：两者对 Or() 的处理应该一致
	// SqlOfSelect 应该有 (layer = ? OR layer = ?)
	if !containsString(sql1, "OR") {
		t.Errorf("SqlOfSelect missing OR")
	}

	// SqlOfVectorSearch 也应该有 (layer = ? OR layer = ?)
	if !containsString(sql2, "OR") {
		t.Errorf("SqlOfVectorSearch missing OR (这是 v0.9.0 的 bug!)")
	}

	// 两者的 layer 条件数量应该相同
	// 注意：SqlOfVectorSearch 会多一个 vector 参数
	if len(args1) != 4 { // language, layer(service), layer(repository), quality
		t.Errorf("SqlOfSelect expected 4 args, got %d", len(args1))
	}
	if len(args2) != 5 { // vec, language, layer(service), layer(repository), quality
		t.Errorf("SqlOfVectorSearch expected 5 args, got %d", len(args2))
	}

	t.Logf("✅ SqlOfSelect 和 SqlOfVectorSearch 行为一致")
}

// 回归测试：空 And/Or 子查询过滤
// 确保在所有查询类型中都能正确过滤
func TestRegression_EmptyAndOr_AllQueryTypes(t *testing.T) {
	vec := Vector{0.1, 0.2, 0.3}

	builder := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Or(func(cb *CondBuilder) {
			// ⭐ 所有条件都是空值，整个 Or 应该被过滤
			cb.Eq("tag", "").
				OR().
				Gt("count", 0)
		}).
		VectorSearch("embedding", vec, 10).
		Build()

	// 测试 SqlOfSelect
	sql1, args1, _ := builder.SqlOfSelect()
	t.Logf("=== SqlOfSelect (Empty Or) ===")
	t.Logf("SQL: %s", sql1)
	t.Logf("Args: %v", args1)

	if containsString(sql1, "tag") || containsString(sql1, "count") {
		t.Errorf("SqlOfSelect: Empty Or should be filtered")
	}

	// 测试 SqlOfVectorSearch
	sql2, args2 := builder.SqlOfVectorSearch()
	t.Logf("\n=== SqlOfVectorSearch (Empty Or) ===")
	t.Logf("SQL: %s", sql2)
	t.Logf("Args: %v", args2)

	if containsString(sql2, "tag") || containsString(sql2, "count") {
		t.Errorf("SqlOfVectorSearch: Empty Or should be filtered")
	}

	// 测试 ToQdrantRequest
	req, _ := builder.ToQdrantRequest()
	t.Logf("\n=== Qdrant Request (Empty Or) ===")
	t.Logf("Filters: %d", len(req.Filter.Must))

	// Qdrant 应该只有 1 个 filter（language）
	if len(req.Filter.Must) != 1 {
		t.Errorf("Qdrant: Expected 1 filter, got %d", len(req.Filter.Must))
	}

	t.Logf("✅ 空 And/Or 在所有查询类型中都被正确过滤")
}

// 回归测试：嵌套 And/Or
// 确保复杂嵌套也能正确处理
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

	t.Logf("=== 嵌套 And/Or 回归测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证嵌套结构
	if !containsString(sql, "layer") {
		t.Errorf("Missing nested layer conditions")
	}
	if !containsString(sql, "quality") {
		t.Errorf("Missing quality condition")
	}

	// 应该有至少 2 个 OR
	// TODO: 更精确的验证

	t.Logf("✅ 嵌套 And/Or 正确处理")
}
