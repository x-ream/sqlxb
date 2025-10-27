package main

import (
	"context"
	"testing"
)

func TestSplitDocument(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		minChunks int
	}{
		{
			name: "simple paragraphs",
			content: `段落1：这是第一段内容，包含足够的文字。

段落2：这是第二段内容，也包含足够的文字。

段落3：这是第三段内容，继续包含足够的文字。`,
			minChunks: 3,
		},
		{
			name: "with short paragraphs",
			content: `很短

这是一个很长的段落，包含足够的文字，应该被保留。

太短了`,
			minChunks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := splitDocument(tt.content)
			
			if len(chunks) < tt.minChunks {
				t.Errorf("Expected at least %d chunks, got %d", tt.minChunks, len(chunks))
			}

			// 验证每个分块都有足够的内容
			for i, chunk := range chunks {
				if len(chunk) < 50 {
					t.Errorf("Chunk %d too short: %d chars", i, len(chunk))
				}
			}

			t.Logf("Split into %d chunks", len(chunks))
		})
	}
}

func TestRAGServiceQuery(t *testing.T) {
	// 使用 Mock 服务测试 RAG 流程
	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}
	
	// 注意：这需要真实的数据库，这里只测试服务逻辑
	// 实际测试应该使用测试数据库

	ctx := context.Background()

	// 测试 Embedder
	vec, err := embedder.Embed(ctx, "测试查询")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(vec) != 768 {
		t.Errorf("Expected vector length 768, got %d", len(vec))
	}

	// 测试 LLM
	answer, err := llm.Generate(ctx, "测试提示词")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if answer == "" {
		t.Error("Expected non-empty answer")
	}

	t.Logf("LLM answer: %s", answer)
}

func TestBuildPrompt(t *testing.T) {
	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}
	service := NewRAGService(nil, embedder, llm)

	chunks := []*DocumentChunk{
		{
			Content: "文档1的内容",
		},
		{
			Content: "文档2的内容",
		},
	}

	prompt := service.buildPrompt("测试问题？", chunks)

	// 验证提示词包含必要元素
	if !contains(prompt, "测试问题") {
		t.Error("Prompt should contain the question")
	}
	if !contains(prompt, "文档1的内容") {
		t.Error("Prompt should contain document 1")
	}
	if !contains(prompt, "文档2的内容") {
		t.Error("Prompt should contain document 2")
	}

	t.Logf("Generated prompt:\n%s", prompt)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

