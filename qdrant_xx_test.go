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

// 测试 QDRANT_XX 自定义参数
func TestQdrantX_CustomParam(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		VectorSearch("embedding", queryVector, 10).
		QdrantX(func(qx *QdrantBuilderX) {
			// ⭐ 用户自定义参数（未来 Qdrant 可能新增的功能）
			qx.X("quantization", map[string]interface{}{
				"rescore": true,
			}).
				X("prefetch", 100)
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== QDRANT_XX 自定义参数测试 ===\n%s", jsonStr)

	// 验证 JSON 包含自定义参数
	if !containsString(jsonStr, "quantization") {
		t.Errorf("Expected custom param 'quantization' in JSON")
	}

	if !containsString(jsonStr, "prefetch") {
		t.Errorf("Expected custom param 'prefetch' in JSON")
	}

	// 解析 JSON 验证
	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// 验证 quantization 参数
	if quantization, ok := reqMap["quantization"]; ok {
		if qMap, ok := quantization.(map[string]interface{}); ok {
			if rescore, ok := qMap["rescore"]; !ok || rescore != true {
				t.Errorf("Expected quantization.rescore = true")
			}
		} else {
			t.Errorf("Expected quantization to be map")
		}
	} else {
		t.Errorf("Expected quantization in reqMap")
	}

	// 验证 prefetch 参数
	if prefetch, ok := reqMap["prefetch"]; ok {
		if prefetchVal, ok := prefetch.(float64); !ok || int(prefetchVal) != 100 {
			t.Errorf("Expected prefetch = 100, got %v", prefetch)
		}
	} else {
		t.Errorf("Expected prefetch in reqMap")
	}

	t.Logf("✅ 自定义参数正确添加到 JSON")
}

// 测试 PostgreSQL 忽略 QDRANT_XX
func TestPostgreSQL_IgnoresQdrantXX(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.HnswEf(256).
				X("quantization", map[string]interface{}{
					"rescore": true,
				})
		}).
		Build()

	// PostgreSQL SQL
	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== PostgreSQL 忽略 QDRANT_XX ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ SQL 应该干净，不包含 quantization
	expectedSQL := "SELECT *, embedding <-> ? AS distance FROM code_vectors WHERE language = ? ORDER BY distance LIMIT 20"

	if sql != expectedSQL {
		t.Errorf("Expected clean SQL, got: %s", sql)
	}

	t.Logf("✅ PostgreSQL 正确忽略 QDRANT_XX 参数")
}

// 测试 X() 扩展点的实际用例
func TestQdrantX_RealWorldExample(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 实际场景：使用 Qdrant 的 Quantization 功能
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		QdrantX(func(qx *QdrantBuilderX) {
			// 标准参数
			qx.HnswEf(256).
				ScoreThreshold(0.8)

			// ⭐ 使用 X() 设置 Qdrant 的量化重打分功能
			// 这是 Qdrant 1.7+ 的新功能，sqlxb 可能还未封装
			qx.X("quantization", map[string]interface{}{
				"rescore":      true,
				"oversampling": 2.0,
			})

			// ⭐ 使用 X() 设置预取参数（性能优化）
			qx.X("lookup_from", map[string]interface{}{
				"collection": "code_metadata",
				"vector":     "meta_embedding",
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("=== 真实场景：量化重打分 + 预取 ===\n%s", jsonStr)

	// 验证包含自定义参数
	if !containsString(jsonStr, "quantization") {
		t.Errorf("Expected quantization param")
	}

	if !containsString(jsonStr, "lookup_from") {
		t.Errorf("Expected lookup_from param")
	}

	t.Logf("✅ X() 扩展点可以使用 Qdrant 的最新功能")
}
