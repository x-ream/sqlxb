# Qdrant 高级 API

**版本**: v0.10.0  
**状态**: ✅ 已实现

---

## 📋 概述

sqlxb v0.10.0 新增 Qdrant 高级功能：
- **Recommend API**: 基于正负样本的推荐查询
- **Discover API**: 基于上下文的探索性查询
- **Scroll API**: 大数据集游标遍历

---

## 🎯 Recommend API

### 用途

基于用户喜欢（正样本）和不喜欢（负样本）的内容进行推荐。

### 使用示例

#### 基本推荐

```go
built := sqlxb.Of(&Article{}).
    Eq("category", "tech").
    QdrantX(func(qx *QdrantBuilderX) {
        qx.Recommend(func(rb *RecommendBuilder) {
            rb.Positive(123, 456, 789)  // 用户喜欢的文章 ID
            rb.Negative(111, 222)        // 用户不喜欢的文章 ID（可选）
            rb.Limit(20)                 // 返回数量
        })
    }).
    Build()

json, err := built.ToQdrantRecommendJSON()
```

**生成的 JSON**:
```json
{
  "positive": [123, 456, 789],
  "negative": [111, 222],
  "limit": 20,
  "filter": {
    "must": [
      {"key": "category", "match": {"value": "tech"}}
    ]
  }
}
```

---

#### 只用正样本

```go
qx.Recommend(func(rb *RecommendBuilder) {
    rb.Positive(100, 200, 300).Limit(15)
})
```

---

#### 结合 Qdrant 参数

```go
qx.Recommend(func(rb *RecommendBuilder) {
    rb.Positive(123, 456)
    rb.Negative(789)
    rb.Limit(20)
}).
HnswEf(256).              // 高精度搜索
ScoreThreshold(0.8).       // 相似度阈值
WithVector(true)           // 返回向量
```

**生成的 JSON**:
```json
{
  "positive": [123, 456],
  "negative": [789],
  "limit": 20,
  "with_vector": true,
  "score_threshold": 0.8,
  "params": {
    "hnsw_ef": 256
  }
}
```

---

## 🔍 Discover API

### 用途

基于用户的浏览/交互历史，发现"中间地带"的新内容。

### 使用示例

#### 基本探索

```go
built := sqlxb.Of(&Article{}).
    Eq("category", "tech").
    QdrantX(func(qx *QdrantBuilderX) {
        qx.Discover(func(db *DiscoverBuilder) {
            db.Context(101, 102, 103)  // 用户浏览历史
            db.Limit(20)
        })
    }).
    Build()

json, err := built.ToQdrantDiscoverJSON()
```

**生成的 JSON**:
```json
{
  "context": [101, 102, 103],
  "limit": 20,
  "filter": {
    "must": [
      {"key": "category", "match": {"value": "tech"}}
    ]
  }
}
```

---

#### 结合 Qdrant 参数

```go
qx.Discover(func(db *DiscoverBuilder) {
    db.Context(100, 200, 300, 400)
    db.Limit(15)
}).
HnswEf(256).
ScoreThreshold(0.75).
WithVector(true)
```

---

## 🔄 Scroll API

### 用途

遍历大量结果（10K+），避免 OFFSET 性能问题。

### 使用示例

#### 初始查询

```go
// 第一次查询（不设置 scroll_id）
built := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVec, 100).
    Build()

json, err := built.ToQdrantJSON()
// 发送到 Qdrant，获得 scroll_id
```

---

#### 继续滚动

```go
// 使用返回的 scroll_id 继续获取
built := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    QdrantX(func(qx *QdrantBuilderX) {
        qx.ScrollID("scroll-12345-abcde-xyz")
    }).
    Build()

json, err := built.ToQdrantScrollJSON()
```

**生成的 JSON**:
```json
{
  "scroll_id": "scroll-12345-abcde-xyz",
  "limit": 100,
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}}
    ]
  }
}
```

---

## 🎯 实际应用场景

### 场景 1: 文章推荐系统

```go
// 用户阅读历史：喜欢 Golang 和分布式系统的文章，不喜欢 PHP
built := sqlxb.Of(&Article{}).
    Eq("status", "published").
    QdrantX(func(qx *QdrantBuilderX) {
        qx.Recommend(func(rb *RecommendBuilder) {
            rb.Positive(101, 102, 103)  // Golang + 分布式文章
            rb.Negative(201)             // PHP 文章（不感兴趣）
            rb.Limit(10)                 // 返回 10 条
        }).
        ScoreThreshold(0.75)             // 最低相似度 75%
    }).
    Build()
```

---

### 场景 2: 代码库推荐

```go
// 基于用户 Star 的项目推荐类似项目
built := sqlxb.Of(&Repository{}).
    Eq("language", "go").
    QdrantX(func(qx *QdrantBuilderX) {
        qx.Recommend(func(rb *RecommendBuilder) {
            rb.Positive(userStarredRepos...)  // 用户 Star 的项目
            rb.Negative(userIgnoredRepos...)  // 用户忽略的项目
            rb.Limit(20)                      // 返回 20 个
        })
    }).
    Build()
```

---

### 场景 3: 探索性搜索（Discover）

```go
// 用户阅读了几篇文章后，系统发现"共同主题"
built := sqlxb.Of(&Article{}).
    Eq("status", "published").
    QdrantX(func(qx *QdrantBuilderX) {
        qx.Discover(func(db *DiscoverBuilder) {
            db.Context(101, 102, 103, 104)  // 用户阅读历史
            db.Limit(20)
        }).
        ScoreThreshold(0.7)
    }).
    Build()

// 可能发现：这些文章的共同主题是"云原生"或"微服务"
```

---

### 场景 4: 大规模数据导出（Scroll）

```go
// 导出 100 万条代码向量
scrollID := ""
allResults := []CodeVector{}

for {
    var built *Built
    if scrollID == "" {
        // 初始查询
        built = sqlxb.Of(&CodeVector{}).
            Eq("language", "golang").
            VectorSearch("embedding", queryVec, 1000).
            Build()
        json, _ := built.ToQdrantJSON()
        // 调用 Qdrant，获得 scroll_id 和首批结果
    } else {
        // 继续滚动
        built = sqlxb.Of(&CodeVector{}).
            Eq("language", "golang").
            QdrantX(func(qx *QdrantBuilderX) {
                qx.ScrollID(scrollID)
            }).
            Build()
        json, _ := built.ToQdrantScrollJSON()
        // 调用 Qdrant，获得下一批结果
    }
    
    if len(results) == 0 {
        break
    }
    
    allResults = append(allResults, results...)
}
```

---

## 🎊 API 对比

| 功能 | 传统 Search | Recommend | Discover | Scroll |
|------|------------|-----------|----------|--------|
| **输入** | Query Vector | Positive/Negative IDs | Context IDs | Scroll ID |
| **用途** | 相似度搜索 | 基于偏好推荐 | 探索共性主题 | 大数据集遍历 |
| **性能** | ✅ 快 | ✅ 快 | ✅ 快 | ✅ 恒定 |
| **适合场景** | 查询相似内容 | 个性化推荐 | 内容探索 | 数据导出/批处理 |

---

## 📚 参考文档

- [Qdrant Recommendation API](https://qdrant.tech/documentation/concepts/explore/#recommendation-api)
- [Qdrant Discovery API](https://qdrant.tech/documentation/concepts/explore/#discovery-api)
- [Qdrant Scroll API](https://qdrant.tech/documentation/concepts/points/#scroll-points)
- [QdrantX 使用指南](./QDRANT_X_USAGE.md)

---

**版本**: v0.10.0  
**更新日期**: 2025-10-27  
**新增**: Recommend, Discover, Scroll API

