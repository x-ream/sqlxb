package main

import (
	"testing"

	"github.com/fndome/xb"
)

// 注意：这些测试需要实际的 Qdrant 环境
// 可以通过 Docker 快速启动：
// docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant

func TestBuildQdrantSearchJSON(t *testing.T) {
	// 测试 xb 生成的 Qdrant JSON 是否正确
	queryVector := make([]float32, 768)
	for i := range queryVector {
		queryVector[i] = 0.1
	}

	built := xb.Of(&Document{}).
		VectorSearch("embedding", queryVector, 20).
		Eq("doc_type", "article").
		Eq("language", "zh").
		QdrantX(func(qx *xb.QdrantBuilderX) {
			qx.ScoreThreshold(0.8).
				HnswEf(128).
				WithVector(true)
		}).
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("Generated Qdrant JSON:\n%s", jsonStr)

	// 验证 JSON 包含关键字段
	if !containsString(jsonStr, "vector") {
		t.Error("JSON should contain 'vector' field")
	}
	if !containsString(jsonStr, "filter") {
		t.Error("JSON should contain 'filter' field")
	}
	if !containsString(jsonStr, "score_threshold") {
		t.Error("JSON should contain 'score_threshold' field")
	}
	if !containsString(jsonStr, "hnsw_ef") {
		t.Error("JSON should contain 'hnsw_ef' field")
	}
}

func TestBuildRecommendJSON(t *testing.T) {
	built := xb.Of(&Document{}).
		QdrantX(func(qx *xb.QdrantBuilderX) {
			qx.Recommend(func(rb *xb.RecommendBuilder) {
				rb.Positive(123, 456).
					Negative(789).
					Limit(20)
			})
		}).
		Build()

	jsonStr, err := built.ToQdrantRecommendJSON()
	if err != nil {
		t.Fatalf("ToQdrantRecommendJSON failed: %v", err)
	}
	t.Logf("Generated Recommend JSON:\n%s", jsonStr)

	// 验证 JSON 包含关键字段
	if !containsString(jsonStr, "positive") {
		t.Error("JSON should contain 'positive' field")
	}
	if !containsString(jsonStr, "negative") {
		t.Error("JSON should contain 'negative' field")
	}
	if !containsString(jsonStr, "limit") {
		t.Error("JSON should contain 'limit' field")
	}
}

func TestAutoFiltering(t *testing.T) {
	// 测试自动过滤
	var emptyDocType string
	var emptyLanguage string

	built := xb.Of(&Document{}).
		VectorSearch("embedding", make([]float32, 768), 10).
		Eq("doc_type", emptyDocType).  // 应被自动过滤
		Eq("language", emptyLanguage). // 应被自动过滤
		Build()

	jsonStr, err := built.ToQdrantJSON()
	if err != nil {
		t.Fatalf("ToQdrantJSON failed: %v", err)
	}

	t.Logf("JSON with auto-filtering:\n%s", jsonStr)

	// 验证空字符串条件被过滤
	if containsString(jsonStr, "doc_type") {
		t.Error("Empty doc_type should be filtered out")
	}
	if containsString(jsonStr, "language") {
		t.Error("Empty language should be filtered out")
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
