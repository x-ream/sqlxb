package main

import (
	"context"
	"fmt"
	"strings"
)

// RAGService RAG 服务
type RAGService struct {
	repo     *ChunkRepository
	embedder EmbeddingService
	llm      LLMService
}

func NewRAGService(repo *ChunkRepository, embedder EmbeddingService, llm LLMService) *RAGService {
	return &RAGService{
		repo:     repo,
		embedder: embedder,
		llm:      llm,
	}
}

// Query RAG 查询
func (s *RAGService) Query(ctx context.Context, req RAGQueryRequest) (*RAGQueryResponse, error) {
	// 1. 将问题转换为向量
	queryVector, err := s.embedder.Embed(ctx, req.Question)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	// 2. 检索相关文档
	topK := 5 // 默认值
	if req.TopK != nil && *req.TopK > 0 {
		topK = *req.TopK
	}

	chunks, err := s.repo.VectorSearch(queryVector, req.DocType, req.Language, topK)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(chunks) == 0 {
		return &RAGQueryResponse{
			Answer:   "抱歉，没有找到相关文档。",
			Sources:  []*DocumentChunk{},
			Metadata: map[string]interface{}{"chunks_found": 0},
		}, nil
	}

	// 3. 构建 LLM 提示词
	prompt := s.buildPrompt(req.Question, chunks)

	// 4. 调用 LLM 生成答案
	answer, err := s.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm generation failed: %w", err)
	}

	return &RAGQueryResponse{
		Answer:  answer,
		Sources: chunks,
		Metadata: map[string]interface{}{
			"chunks_found": len(chunks),
			"top_k":        topK,
		},
	}, nil
}

// buildPrompt 构建 LLM 提示词
func (s *RAGService) buildPrompt(question string, chunks []*DocumentChunk) string {
	var sb strings.Builder

	sb.WriteString("请根据以下文档内容回答问题。\n\n")
	sb.WriteString("相关文档：\n")

	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("\n[文档 %d]\n", i+1))
		sb.WriteString(chunk.Content)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n问题：%s\n\n", question))
	sb.WriteString("请基于上述文档内容进行回答。如果文档中没有相关信息，请明确说明。")

	return sb.String()
}

// EmbeddingService 嵌入服务接口
type EmbeddingService interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// LLMService LLM 服务接口
type LLMService interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// MockEmbeddingService 模拟嵌入服务（用于演示）
type MockEmbeddingService struct{}

func (s *MockEmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	// 实际应用中应该调用真实的 Embedding API（如 OpenAI）
	// 这里返回模拟向量
	vec := make([]float32, 768)
	for i := range vec {
		vec[i] = 0.1
	}
	return vec, nil
}

// MockLLMService 模拟 LLM 服务（用于演示）
type MockLLMService struct{}

func (s *MockLLMService) Generate(ctx context.Context, prompt string) (string, error) {
	// 实际应用中应该调用真实的 LLM API（如 OpenAI, DeepSeek）
	return "这是基于检索文档生成的答案...", nil
}
