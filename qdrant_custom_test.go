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
	"encoding/json"
	"testing"
)

// ============================================================================
// Custom 接口测试
// ============================================================================

func TestQdrantCustom_ImplementsCustomInterface(t *testing.T) {
	// ✅ 验证 QdrantCustom 实现了 Custom 接口
	var _ Custom = (*QdrantCustom)(nil)
	t.Log("✅ QdrantCustom implements Custom interface")
}

// ============================================================================
// JsonOfSelect 统一接口测试
// ============================================================================

func TestJsonOfSelect_WithQdrantCustom(t *testing.T) {
	// 测试通过 Custom 自动调用 Qdrant 逻辑
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of("code_vectors").
		Custom(NewQdrantCustom()).
		VectorSearch("embedding", queryVector, 20).
		Eq("language", "golang").
		Build()

	// ⭐ 统一接口：JsonOfSelect()
	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	// 验证 JSON 结构
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证基本字段
	if result["limit"] != float64(20) {
		t.Errorf("limit = %v, want 20", result["limit"])
	}

	if result["with_payload"] != true {
		t.Errorf("with_payload = %v, want true", result["with_payload"])
	}

	t.Logf("✅ JsonOfSelect with QdrantCustom works")
}

// ============================================================================
// Custom 预设模式测试
// ============================================================================

func TestQdrantCustom_PresetModes(t *testing.T) {
	tests := []struct {
		name         string
		custom       *QdrantCustom
		expectHnswEf int
		expectScore  float32
		expectVector bool
	}{
		{"Default", NewQdrantCustom(), 128, 0.0, false}, // 默认不返回向量（节省带宽）
		{"CustomHnswEf", func() *QdrantCustom {
			c := NewQdrantCustom()
			c.DefaultHnswEf = 512
			c.DefaultScoreThreshold = 0.85
			return c
		}(), 512, 0.85, false}, // 继承默认值 false
		{"CustomHighSpeed", func() *QdrantCustom {
			c := NewQdrantCustom()
			c.DefaultHnswEf = 32
			c.DefaultScoreThreshold = 0.5
			c.DefaultWithVector = false
			return c
		}(), 32, 0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.custom.DefaultHnswEf != tt.expectHnswEf {
				t.Errorf("DefaultHnswEf = %d, want %d", tt.custom.DefaultHnswEf, tt.expectHnswEf)
			}

			if tt.custom.DefaultScoreThreshold != tt.expectScore {
				t.Errorf("DefaultScoreThreshold = %f, want %f", tt.custom.DefaultScoreThreshold, tt.expectScore)
			}

			if tt.custom.DefaultWithVector != tt.expectVector {
				t.Errorf("DefaultWithVector = %v, want %v", tt.custom.DefaultWithVector, tt.expectVector)
			}

			t.Logf("✅ %s mode: HnswEf=%d, ScoreThreshold=%f, WithVector=%v",
				tt.name, tt.custom.DefaultHnswEf, tt.custom.DefaultScoreThreshold, tt.custom.DefaultWithVector)
		})
	}
}

// ============================================================================
// 向后兼容性测试
// ============================================================================

func TestBackwardCompatibility_ToQdrantJSON(t *testing.T) {
	// ✅ 旧 API 仍然可用
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of("code_vectors").
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	t.Logf("✅ Old API (ToQdrantJSON) still works")
}

// ============================================================================
// Custom 为 nil 的情况
// ============================================================================

func TestJsonOfSelect_CustomIsNil(t *testing.T) {
	// 没有设置 Custom 的情况
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of("code_vectors").
		VectorSearch("embedding", queryVector, 10).
		Build()

	// built.Custom 应该是 nil
	if built.Custom != nil {
		t.Fatalf("Custom should be nil, got %T", built.Custom)
	}

	// JsonOfSelect 应该返回错误
	_, err := built.JsonOfSelect()
	if err == nil {
		t.Fatalf("JsonOfSelect should return error when Custom is nil")
	}

	t.Logf("✅ JsonOfSelect correctly returns error when Custom is nil: %v", err)
}

// ============================================================================
// Custom 切换测试（运行时切换）
// ============================================================================

func TestCustomSwitch_RuntimeSelection(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 模拟根据配置选择不同的 Custom
	// ⭐ 用户可以根据需要手动配置
	customs := map[string]*QdrantCustom{
		"default": NewQdrantCustom(),
		"high_precision": func() *QdrantCustom {
			c := NewQdrantCustom()
			c.DefaultHnswEf = 512
			c.DefaultScoreThreshold = 0.85
			return c
		}(),
		"high_speed": func() *QdrantCustom {
			c := NewQdrantCustom()
			c.DefaultHnswEf = 32
			c.DefaultScoreThreshold = 0.5
			c.DefaultWithVector = false
			return c
		}(),
	}

	for name, custom := range customs {
		t.Run(name, func(t *testing.T) {
			built := Of("code_vectors").
				Custom(custom).
				VectorSearch("embedding", queryVector, 10).
				Build()

			jsonStr, err := built.JsonOfSelect()
			if err != nil {
				t.Fatalf("JsonOfSelect failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			t.Logf("✅ %s mode works", name)
		})
	}
}

// ============================================================================
// 自定义配置测试
// ============================================================================

func TestQdrantCustom_CustomConfiguration(t *testing.T) {
	// 用户自定义配置
	customConfig := &QdrantCustom{
		DefaultHnswEf:         256,
		DefaultScoreThreshold: 0.75,
		DefaultWithVector:     true,
	}

	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of("code_vectors").
		Custom(customConfig).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	t.Logf("✅ Custom configuration applied")
}
