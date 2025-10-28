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
package xb

import (
	"encoding/json"
	"testing"
)

// 测试 nil/0 自动过滤
func TestQdrant_NilZeroFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 构建查询，包含 nil/0 值
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang"). // ✅ 有效
		Eq("category", "").       // ⭐ 应被过滤（空字符串）
		Gt("score", 0.8).         // ✅ 有效
		Gt("rank", 0).            // ⭐ 应被过滤（0）
		Lt("complexity", 100).    // ✅ 有效
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== nil/0 过滤测试 ===\n%s", jsonStr)

	// 解析 JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ 关键验证：filter 中应该只有 3 个条件
	// language=golang, score>0.8, complexity<100
	// category="" 和 rank=0 应该被过滤掉
	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	} else {
		mustCount := len(req.Filter.Must)

		// 应该只有 3 个条件（不包括 category="" 和 rank=0）
		if mustCount != 3 {
			t.Errorf("Expected 3 must conditions (filtered out empty string and 0), got %d", mustCount)
			t.Logf("Conditions:")
			for i, cond := range req.Filter.Must {
				t.Logf("  %d: key=%s, match=%v, range=%v", i, cond.Key, cond.Match, cond.Range)
			}
		} else {
			t.Logf("✅ nil/0 过滤成功：只有 3 个有效条件")
		}

		// 验证不包含 category 和 rank
		for _, cond := range req.Filter.Must {
			if cond.Key == "category" {
				t.Errorf("category with empty value should be filtered")
			}
			if cond.Key == "rank" {
				t.Errorf("rank with value 0 should be filtered")
			}
		}
	}
}

// 测试全部为 nil/0 的情况
func TestQdrant_AllNilZero(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 所有条件都是 nil/0
	built := Of(&CodeVectorForQdrant{}).
		Eq("category", "").
		Gt("rank", 0).
		Lt("count", 0).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 全部 nil/0 测试 ===\n%s", jsonStr)

	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// filter 应该为空或没有 must 条件
	if req.Filter != nil && len(req.Filter.Must) > 0 {
		t.Errorf("Expected no filter conditions, got %d", len(req.Filter.Must))
	} else {
		t.Logf("✅ 所有 nil/0 条件被过滤")
	}
}

// 测试 PostgreSQL 也有相同的过滤
func TestPostgreSQL_NilZeroFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 相同的查询
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Eq("category", ""). // ⭐ 应被过滤
		Gt("score", 0.8).
		Gt("rank", 0). // ⭐ 应被过滤
		VectorSearch("embedding", queryVector, 20).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== PostgreSQL nil/0 过滤测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// SQL 不应该包含 category 和 rank
	if containsString(sql, "category") {
		t.Errorf("category with empty value should be filtered from SQL")
	}
	if containsString(sql, "rank") {
		t.Errorf("rank with value 0 should be filtered from SQL")
	}

	// 应该只有 2 个参数：queryVector 和 "golang"
	// (加上 0.8 是 3 个)
	if len(args) != 3 {
		t.Errorf("Expected 3 args (vector, language, score), got %d: %v", len(args), args)
	} else {
		t.Logf("✅ PostgreSQL nil/0 过滤成功")
	}
}
