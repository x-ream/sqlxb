# RAG 最佳实践指南

## 📋 概述

本文档介绍如何使用 xb 构建高性能的 RAG（Retrieval-Augmented Generation）应用。涵盖文档分块、向量存储、混合检索、重排序等关键技术。

## 🏗️ RAG 架构设计

### 典型 RAG 流程

```
用户问题 → Embedding → 向量检索 → 重排序 → 上下文增强 → LLM 生成
```

### xb 在 RAG 中的角色

```
┌─────────────────────────────────────────────────────────┐
│                    文档摄入流程                          │
└─────────────────────────────────────────────────────────┘
  原始文档 → 分块 → Embedding → xb.Insert() → Qdrant

┌─────────────────────────────────────────────────────────┐
│                    检索流程                              │
└─────────────────────────────────────────────────────────┘
  用户查询 → Embedding → xb.VectorSearch() 
           → 标量过滤 → 重排序 → 返回上下文
```

## 📦 数据模型设计

### 基础 Chunk 模型

```go
package models

import "time"

type DocumentChunk struct {
    ID        int64     `json:"id" db:"id"`
    DocID     *int64    `json:"doc_id" db:"doc_id"`           // 原文档ID（非主键，可为 nil）
    ChunkID   int       `json:"chunk_id" db:"chunk_id"`       // 块序号
    Content   string    `json:"content" db:"content"`         // 文本内容
    Embedding []float32 `json:"embedding" db:"embedding"`     // 向量
    
    // 元数据字段
    DocType   string    `json:"doc_type" db:"doc_type"`       // 文档类型
    Language  string    `json:"language" db:"language"`       // 语言
    Source    string    `json:"source" db:"source"`           // 来源
    Author    string    `json:"author" db:"author"`           // 作者
    Tags      string    `json:"tags" db:"tags"`               // 标签（JSON数组）
    
    // 层级信息
    Section   string    `json:"section" db:"section"`         // 章节
    Level     int       `json:"level" db:"level"`             // 层级
    
    // 统计信息
    TokenCount int      `json:"token_count" db:"token_count"` // Token数
    CharCount  int      `json:"char_count" db:"char_count"`   // 字符数
    
    // 时间戳
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### 扩展元数据模型

```go
// 用于需要复杂元数据的场景
type DocumentChunkV2 struct {
    ID         int64              `json:"id"`
    Content    string             `json:"content"`
    Embedding  []float32          `json:"embedding"`
    Metadata   map[string]interface{} `json:"metadata"` // 灵活的元数据
    CreatedAt  time.Time          `json:"created_at"`
}

// 元数据示例
type ChunkMetadata struct {
    DocID       *int64   `json:"doc_id"`  // 原文档ID（非主键，可为 nil）
    DocType     string   `json:"doc_type"`
    Title       string   `json:"title"`
    URL         string   `json:"url"`
    Author      string   `json:"author"`
    PublishDate string   `json:"publish_date"`
    Tags        []string `json:"tags"`
    Section     string   `json:"section"`
    PageNumber  int      `json:"page_number"`
    Language    string   `json:"language"`
}
```

## ✂️ 文档分块策略

### 固定长度分块

```go
func ChunkByFixedSize(text string, chunkSize int, overlap int) []string {
    var chunks []string
    runes := []rune(text)
    
    for i := 0; i < len(runes); i += chunkSize - overlap {
        end := i + chunkSize
        if end > len(runes) {
            end = len(runes)
        }
        chunks = append(chunks, string(runes[i:end]))
        if end == len(runes) {
            break
        }
    }
    
    return chunks
}

// 使用示例
chunks := ChunkByFixedSize(document, 500, 50) // 500字符，50字符重叠
```

### 语义分块

```go
func ChunkBySentence(text string, maxSentences int) []string {
    // 按句子分割
    sentences := strings.Split(text, "。")
    
    var chunks []string
    var currentChunk []string
    
    for _, sentence := range sentences {
        currentChunk = append(currentChunk, sentence)
        
        if len(currentChunk) >= maxSentences {
            chunks = append(chunks, strings.Join(currentChunk, "。")+"。")
            currentChunk = currentChunk[len(currentChunk)-1:] // 保留最后一句作为上下文
        }
    }
    
    if len(currentChunk) > 0 {
        chunks = append(chunks, strings.Join(currentChunk, "。")+"。")
    }
    
    return chunks
}
```

### 层级分块（推荐）

```go
type HierarchicalChunk struct {
    Level   int    // 0: 文档, 1: 章节, 2: 段落, 3: 句子
    Content string
    Parent  int64  // 父级 ID
}

func ChunkHierarchical(document string) []HierarchicalChunk {
    var chunks []HierarchicalChunk
    
    // Level 0: 全文摘要
    summary := generateSummary(document)
    chunks = append(chunks, HierarchicalChunk{
        Level:   0,
        Content: summary,
    })
    
    // Level 1: 章节
    sections := splitBySections(document)
    for _, section := range sections {
        chunks = append(chunks, HierarchicalChunk{
            Level:   1,
            Content: section,
            Parent:  0,
        })
        
        // Level 2: 段落
        paragraphs := splitByParagraphs(section)
        for _, para := range paragraphs {
            chunks = append(chunks, HierarchicalChunk{
                Level:   2,
                Content: para,
                Parent:  int64(len(chunks) - 1),
            })
        }
    }
    
    return chunks
}
```

## 🔍 向量检索策略

### 基础向量检索

```go
func BasicVectorSearch(query string, embeddingFunc func(string) ([]float32, error)) (map[string]interface{}, error) {
    // 生成查询向量
    queryVector, err := embeddingFunc(query)
    if err != nil {
        return nil, err
    }
    
    // 构建查询
    built := xb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector, 10).
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.ScoreThreshold(0.7)
        }).
        Build()

    return built.ToQdrantJSON()
}
```

### 混合检索（向量 + 标量）

```go
func HybridSearch(query string, filters SearchFilters, embeddingFunc func(string) ([]float32, error)) (map[string]interface{}, error) {
    queryVector, err := embeddingFunc(query)
    if err != nil {
        return nil, err
    }
    
    builder := xb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector)
    
    // 标量过滤
    // ⭐ xb 自动过滤 nil/0/空字符串/time.Time零值/空切片，直接传参
    built := builder.
        Eq("doc_type", filters.DocType).        // 自动过滤空字符串
        Eq("language", filters.Language).       // 自动过滤空字符串
        In("tags", filters.Tags...).            // 自动过滤空切片
        Gte("created_at", filters.AfterDate).   // 自动过滤零值
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.ScoreThreshold(0.65)
        }).
        Build()

    return built.ToQdrantJSON()
}

type SearchFilters struct {
    DocType   string
    Language  string
    Tags      []string
    AfterDate time.Time
}
```

### 多阶段检索

```go
func MultiStageSearch(query string) ([]DocumentChunk, error) {
    // 阶段1: 粗召回（宽松条件，多返回结果）
    built1 := xb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector, 100).
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.ScoreThreshold(0.5) // 较低阈值
        }).
        Build()

    stage1JSON, err := built1.ToQdrantJSON()
    if err != nil {
        return nil, err
    }
    
    // 执行查询（伪代码）
    stage1Results := executeQdrantQuery(stage1JSON)
    
    if err != nil {
        return nil, err
    }
    
    // 阶段2: 精排序（使用更复杂的模型）
    rerankedResults := rerankWithCrossEncoder(query, stage1Results)
    
    // 阶段3: 多样性控制
    diverseResults := applyMMR(rerankedResults, 0.7, 10)
    
    return diverseResults, nil
}
```

### 上下文扩展

```go
func SearchWithContext(query string, expandWindow int) ([]DocumentChunk, error) {
    // 先找到最相关的 chunks
    relevantChunks, err := BasicVectorSearch(query, embeddingFunc)
    if err != nil {
        return nil, err
    }
    
    var allChunks []DocumentChunk
    
    // 为每个相关 chunk 获取前后文
    for _, chunk := range relevantChunks {
        // 获取前面的 chunks
        prevChunks, _ := xb.Of(&DocumentChunk{}).
            Eq("doc_id", chunk.DocID).
            Gte("chunk_id", chunk.ChunkID-expandWindow).
            Lt("chunk_id", chunk.ChunkID).
            OrderBy("chunk_id", xb.ASC).
            Build()
        
        // 获取后面的 chunks
        nextChunks, _ := xb.Of(&DocumentChunk{}).
            Eq("doc_id", chunk.DocID).
            Gt("chunk_id", chunk.ChunkID).
            Lte("chunk_id", chunk.ChunkID+expandWindow).
            OrderBy("chunk_id", xb.ASC).
            Build()
        
        // 合并上下文
        allChunks = append(allChunks, prevChunks...)
        allChunks = append(allChunks, chunk)
        allChunks = append(allChunks, nextChunks...)
    }
    
    return allChunks, nil
}
```

## 🎯 重排序策略

### MMR (Maximal Marginal Relevance)

```go
func ApplyMMR(chunks []DocumentChunk, lambda float64, topK int) []DocumentChunk {
    if len(chunks) == 0 {
        return chunks
    }
    
    selected := []DocumentChunk{chunks[0]} // 先选择最相关的
    remaining := chunks[1:]
    
    for len(selected) < topK && len(remaining) > 0 {
        var bestIdx int
        var bestScore float64 = -1
        
        for i, candidate := range remaining {
            // MMR 分数 = λ * 相关性 - (1-λ) * 最大相似度
            relevance := candidate.Score
            maxSim := maxSimilarity(candidate, selected)
            
            mmrScore := lambda*relevance - (1-lambda)*maxSim
            
            if mmrScore > bestScore {
                bestScore = mmrScore
                bestIdx = i
            }
        }
        
        selected = append(selected, remaining[bestIdx])
        remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
    }
    
    return selected
}

func maxSimilarity(chunk DocumentChunk, selected []DocumentChunk) float64 {
    var maxSim float64
    for _, s := range selected {
        sim := cosineSimilarity(chunk.Embedding, s.Embedding)
        if sim > maxSim {
            maxSim = sim
        }
    }
    return maxSim
}
```

### Cross-Encoder 重排序

```go
// 使用更强大的模型对候选结果重新评分
func RerankWithCrossEncoder(query string, chunks []DocumentChunk, model CrossEncoderModel) []DocumentChunk {
    type scoredChunk struct {
        chunk DocumentChunk
        score float64
    }
    
    var scored []scoredChunk
    
    for _, chunk := range chunks {
        // 使用 Cross-Encoder 计算查询和文档的相关性
        score := model.Score(query, chunk.Content)
        scored = append(scored, scoredChunk{chunk: chunk, score: score})
    }
    
    // 按分数排序
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].score > scored[j].score
    })
    
    // 返回重排序后的结果
    var result []DocumentChunk
    for _, s := range scored {
        result = append(result, s.chunk)
    }
    
    return result
}
```

## 📊 完整 RAG 应用示例

### RAG 服务

```go
package rag

import (
    "context"
    "github.com/fndome/xb"
)

type RAGService struct {
    db            *sqlx.DB
    qdrantClient  *qdrant.Client
    embeddingFunc func(string) ([]float32, error)
    llmClient     *openai.Client
}

func NewRAGService(db *sqlx.DB, qdrant *qdrant.Client, embedFunc func(string) ([]float32, error), llm *openai.Client) *RAGService {
    return &RAGService{
        db:            db,
        qdrantClient:  qdrant,
        embeddingFunc: embedFunc,
        llmClient:     llm,
    }
}

// 完整的 RAG 查询流程
func (s *RAGService) Query(ctx context.Context, query string, options QueryOptions) (*RAGResponse, error) {
    // 1. 生成查询向量
    queryVector, err := s.embeddingFunc(query)
    if err != nil {
        return nil, fmt.Errorf("embedding error: %w", err)
    }
    
    // 2. 向量检索 + 标量过滤
    built := xb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector, options.TopK * 2).  // 粗召回
        Eq("language", options.Language).
        In("doc_type", options.DocTypes...).
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.ScoreThreshold(0.6)
        }).
        Build()

    qdrantJSON, err := built.ToQdrantJSON()
    
    if err != nil {
        return nil, err
    }
    
    // 3. 执行 Qdrant 查询
    searchResults, err := s.qdrantClient.Search(ctx, qdrantQuery)
    if err != nil {
        return nil, err
    }
    
    // 4. 重排序（MMR）
    chunks := parseChunks(searchResults)
    rerankedChunks := ApplyMMR(chunks, 0.7, options.TopK)
    
    // 5. 上下文扩展
    expandedContext := s.expandContext(ctx, rerankedChunks, 1)
    
    // 6. 构建 Prompt
    prompt := s.buildPrompt(query, expandedContext)
    
    // 7. 调用 LLM 生成回答
    answer, err := s.generateAnswer(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    return &RAGResponse{
        Answer:      answer,
        Sources:     rerankedChunks,
        TokensUsed:  calculateTokens(prompt, answer),
        SearchScore: averageScore(rerankedChunks),
    }, nil
}

func (s *RAGService) buildPrompt(query string, chunks []DocumentChunk) string {
    var contextParts []string
    for i, chunk := range chunks {
        contextParts = append(contextParts, fmt.Sprintf("[文档%d]\n%s\n", i+1, chunk.Content))
    }
    
    context := strings.Join(contextParts, "\n")
    
    return fmt.Sprintf(`基于以下文档内容回答用户问题。如果文档中没有相关信息，请明确说明。

文档内容:
%s

用户问题: %s

回答:`, context, query)
}

func (s *RAGService) generateAnswer(ctx context.Context, prompt string) (string, error) {
    resp, err := s.llmClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem,
                Content: "你是一个专业的文档助手，基于提供的文档内容回答问题。",
            },
            {
                Role:    openai.ChatMessageRoleUser,
                Content: prompt,
            },
        },
        Temperature: 0.3,
    })
    
    if err != nil {
        return "", err
    }
    
    return resp.Choices[0].Message.Content, nil
}

type QueryOptions struct {
    Language string
    DocTypes []string
    TopK     int
}

type RAGResponse struct {
    Answer      string
    Sources     []DocumentChunk
    TokensUsed  int
    SearchScore float64
}
```

### HTTP API

```go
package api

import (
    "encoding/json"
    "net/http"
)

type QueryRequest struct {
    Query    string   `json:"query"`
    Language string   `json:"language"`
    DocTypes []string `json:"doc_types"`
    TopK     int      `json:"top_k"`
}

func (h *Handler) HandleRAGQuery(w http.ResponseWriter, r *http.Request) {
    var req QueryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 默认值
    if req.TopK == 0 {
        req.TopK = 5
    }
    if req.Language == "" {
        req.Language = "zh"
    }
    
    // 执行 RAG 查询
    response, err := h.ragService.Query(r.Context(), req.Query, rag.QueryOptions{
        Language: req.Language,
        DocTypes: req.DocTypes,
        TopK:     req.TopK,
    })
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## 🎯 性能优化

### 1. 批量向量化

```go
func BatchEmbed(texts []string, batchSize int, embedFunc func([]string) ([][]float32, error)) ([][]float32, error) {
    var allEmbeddings [][]float32
    
    for i := 0; i < len(texts); i += batchSize {
        end := i + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        embeddings, err := embedFunc(batch)
        if err != nil {
            return nil, err
        }
        
        allEmbeddings = append(allEmbeddings, embeddings...)
    }
    
    return allEmbeddings, nil
}
```

### 2. 缓存策略

```go
type EmbeddingCache struct {
    cache *lru.Cache
}

func (c *EmbeddingCache) GetOrCompute(text string, computeFunc func(string) ([]float32, error)) ([]float32, error) {
    // 使用文本的 hash 作为 key
    key := hash(text)
    
    if cached, ok := c.cache.Get(key); ok {
        return cached.([]float32), nil
    }
    
    embedding, err := computeFunc(text)
    if err != nil {
        return nil, err
    }
    
    c.cache.Add(key, embedding)
    return embedding, nil
}
```

### 3. 异步索引

```go
func (s *RAGService) AsyncIndex(documents []Document) error {
    // 使用 worker pool 并行处理
    jobs := make(chan Document, len(documents))
    results := make(chan error, len(documents))
    
    // 启动 workers
    for w := 0; w < runtime.NumCPU(); w++ {
        go s.indexWorker(jobs, results)
    }
    
    // 发送任务
    for _, doc := range documents {
        jobs <- doc
    }
    close(jobs)
    
    // 收集结果
    for range documents {
        if err := <-results; err != nil {
            return err
        }
    }
    
    return nil
}

func (s *RAGService) indexWorker(jobs <-chan Document, results chan<- error) {
    for doc := range jobs {
        err := s.indexDocument(doc)
        results <- err
    }
}
```

## 📊 监控与评估

### 检索质量指标

```go
type RetrievalMetrics struct {
    Precision    float64 // 准确率
    Recall       float64 // 召回率
    MRR          float64 // Mean Reciprocal Rank
    NDCG         float64 // Normalized Discounted Cumulative Gain
    AvgLatency   time.Duration
}

func EvaluateRetrieval(queries []TestQuery) RetrievalMetrics {
    var metrics RetrievalMetrics
    
    for _, q := range queries {
        results := executeQuery(q.Query)
        
        // 计算各项指标
        metrics.Precision += calculatePrecision(results, q.RelevantDocs)
        metrics.Recall += calculateRecall(results, q.RelevantDocs)
        metrics.MRR += calculateMRR(results, q.RelevantDocs)
    }
    
    // 平均值
    n := float64(len(queries))
    metrics.Precision /= n
    metrics.Recall /= n
    metrics.MRR /= n
    
    return metrics
}
```

## 🎓 最佳实践总结

1. **分块策略**
   - 推荐 200-500 tokens/chunk
   - 使用 50-100 tokens 重叠
   - 考虑层级分块

2. **元数据设计**
   - 添加足够的元数据用于过滤
   - 保留文档层级信息
   - 记录分块统计信息

3. **检索优化**
   - 使用混合检索（向量+标量）
   - 多阶段检索（粗召回+精排序）
   - 应用 MMR 增加多样性

4. **性能调优**
   - 批量处理向量化
   - 使用缓存减少重复计算
   - 异步索引提高吞吐

5. **监控评估**
   - 定期评估检索质量
   - 监控查询延迟
   - 收集用户反馈

---

**下一步**: 查看 [LANGCHAIN_INTEGRATION.md](./LANGCHAIN_INTEGRATION.md) 了解如何集成 LangChain。

