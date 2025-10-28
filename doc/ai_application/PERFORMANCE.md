# AI 应用性能优化指南

## 📋 概述

本文档介绍在 AI 应用场景下如何优化 xb 的性能，涵盖向量检索、RAG 应用和大规模部署。

## 🎯 性能指标

### 关键指标

- **查询延迟 (P95)**: < 100ms
- **吞吐量 (QPS)**: > 500
- **向量检索精度**: > 0.90 (Recall@10)
- **内存使用**: < 2GB (100万文档)

## 🚀 向量检索优化

### 1. 索引优化

```go
// Qdrant 索引配置
type QdrantConfig struct {
    // HNSW 参数
    HNSWConfig HNSWConfig `json:"hnsw_config"`
    // 量化配置（降低内存使用）
    QuantizationConfig *QuantizationConfig `json:"quantization_config,omitempty"`
}

type HNSWConfig struct {
    M              int `json:"m"`                // 16-64, 越大越精确但越慢
    EfConstruct    int `json:"ef_construct"`    // 100-512
    FullScanThreshold int `json:"full_scan_threshold"` // 10000
}

// 推荐配置
var RecommendedConfig = QdrantConfig{
    HNSWConfig: HNSWConfig{
        M:              32,    // 平衡性能和精度
        EfConstruct:    256,   // 构建时的搜索范围
        FullScanThreshold: 10000,
    },
    QuantizationConfig: &QuantizationConfig{
        Scalar: &ScalarQuantization{
            Type: "int8",  // 减少75%内存
        },
    },
}
```

### 2. 批量向量化

```go
// 批量生成 embedding
func BatchEmbed(texts []string, batchSize int) ([][]float32, error) {
    var results [][]float32
    
    for i := 0; i < len(texts); i += batchSize {
        end := min(i+batchSize, len(texts))
        batch := texts[i:end]
        
        // 并行调用 embedding API
        embeddings, err := embeddingClient.EmbedBatch(batch)
        if err != nil {
            return nil, err
        }
        
        results = append(results, embeddings...)
    }
    
    return results, nil
}

// 推荐 batch size: 16-32 (OpenAI), 64-128 (本地模型)
```

### 3. 缓存策略

```go
type EmbeddingCache struct {
    cache *ristretto.Cache
    ttl   time.Duration
}

func NewEmbeddingCache(maxSize int64, ttl time.Duration) *EmbeddingCache {
    cache, _ := ristretto.NewCache(&ristretto.Config{
        NumCounters: maxSize * 10,
        MaxCost:     maxSize,
        BufferItems: 64,
    })
    
    return &EmbeddingCache{cache: cache, ttl: ttl}
}

func (c *EmbeddingCache) GetOrCompute(text string, fn func(string) ([]float32, error)) ([]float32, error) {
    key := hashText(text)
    
    if val, found := c.cache.Get(key); found {
        return val.([]float32), nil
    }
    
    embedding, err := fn(text)
    if err != nil {
        return nil, err
    }
    
    c.cache.SetWithTTL(key, embedding, 1, c.ttl)
    return embedding, nil
}

// 使用示例
cache := NewEmbeddingCache(1000, 1*time.Hour)
embedding, _ := cache.GetOrCompute(query, embeddingFunc)
```

## 📊 查询优化

### 1. 合理设置 Top-K

```go
// ❌ 不推荐：直接返回大量结果
built1 := xb.Of(&Doc{}).
    VectorSearch("embedding", vector, 100).
    Build()
json1, _ := built1.ToQdrantJSON()  // 返回太多

// ✅ 推荐：分阶段获取
// 阶段1: 粗召回
built2 := xb.Of(&Doc{}).
    VectorSearch("embedding", vector, 50).
    Build()
json2, _ := built2.ToQdrantJSON()

// 执行查询，然后阶段2: 精排序 + MMR
stage1Results := executeQdrantQuery(json2)
stage2 := rerankAndDiversify(stage1Results, 10)
```

### 2. 使用过滤减少计算量

```go
// ❌ 先检索再过滤
allResults := vectorSearch(vector, 1000)
filtered := filterByDate(allResults, last7Days)  // 浪费

// ✅ 先过滤再检索
built := xb.Of(&Doc{}).
    VectorSearch("embedding", vector, 20).
    Gte("created_at", last7Days).  // 减少候选集
    Build()
json, _ := built.ToQdrantJSON()
```

### 3. 并行查询

```go
func ParallelSearch(queries []string) ([][]Document, error) {
    results := make([][]Document, len(queries))
    errs := make([]error, len(queries))
    
    var wg sync.WaitGroup
    for i, q := range queries {
        wg.Add(1)
        go func(idx int, query string) {
            defer wg.Done()
            
            vector, _ := embedText(query)
            docs, err := executeSearch(vector)
            
            results[idx] = docs
            errs[idx] = err
        }(i, q)
    }
    
    wg.Wait()
    
    // 检查错误
    for _, err := range errs {
        if err != nil {
            return nil, err
        }
    }
    
    return results, nil
}
```

## 🎯 RAG 应用优化

### 1. 文档分块优化

```go
// 根据 Token 限制动态分块
func SmartChunk(text string, maxTokens int, overlap int) []string {
    tokens := tokenize(text)
    
    var chunks []string
    for i := 0; i < len(tokens); i += maxTokens - overlap {
        end := min(i+maxTokens, len(tokens))
        chunk := detokenize(tokens[i:end])
        chunks = append(chunks, chunk)
        
        if end == len(tokens) {
            break
        }
    }
    
    return chunks
}

// 推荐参数
// - maxTokens: 256-512 (平衡上下文和精度)
// - overlap: 50-100 (保持连贯性)
```

### 2. 重排序优化

```go
// 使用轻量级模型快速重排序
func FastRerank(query string, docs []Document, topK int) []Document {
    // 1. 使用 bi-encoder 粗排（已有的向量相似度）
    // 2. 只对 top 20-30 使用 cross-encoder 精排
    
    if len(docs) <= topK {
        return docs
    }
    
    // 只对候选集使用重排序模型
    candidates := docs[:min(30, len(docs))]
    
    scores := make([]float64, len(candidates))
    for i, doc := range candidates {
        scores[i] = crossEncoderScore(query, doc.Content)
    }
    
    // 排序并返回 top-K
    sorted := sortByScore(candidates, scores)
    return sorted[:topK]
}
```

### 3. 流式生成

```go
// 边检索边生成，减少首字节延迟
func StreamingRAG(query string, responseChan chan<- string) {
    // 1. 快速返回简单回答
    responseChan <- "正在搜索相关文档...\n"
    
    // 2. 执行检索
    docs := search(query)
    responseChan <- fmt.Sprintf("找到 %d 个相关文档\n", len(docs))
    
    // 3. 流式生成答案
    prompt := buildPrompt(query, docs)
    
    for chunk := range llmClient.StreamCompletion(prompt) {
        responseChan <- chunk
    }
    
    close(responseChan)
}
```

## 🏗️ 架构优化

### 1. 连接池

```go
var (
    qdrantPool *QdrantPool
    dbPool     *pgxpool.Pool
)

func InitializePools() error {
    // Qdrant 连接池
    qdrantPool = NewQdrantPool(&QdrantPoolConfig{
        MinConns: 10,
        MaxConns: 100,
        MaxIdleTime: 5 * time.Minute,
    })
    
    // PostgreSQL 连接池
    poolConfig, _ := pgxpool.ParseConfig(databaseURL)
    poolConfig.MaxConns = 50
    poolConfig.MinConns = 10
    
    dbPool, _ = pgxpool.ConnectConfig(context.Background(), poolConfig)
    
    return nil
}
```

### 2. 读写分离

```go
type SplitBackend struct {
    writeBackend *QdrantClient  // 主节点
    readBackends []*QdrantClient // 副本节点（多个）
    lb           *LoadBalancer
}

func (s *SplitBackend) Search(ctx context.Context, query map[string]interface{}) (interface{}, error) {
    // 读操作：负载均衡到副本
    backend := s.lb.Next()
    return backend.Search(ctx, query)
}

func (s *SplitBackend) Insert(ctx context.Context, docs []Document) error {
    // 写操作：只写主节点
    return s.writeBackend.Insert(ctx, docs)
}
```

### 3. 缓存层

```go
type CachedRetriever struct {
    backend VectorStore
    cache   *Cache
}

func (r *CachedRetriever) Search(query string) ([]Document, error) {
    cacheKey := hashQuery(query)
    
    // 1. 检查缓存
    if cached, found := r.cache.Get(cacheKey); found {
        return cached.([]Document), nil
    }
    
    // 2. 查询后端
    docs, err := r.backend.Search(query)
    if err != nil {
        return nil, err
    }
    
    // 3. 缓存结果（1小时）
    r.cache.Set(cacheKey, docs, 1*time.Hour)
    
    return docs, nil
}
```

## 📈 监控与调优

### 1. 性能指标收集

```go
type Metrics struct {
    QueryLatency   prometheus.Histogram
    EmbeddingLatency prometheus.Histogram
    CacheHitRate   prometheus.Counter
    ErrorRate      prometheus.Counter
}

func RecordQuery(duration time.Duration, cacheHit bool) {
    metrics.QueryLatency.Observe(duration.Seconds())
    
    if cacheHit {
        metrics.CacheHitRate.Inc()
    }
}
```

### 2. 自动调优

```go
// 根据负载动态调整参数
type AdaptiveConfig struct {
    currentLoad   int32
    maxTopK       int
    cacheSize     int64
}

func (c *AdaptiveConfig) GetTopK() int {
    load := atomic.LoadInt32(&c.currentLoad)
    
    // 高负载时减少 top-K
    if load > 80 {
        return 10
    } else if load > 50 {
        return 20
    }
    
    return c.maxTopK
}
```

## 🎯 性能基准

### 硬件配置

- CPU: 8 核
- 内存: 16GB
- 存储: NVMe SSD

### 向量检索基准

| 数据规模 | Top-K | P50 | P95 | P99 | QPS |
|---------|-------|-----|-----|-----|-----|
| 100K    | 10    | 8ms | 15ms| 25ms| 2000|
| 1M      | 10    | 20ms| 40ms| 60ms| 1000|
| 10M     | 10    | 50ms| 90ms| 150ms| 400|

### RAG 端到端延迟

| 组件 | 延迟 (P95) | 占比 |
|-----|-----------|------|
| Embedding | 100ms | 40% |
| 向量检索 | 50ms | 20% |
| 重排序 | 30ms | 12% |
| LLM 生成 | 70ms | 28% |
| **总计** | **250ms** | **100%** |

## 🎓 最佳实践总结

1. **索引优化**
   - 使用 HNSW 索引
   - 启用量化减少内存
   - 定期重建索引

2. **查询优化**
   - 合理设置 Top-K (10-20)
   - 使用标量过滤减少计算
   - 并行执行多个查询

3. **缓存策略**
   - 缓存 Embedding (1小时)
   - 缓存热门查询 (15分钟)
   - 使用 LRU 淘汰策略

4. **架构设计**
   - 连接池复用连接
   - 读写分离提高吞吐
   - 多级缓存减少延迟

5. **监控调优**
   - 监控关键指标
   - 定期性能测试
   - 根据负载动态调整

---

**相关文档**: [RAG_BEST_PRACTICES.md](./RAG_BEST_PRACTICES.md)

