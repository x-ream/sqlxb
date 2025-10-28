package xb

import (
	"encoding/json"
	"testing"
)

// TestQdrantRecommend 测试 Recommend API
func TestQdrantRecommend(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Recommend(func(rb *RecommendBuilder) {
				rb.Positive(123, 456, 789)
				rb.Negative(111, 222)
				rb.Limit(20)
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantRecommendJSON()
	if err != nil {
		t.Fatalf("ToQdrantRecommendJSON failed: %v", err)
	}

	t.Logf("Recommend JSON:\n%s", jsonStr)

	// 验证 JSON 结构
	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证正样本
	positive, ok := reqMap["positive"].([]interface{})
	if !ok || len(positive) != 3 {
		t.Errorf("Expected 3 positive samples, got %v", positive)
	}

	// 验证负样本
	negative, ok := reqMap["negative"].([]interface{})
	if !ok || len(negative) != 2 {
		t.Errorf("Expected 2 negative samples, got %v", negative)
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

// TestQdrantRecommendWithQdrantParams 测试 Recommend + Qdrant 参数
func TestQdrantRecommendWithQdrantParams(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("category", "tech").
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Recommend(func(rb *RecommendBuilder) {
				rb.Positive(100, 200, 300).Limit(10)
			}).
				HnswEf(256).
				ScoreThreshold(0.8).
				WithVector(true)
		}).
		Build()

	jsonStr, err := built.ToQdrantRecommendJSON()
	if err != nil {
		t.Fatalf("ToQdrantRecommendJSON failed: %v", err)
	}

	t.Logf("Recommend with params JSON:\n%s", jsonStr)

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
	if !ok || threshold != 0.8 {
		t.Errorf("Expected score_threshold 0.8, got %v", threshold)
	}

	// 验证 with_vector
	withVector, ok := reqMap["with_vector"].(bool)
	if !ok || !withVector {
		t.Errorf("Expected with_vector true, got %v", withVector)
	}
}

// TestQdrantScroll 测试 Scroll API
func TestQdrantScroll(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Eq("status", "active").
		QdrantX(func(qx *QdrantBuilderX) {
			qx.ScrollID("scroll-12345-abcde-xyz")
		}).
		Build()

	jsonStr, err := built.ToQdrantScrollJSON()
	if err != nil {
		t.Fatalf("ToQdrantScrollJSON failed: %v", err)
	}

	t.Logf("Scroll JSON:\n%s", jsonStr)

	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证 scroll_id
	scrollID, ok := reqMap["scroll_id"].(string)
	if !ok || scrollID != "scroll-12345-abcde-xyz" {
		t.Errorf("Expected scroll_id 'scroll-12345-abcde-xyz', got %v", scrollID)
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

// TestQdrantRecommendOnlyPositive 测试只有正样本的推荐
func TestQdrantRecommendOnlyPositive(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		QdrantX(func(qx *QdrantBuilderX) {
			qx.Recommend(func(rb *RecommendBuilder) {
				rb.Positive(123, 456).Limit(15)
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantRecommendJSON()
	if err != nil {
		t.Fatalf("ToQdrantRecommendJSON failed: %v", err)
	}

	t.Logf("Recommend (positive only) JSON:\n%s", jsonStr)

	var reqMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &reqMap); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// 验证正样本
	positive, ok := reqMap["positive"].([]interface{})
	if !ok || len(positive) != 2 {
		t.Errorf("Expected 2 positive samples, got %v", positive)
	}

	// 验证负样本不存在或为空
	negative, _ := reqMap["negative"].([]interface{})
	if len(negative) > 0 {
		t.Errorf("Expected no negative samples, got %v", negative)
	}
}
