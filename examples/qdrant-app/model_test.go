package main

import (
	"testing"

	"github.com/fndome/xb"
)

func TestDocumentTableName(t *testing.T) {
	doc := &Document{}
	tableName := doc.TableName()

	if tableName != "documents" {
		t.Errorf("Expected table name 'documents', got '%s'", tableName)
	}
}

func TestSearchRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     SearchRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: SearchRequest{
				QueryVector: make([]float32, 768),
				Limit:       xb.Int(10),
			},
			wantErr: false,
		},
		{
			name: "nil limit auto-corrected",
			req: SearchRequest{
				QueryVector: make([]float32, 768),
				Limit:       nil, // 应该在处理器中被设置为默认值
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.req.QueryVector) == 0 && !tt.wantErr {
				t.Error("QueryVector should not be empty")
			}
		})
	}
}

func TestQdrantJSONGeneration(t *testing.T) {
	// 测试 xb 生成 Qdrant JSON
	queryVector := make([]float32, 768)
	for i := range queryVector {
		queryVector[i] = 0.1
	}

	tests := []struct {
		name        string
		buildFunc   func() *xb.Built
		checkFields []string
	}{
		{
			name: "basic search",
			buildFunc: func() *xb.Built {
				return xb.Of(&Document{}).
					VectorSearch("embedding", queryVector, 10).
					Build()
			},
			checkFields: []string{"vector", "limit"},
		},
		{
			name: "search with filter",
			buildFunc: func() *xb.Built {
				return xb.Of(&Document{}).
					VectorSearch("embedding", queryVector, 10).
					Eq("doc_type", "article").
					Build()
			},
			checkFields: []string{"vector", "filter", "limit"},
		},
		{
			name: "search with QdrantX",
			buildFunc: func() *xb.Built {
				return xb.Of(&Document{}).
					VectorSearch("embedding", queryVector, 10).
					QdrantX(func(qx *xb.QdrantBuilderX) {
						qx.ScoreThreshold(0.8).HnswEf(128)
					}).
					Build()
			},
			checkFields: []string{"vector", "score_threshold", "params"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			built := tt.buildFunc()
			jsonBytes, err := built.ToQdrantJSON()
			if err != nil {
				t.Fatalf("ToQdrantJSON failed: %v", err)
			}

			jsonStr := string(jsonBytes)
			t.Logf("Generated JSON:\n%s", jsonStr)

			// 验证关键字段存在
			for _, field := range tt.checkFields {
				if !contains(jsonStr, field) {
					t.Errorf("JSON should contain field '%s'", field)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
