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

// 测试 PostgreSQL 自动忽略 Qdrant 专属配置
func TestPostgreSQL_IgnoresQdrantConfig(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 构建包含 Qdrant 专属配置的查询
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.HnswEf(256).
				ScoreThreshold(0.8).
				Exact(true).
				WithVector(true)
		}).
		Build()

	// 生成 PostgreSQL SQL
	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== PostgreSQL 忽略 Qdrant 配置 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ 验证：SQL 应该是干净的，不包含 Qdrant 专属配置
	expectedSQL := "SELECT *, embedding <-> ? AS distance FROM code_vectors WHERE language = ? ORDER BY distance LIMIT 20"

	if sql != expectedSQL {
		t.Errorf("Expected SQL:\n%s\nGot:\n%s", expectedSQL, sql)
	}

	// 应该只有 2 个参数：queryVector 和 "golang"
	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Logf("✅ PostgreSQL 正确忽略所有 Qdrant 专属配置")
}

// 测试 Qdrant 应用配置，PostgreSQL 忽略（同一查询）
func TestSameQuery_TwoBackends(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 构建查询
	builder := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Gt("quality_score", 0.8).
		VectorSearch("embedding", queryVector, 20).   // ⭐ 通用方法（外部）
		WithHashDiversity("semantic_hash").           // ⭐ 通用方法（外部）
		QdrantX(func(qx *QdrantBuilderX) {
			// ⭐ 只配置 Qdrant 专属参数
			qx.HnswEf(256).
				ScoreThreshold(0.75)
		})

	built := builder.Build()

	// ===== 后端 1: PostgreSQL =====
	t.Log("\n=== PostgreSQL 后端 ===")
	sql, args := built.SqlOfVectorSearch()
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))

	// ⭐ PostgreSQL 应该忽略 Qdrant 配置和多样性
	if !containsString(sql, "language") {
		t.Errorf("Expected language condition")
	}
	if !containsString(sql, "quality_score") {
		t.Errorf("Expected quality_score condition")
	}
	if !containsString(sql, "LIMIT 20") {
		t.Errorf("Expected LIMIT 20 (not 100)")
	}

	// ===== 后端 2: Qdrant =====
	t.Log("\n=== Qdrant 后端 ===")
	jsonStr, _ := built.ToQdrantJSON()
	t.Logf("JSON:\n%s", jsonStr)

	// ⭐ Qdrant 应该应用配置和多样性
	if !containsString(jsonStr, "hnsw_ef\": 256") {
		t.Errorf("Expected hnsw_ef: 256")
	}
	if !containsString(jsonStr, "score_threshold\": 0.75") {
		t.Errorf("Expected score_threshold: 0.75")
	}
	if !containsString(jsonStr, "\"limit\": 100") {
		t.Errorf("Expected limit: 100 (diversity)")
	}

	t.Log("\n=== 总结 ===")
	t.Logf("✅ 同一查询，PostgreSQL 忽略 Qdrant 配置")
	t.Logf("✅ 同一查询，Qdrant 应用所有配置")
	t.Logf("✅ 优雅降级正常工作")
}
