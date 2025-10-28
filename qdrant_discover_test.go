package xb

import (
	"encoding/json"
	"testing"
)

// TestQdrantDiscover 测试 Discover API
func TestQdrantDiscover(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Discover(func(db *DiscoverBuilder) {
				db.Context(101, 102, 103) // 用户浏览历史
				db.Limit(20)
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantDiscoverJSON()
	if err != nil {
		t.Fatalf("ToQdrantDiscoverJSON failed: %v", err)
	}

	t.Logf("Discover JSON:\n%s", jsonStr)

	// 验证 JSON 结构
	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证上下文
	context, ok := reqMap["context"].([]interface{})
	if !ok || len(context) != 3 {
		t.Errorf("Expected 3 context samples, got %v", context)
	}

	// 验证 limit
	limit, ok := reqMap["limit"].(float64)
	if !ok || int(limit) != 20 {
		t.Errorf("Expected limit 20, got %v", limit)
	}

	// 验证过滤器
	filter, ok := reqMap["filter"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected filter to exist")
	} else {
		must, _ := filter["must"].([]interface{})
		if len(must) != 1 {
			t.Errorf("Expected 1 must condition, got %d", len(must))
		}
	}
}

// TestQdrantDiscoverWithQdrantParams 测试 Discover + Qdrant 参数
func TestQdrantDiscoverWithQdrantParams(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("category", "tech").
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Discover(func(db *DiscoverBuilder) {
				db.Context(100, 200, 300, 400).Limit(15)
			}).
				HnswEf(256).
				ScoreThreshold(0.75).
				WithVector(true)
		}).
		Build()

	jsonStr, err := built.ToQdrantDiscoverJSON()
	if err != nil {
		t.Fatalf("ToQdrantDiscoverJSON failed: %v", err)
	}

	t.Logf("Discover with params JSON:\n%s", jsonStr)

	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证 HNSW EF
	params, ok := reqMap["params"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected params to exist")
	} else {
		hnswEf, _ := params["hnsw_ef"].(float64)
		if int(hnswEf) != 256 {
			t.Errorf("Expected hnsw_ef 256, got %v", hnswEf)
		}
	}

	// 验证 score_threshold
	threshold, ok := reqMap["score_threshold"].(float64)
	if !ok || threshold != 0.75 {
		t.Errorf("Expected score_threshold 0.75, got %v", threshold)
	}

	// 验证 with_vector
	withVector, ok := reqMap["with_vector"].(bool)
	if !ok || !withVector {
		t.Errorf("Expected with_vector true, got %v", withVector)
	}
}

// TestQdrantDiscoverSimple 测试简单的 Discover
func TestQdrantDiscoverSimple(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Discover(func(db *DiscoverBuilder) {
				db.Context(123, 456, 789).Limit(10)
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantDiscoverJSON()
	if err != nil {
		t.Fatalf("ToQdrantDiscoverJSON failed: %v", err)
	}

	t.Logf("Simple Discover JSON:\n%s", jsonStr)

	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证上下文
	context, ok := reqMap["context"].([]interface{})
	if !ok || len(context) != 3 {
		t.Errorf("Expected 3 context samples, got %v", context)
	}

	// 验证 limit
	limit, ok := reqMap["limit"].(float64)
	if !ok || int(limit) != 10 {
		t.Errorf("Expected limit 10, got %v", limit)
	}
}
