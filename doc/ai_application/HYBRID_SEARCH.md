# 混合检索策略指南

## 📋 概述

混合检索（Hybrid Search）结合向量相似度检索和传统标量过滤，是构建高质量 RAG 应用的关键技术。

## 🎯 什么是混合检索

```
传统检索: WHERE status='active' AND category='tech'
向量检索: ORDER BY embedding <=> query_vector
混合检索: WHERE status='active' AND category='tech' 
         ORDER BY embedding <=> query_vector
```

## 🏗️ sqlxb 混合检索实现

### 基础混合查询

```go
func HybridSearch(queryVector []float32, status string, category string) (string, error) {
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", queryVector, 20).  // 向量检索，返回 20 条
        Eq("status", status).                         // 标量过滤
        Eq("category", category).                     // 标量过滤
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(0.7)
        }).
        Build()

    return built.ToQdrantJSON()
}
```

### 复杂过滤条件

```go
func AdvancedHybridSearch(params SearchParams) (map[string]interface{}, error) {
    // ⭐ sqlxb 自动过滤 nil/0/空字符串/time.Time零值，直接传参即可
    builder := sqlxb.Of(&Document{}).
        VectorSearch("embedding", params.QueryVector, params.TopK).
        Eq("status", params.Status).            // 自动过滤空字符串
        Gte("created_at", params.StartDate).    // 自动过滤零值
        Lte("created_at", params.EndDate).      // 自动过滤零值
        In("category", params.Categories...).   // 自动过滤空切片
        Ne("status", "deleted").
        Or(func(cb *sqlxb.CondBuilder) {
            for _, tag := range params.Tags {
                cb.Like("tags", tag).OR()  // ⭐ sqlxb 自动添加 %tag%
            }
        })  // 空切片时 Or() 会被自动过滤
    
    built := builder.
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(float32(params.MinScore))
        }).
        Build()

    return built.ToQdrantJSON()
}
```

## 🎯 检索策略

### 1. 先过滤后检索（推荐）

```go
// 适用于：过滤条件能显著减少候选集
func FilterThenSearch(vector []float32, mustFilters map[string]interface{}) (string, error) {
    built := sqlxb.Of(&Document{}).
        Eq("status", mustFilters["status"]).       // 先过滤
        Eq("language", mustFilters["language"]).   // 缩小范围
        VectorSearch("embedding", vector, 10).      // 再向量检索，返回 10 条
        Build()

    return built.ToQdrantJSON()
}
```

### 2. 先检索后过滤

```go
// 适用于：需要大量候选结果再精筛
func SearchThenFilter(vector []float32, optionalFilters map[string]interface{}) (string, error) {
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector, 100).  // 粗召回，100 条
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(0.6)  // 相似度阈值
        }).
        Build()

    // 注意：后置过滤需要在应用层处理，取前 10 条
    return built.ToQdrantJSON()
}
```

### 3. 分阶段混合检索

```go
func MultiStageHybridSearch(params SearchParams) ([]Document, error) {
    // 阶段 1: 宽松向量检索 + 核心过滤
    built1 := sqlxb.Of(&Document{}).
        VectorSearch("embedding", params.Vector, 100).  // 粗召回 100 条
        Eq("language", params.Language).                 // 核心过滤
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(0.5)  // 宽松阈值
        }).
        Build()

    stage1JSON, err := built1.ToQdrantJSON()
    if err != nil {
        return nil, err
    }
    
    // 执行查询获取阶段1结果（伪代码）
    stage1Results := executeQdrantQuery(stage1JSON)
    
    // 阶段 2: 应用额外过滤
    filtered := applyBusinessFilters(stage1Results, params.Filters)
    
    // 阶段 3: 重排序
    reranked := rerankResults(filtered, params.RerankModel)
    
    // 阶段 4: 多样性控制
    diverse := applyMMR(reranked, params.Lambda, params.TopK)
    
    return diverse, nil
}
```

## 🔍 常见场景

### 场景 1: 时间敏感检索

```go
// 优先返回最新内容
func TimeAwareSearch(query string, vector []float32) (string, error) {
    sevenDaysAgo := time.Now().AddDate(0, 0, -7)
    
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector, 20).  // 返回 20 条
        Gte("published_at", sevenDaysAgo).       // 最近 7 天
        Eq("status", "published").                // 已发布
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(0.65)
        }).
        Build()

    return built.ToQdrantJSON()
}
```

### 场景 2: 多语言检索

```go
func MultilingualSearch(vector []float32, preferredLang string) (string, error) {
    // 优先返回首选语言，但也包含其他语言
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector, 20).
        Or(func(cb *sqlxb.CondBuilder) {
            cb.Eq("language", preferredLang).OR().  // 首选语言
               Eq("language", "en")                  // 备选语言
        }).
        Build()
    
    return built.ToQdrantJSON()
}
```

### 场景 3: 权限过滤检索

```go
func PermissionAwareSearch(vector []float32, userID int64, userRoles []string) (map[string]interface{}, error) {
    return sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector).
        Or(func(cb *sqlxb.CondBuilder) {
            // 公开文档
            cb.Eq("visibility", "public").OR()
            
            // 用户自己的文档
            cb.Eq("owner_id", userID).OR()
            
            // 角色可访问的文档
            cb.In("required_role", userRoles)
        }).
        Ne("status", "deleted").
        Build()

    return built.ToQdrantJSON()
}
```

### 场景 4: 层级分类检索

```go
// 支持层级分类：科技 > 人工智能 > 机器学习
func HierarchicalSearch(vector []float32, category string) (map[string]interface{}, error) {
    return sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector).
        Or(func(cb *sqlxb.CondBuilder) {
            // 精确匹配
            cb.Eq("category", category).OR()
            
            // 父类别
            cb.Like("category", category+":%").OR()
            
            // 子类别
            cb.Like("category", "%:"+category)
        }).
        Build()

    return built.ToQdrantJSON()
}
```

## 🎨 高级技巧

### 1. 动态权重

```go
// 根据文档新鲜度调整相似度分数
func FreshnessWeightedSearch(vector []float32) ([]Document, error) {
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector, 50).
        Build()
    
    qdrantJSON, _ := built.ToQdrantJSON()
    
    now := time.Now()
    for i := range results {
        // 计算文档年龄（天）
        age := now.Sub(results[i].CreatedAt).Hours() / 24
        
        // 应用时间衰减：score * e^(-age/30)
        decayFactor := math.Exp(-age / 30.0)
        results[i].Score *= decayFactor
    }
    
    // 重新排序
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    
    return results[:10], nil
}
```

### 2. 个性化检索

```go
func PersonalizedSearch(vector []float32, userID int64) (map[string]interface{}, error) {
    // 获取用户偏好
    userPrefs := getUserPreferences(userID)
    
    builder := sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector)
    
    // 应用个性化过滤
    if len(userPrefs.FavoriteCategories) > 0 {
        builder.In("category", userPrefs.FavoriteCategories)
    }
    
    if len(userPrefs.BlockedAuthors) > 0 {
        builder.NotIn("author_id", userPrefs.BlockedAuthors...)
    }
    
    built := builder.Build()
    return built.ToQdrantJSON()
}
```

### 3. 负反馈过滤

```go
func SearchWithNegativeFeedback(vector []float32, userID int64) (map[string]interface{}, error) {
    // 获取用户已看过/不感兴趣的文档
    viewedDocs := getUserViewHistory(userID, 30) // 最近 30 天
    dislikedDocs := getUserDislikes(userID)
    
    excludeIDs := append(viewedDocs, dislikedDocs...)
    
    built := sqlxb.Of(&Document{}).
        VectorSearch("embedding", vector, 20).
        NotIn("id", excludeIDs).  // 排除已看过的
        Build()

    return built.ToQdrantJSON()
}
```

## 📊 性能对比

| 策略 | 延迟 | 准确率 | 适用场景 |
|-----|------|-------|---------|
| 纯向量检索 | 10ms | 75% | 无结构化过滤需求 |
| 先过滤后检索 | 15ms | 85% | 过滤条件强 |
| 先检索后过滤 | 25ms | 90% | 需要大量候选 |
| 多阶段混合 | 50ms | 95% | 高质量要求 |

## 🎯 最佳实践

1. **优先使用强过滤条件**
   - 将能大幅减少候选集的条件放在前面
   - 例如：status, language, visibility

2. **合理设置 Top-K**
   - 粗召回阶段：Top-K = 50-100
   - 精排序阶段：Top-K = 10-20

3. **使用分数阈值**
   - 粗召回：threshold = 0.5-0.6
   - 精确检索：threshold = 0.7-0.8

4. **避免过度过滤**
   - 过多过滤条件可能导致召回不足
   - 平衡精确性和覆盖率

5. **监控查询性能**
   - 记录每个阶段的耗时
   - 识别性能瓶颈

---

**相关文档**:
- [RAG_BEST_PRACTICES.md](./RAG_BEST_PRACTICES.md)
- [PERFORMANCE.md](./PERFORMANCE.md)

