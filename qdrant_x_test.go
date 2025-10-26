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
	"encoding/json"
	"testing"
)

// 测试基础 QdrantX
func TestQdrantX_Basic(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.HnswEf(256).
				ScoreThreshold(0.8).
				WithVector(true)
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== QdrantX 基础测试 ===\n%s", jsonStr)

	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// 验证配置
	if req.Params.HnswEf != 256 {
		t.Errorf("Expected HnswEf 256, got %d", req.Params.HnswEf)
	}

	if req.ScoreThreshold == nil || *req.ScoreThreshold != 0.8 {
		t.Errorf("Expected ScoreThreshold 0.8, got %v", req.ScoreThreshold)
	}

	if !req.WithVector {
		t.Errorf("Expected WithVector true, got false")
	}

	t.Logf("✅ QdrantX 配置正确应用")
}

// 测试高精度模式
func TestQdrantX_HighPrecision(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		VectorSearch("embedding", queryVector, 10).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.HighPrecision() // ⭐ 高精度模式
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 高精度模式 ===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	if req.Params.HnswEf != 512 {
		t.Errorf("HighPrecision should set HnswEf=512, got %d", req.Params.HnswEf)
	}

	t.Logf("✅ 高精度模式：HnswEf = %d", req.Params.HnswEf)
}

// 测试高速模式
func TestQdrantX_HighSpeed(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		VectorSearch("embedding", queryVector, 10).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.HighSpeed() // ⭐ 高速模式
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 高速模式 ===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	if req.Params.HnswEf != 32 {
		t.Errorf("HighSpeed should set HnswEf=32, got %d", req.Params.HnswEf)
	}

	t.Logf("✅ 高速模式：HnswEf = %d", req.Params.HnswEf)
}

// 测试分页（使用 sqlxb 的 Paged()）⭐
func TestQdrantX_Paged(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 使用 sqlxb 的 Paged() 方法
	built := Of(&CodeVectorForQdrant{}).
		VectorSearch("embedding", queryVector, 20).
		Paged(func(pb *PageBuilder) {
			pb.Page(3).Rows(20) // ⭐ 第 3 页，每页 20 条
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 分页测试（第3页，每页20条）===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	// offset 应该是 (3-1) * 20 = 40
	if req.Offset != 40 {
		t.Errorf("Expected offset 40, got %d", req.Offset)
	}

	if req.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", req.Limit)
	}

	t.Logf("✅ sqlxb Paged() 正确转换为 Qdrant offset/limit")
}

// 测试精确搜索
func TestQdrantX_Exact(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		VectorSearch("embedding", queryVector, 10).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Exact(true) // ⭐ 精确搜索（不使用索引）
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 精确搜索 ===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	if !req.Params.Exact {
		t.Errorf("Expected Exact true, got false")
	}

	t.Logf("✅ 精确搜索模式启用")
}

// 测试组合：多样性 + 模板
func TestQdrantX_WithDiversity(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash"). // ⭐ 多样性
		QdrantX(func(qx *QdrantBuilderX) {
			qx.HighPrecision(). // ⭐ 高精度
						ScoreThreshold(0.75) // ⭐ 阈值
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 多样性 + 模板组合 ===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	// 验证多样性：limit 应该被扩大
	if req.Limit != 100 {
		t.Errorf("Expected limit 100 (diversity), got %d", req.Limit)
	}

	// 验证模板：HnswEf 和 ScoreThreshold
	if req.Params.HnswEf != 512 {
		t.Errorf("Expected HnswEf 512 (HighPrecision), got %d", req.Params.HnswEf)
	}

	if req.ScoreThreshold == nil || *req.ScoreThreshold != 0.75 {
		t.Errorf("Expected ScoreThreshold 0.75, got %v", req.ScoreThreshold)
	}

	t.Logf("✅ 多样性和模板配置都正确应用")
}

// 测试完整配置（通用方法在外部，Qdrant 专属在内部）⭐
func TestQdrantX_CompleteConfig(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	// ⭐ 推荐：通用方法在外部，Qdrant 专属在 QdrantX 内
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").                   // 通用条件
		Gt("quality_score", 0.8).                   // 通用条件
		VectorSearch("embedding", queryVector, 20). // ⭐ 通用向量检索（外部）
		WithHashDiversity("semantic_hash").         // ⭐ 通用多样性（外部）
		QdrantX(func(qx *QdrantBuilderX) {
			// ⭐ 只配置 Qdrant 专属参数
			qx.HnswEf(256).
				ScoreThreshold(0.85).
				WithVector(false)
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 完整配置（通用+专属分离）===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	// 验证向量检索
	if len(req.Vector) != 4 {
		t.Errorf("Expected vector length 4, got %d", len(req.Vector))
	}

	// 验证多样性（limit 扩大）
	if req.Limit != 100 {
		t.Errorf("Expected limit 100 (diversity), got %d", req.Limit)
	}

	// 验证模板配置
	if req.Params.HnswEf != 256 {
		t.Errorf("Expected HnswEf 256, got %d", req.Params.HnswEf)
	}

	if req.ScoreThreshold == nil || *req.ScoreThreshold != 0.85 {
		t.Errorf("Expected ScoreThreshold 0.85, got %v", req.ScoreThreshold)
	}

	if req.WithVector {
		t.Errorf("Expected WithVector false, got true")
	}

	t.Logf("✅ 通用方法和 Qdrant 专属配置正确分离和应用")
}

// 测试不使用 QdrantX（向后兼容）
func TestQdrantX_NotUsed(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 不使用 QdrantX
	built := Of(&CodeVectorForQdrant{}).
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 不使用 QdrantX（向后兼容）===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	// 应该使用默认值
	if req.Params.HnswEf != 128 {
		t.Errorf("Expected default HnswEf 128, got %d", req.Params.HnswEf)
	}

	if req.ScoreThreshold != nil {
		t.Errorf("Expected no ScoreThreshold, got %v", req.ScoreThreshold)
	}

	if req.WithVector {
		t.Errorf("Expected WithVector false, got true")
	}

	t.Logf("✅ 向后兼容：默认配置正确")
}
