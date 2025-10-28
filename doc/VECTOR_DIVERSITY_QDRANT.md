# xb 向量多样性查询 + Qdrant 支持

## 📋 概述

`sqlxb v0.9.x` 添加了向量结果多样性支持和 Qdrant JSON 生成功能。

**核心特性**：
- ✅ 三种多样性策略：哈希去重、最小距离、MMR 算法
- ✅ 自动生成 Qdrant 搜索 JSON
- ✅ **优雅降级**：PostgreSQL 自动忽略多样性参数
- ✅ 一份代码，多种后端（PostgreSQL + Qdrant）

---

## 🎯 解决的问题

### 问题：查询结果缺乏多样性

```
场景：代码向量数据库
  总数据：1000 个代码片段
  
  查询："用户登录相关代码"
  ↓
  返回 Top-20 结果：
  ├── 结果1: login() { username, password }  - 0.98
  ├── 结果2: login() { user, pass }          - 0.97
  ├── 结果3: userLogin() { name, pwd }       - 0.96
  ├── ...
  └── 结果20: 几乎都是重复的登录逻辑        - 0.92

  ❌ 20 个结果太相似，缺乏多样性！
  ❌ 用户看不到不同的实现方式
```

### 解决方案：多样性过滤

```
相同查询 + 多样性：
  ↓
  返回 Top-20 结果：
  ├── 结果1: JWT token 登录               - 0.98
  ├── 结果2: OAuth 社交登录               - 0.95
  ├── 结果3: 生物识别登录                 - 0.93
  ├── 结果4: 短信验证码登录               - 0.91
  ├── ...
  └── 结果20: 20 种不同的登录实现         - 0.85

  ✅ 多样化的结果
  ✅ 用户获得更多灵感
```

---

## 🚀 快速开始

### 1. 安装

```bash
go get github.com/x-ream/xb@v0.9.2
```

### 2. 数据模型

```go
type CodeVector struct {
    Id           int64   `db:"id"`
    Content      string  `db:"content"`
    Embedding    Vector  `db:"embedding"`
    Language     string  `db:"language"`
    SemanticHash string  `db:"semantic_hash"`  // ⭐ 用于哈希去重
}

func (CodeVector) TableName() string {
    return "code_vectors"
}
```

### 3. 基础用法

```go
import "github.com/x-ream/xb"

queryVector := Vector{0.1, 0.2, 0.3, 0.4}

// 不带多样性（传统查询）
builder := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 20)
```

---

## 📚 多样性策略

### 策略 1: 哈希去重（推荐）⭐

**适用场景**：内容相似但不完全相同的结果

```go
// API
builder := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 20).
    WithHashDiversity("semantic_hash")  // ⭐ 基于 semantic_hash 去重

// PostgreSQL: 自动忽略多样性
sql, args := builder.Build().SqlOfVectorSearch()
// SQL: SELECT ... LIMIT 20

// Qdrant: 应用多样性
json, _ := builder.Build().ToQdrantJSON()
// limit: 100 (20 * 5 倍过度获取)
// 应用层基于 semantic_hash 去重到 20 个
```

**语义哈希计算**：

```go
import (
    "crypto/sha256"
    "encoding/hex"
    "strings"
)

func ComputeSemanticHash(content string) string {
    // 归一化代码：去除空白、注释、变量名
    normalized := normalizeCode(content)
    
    // SHA256 哈希
    hash := sha256.Sum256([]byte(normalized))
    return hex.EncodeToString(hash[:8])  // 取前 8 字节
}

func normalizeCode(code string) string {
    // 1. 转小写
    code = strings.ToLower(code)
    
    // 2. 去除空白
    code = strings.Join(strings.Fields(code), " ")
    
    // 3. 去除注释（简化示例）
    // TODO: 更复杂的归一化逻辑
    
    return code
}
```

---

### 策略 2: 最小距离

**适用场景**：确保结果在向量空间中足够分散

```go
// API
builder := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", queryVector, 20).
    WithMinDistance(0.3)  // ⭐ 结果之间最小距离 0.3

// Qdrant JSON
{
  "vector": [0.1, 0.2, 0.3],
  "limit": 100,  // 过度获取
  ...
}

// 应用层过滤伪代码：
func applyMinDistance(results []Result, minDist float32) []Result {
    diverse := []Result{}
    
    for _, result := range results {
        isDiverse := true
        
        for _, selected := range diverse {
            if distance(result, selected) < minDist {
                isDiverse = false
                break
            }
        }
        
        if isDiverse {
            diverse = append(diverse, result)
        }
        
        if len(diverse) >= 20 {
            break
        }
    }
    
    return diverse
}
```

---

### 策略 3: MMR 算法

**适用场景**：平衡相关性和多样性

```go
// API
builder := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", queryVector, 20).
    WithMMR(0.5)  // ⭐ lambda=0.5 平衡

// lambda 参数说明：
// 0.0 = 完全多样性（结果差异最大化）
// 1.0 = 完全相关性（只考虑与查询的相似度）
// 0.5 = 平衡（推荐）
```

**MMR 算法公式**：

```
Score(Di) = λ × Similarity(Di, Query) 
          - (1-λ) × max[Similarity(Di, Dj)]
                    j ∈ Selected

其中：
- Di: 候选结果
- Query: 查询向量
- Selected: 已选择的结果
- λ: 平衡参数
```

---

## 🔄 Qdrant JSON 生成

### 基础 JSON

```go
queryVector := Vector{0.1, 0.2, 0.3, 0.4}

built := sqlxb.Of(&CodeVector{}).
    VectorSearch("embedding", queryVector, 10).
    Build()

json, err := built.ToQdrantJSON()
```

**输出**：

```json
{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "limit": 10,
  "with_payload": true,
  "params": {
    "hnsw_ef": 128
  }
}
```

---

### 带过滤器

```go
built := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    Gt("quality_score", 0.8).
    VectorSearch("embedding", queryVector, 20).
    Build()

json, _ := built.ToQdrantJSON()
```

**输出**：

```json
{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "limit": 20,
  "filter": {
    "must": [
      {
        "key": "language",
        "match": {"value": "golang"}
      },
      {
        "key": "quality_score",
        "range": {"gt": 0.8}
      }
    ]
  },
  "with_payload": true,
  "params": {
    "hnsw_ef": 128
  }
}
```

---

### 带多样性

```go
built := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 20).
    WithHashDiversity("semantic_hash").  // ⭐ 多样性
    Build()

json, _ := built.ToQdrantJSON()
```

**输出**：

```json
{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "limit": 100,  // ⭐ 自动扩大到 20 * 5 = 100
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}}
    ]
  },
  "with_payload": true,
  "params": {
    "hnsw_ef": 128
  }
}
```

**注意**：Qdrant 不原生支持多样性，需要在应用层处理：
1. 获取 100 个结果（过度获取）
2. 基于 `semantic_hash` 去重
3. 返回 Top-20

---

## 🎨 实际应用示例

### 示例 1: 代码搜索

```go
// 用户查询："用户登录相关代码"
queryVector := embedding.Encode("用户登录相关代码")

// 构建查询（一份代码）
builder := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    Gt("quality_score", 0.7).
    VectorSearch("embedding", queryVector, 20).
    WithHashDiversity("semantic_hash")

built := builder.Build()

// 后端 1: PostgreSQL (开发/小规模)
sql, args := built.SqlOfVectorSearch()
rows, err := db.Query(sql, args...)
// SQL 自动忽略多样性，返回 Top-20

// 后端 2: Qdrant (生产/大规模)
json, err := built.ToQdrantJSON()
// POST http://qdrant:6333/collections/code_vectors/points/search
// 获取 100 个，应用层去重到 20 个
```

---

### 示例 2: 文档检索

```go
type Document struct {
    Id           int64
    Title        string
    Content      string
    Embedding    Vector
    Category     string
    ContentHash  string  // 用于去重
}

// 查询
queryVector := embedding.Encode("如何部署 Kubernetes")

results := sqlxb.Of(&Document{}).
    Eq("category", "devops").
    VectorSearch("embedding", queryVector, 10).
    WithHashDiversity("content_hash").
    Build().
    Query()  // 假设有 Query() 方法

// 返回 10 个多样化的 DevOps 文档
```

---

### 示例 3: 推荐系统

```go
// 用户浏览了一篇关于 "Golang 并发" 的文章
// 推荐相关但多样化的文章

articleVector := article.Embedding

recommendations := sqlxb.Of(&Article{}).
    Ne("id", article.Id).  // 排除当前文章
    VectorSearch("embedding", articleVector, 10).
    WithMMR(0.6).  // 60% 相关性，40% 多样性
    Build().
    Query()

// 返回：
// - Golang 并发进阶（高相关）
// - Rust 并发模型（中相关，不同语言）
// - 分布式系统设计（中相关，不同领域）
// ...
```

---

## 🔧 高级配置

### 自定义过度获取因子

```go
// 默认 5 倍过度获取
builder.WithDiversity(sqlxb.DiversityByHash, "semantic_hash")
// limit: 20 * 5 = 100

// 自定义 10 倍
builder.WithDiversity(
    sqlxb.DiversityByHash, 
    "semantic_hash", 
    10,  // ⭐ 过度获取因子
)
// limit: 20 * 10 = 200
```

---

### 组合多种策略

```go
// 先哈希去重，再距离过滤
results := qdrantClient.Search(json)

// 应用层处理：
// 1. 基于 semantic_hash 去重
uniqueResults := deduplicateByHash(results, "semantic_hash")

// 2. 基于最小距离过滤
diverseResults := ensureMinDistance(uniqueResults, 0.3)

// 3. 返回 Top-20
return diverseResults[:20]
```

---

## 💡 最佳实践

### 1. 选择合适的策略

```
内容去重（代码、文档） → DiversityByHash ⭐
向量空间分散（图像、音频） → DiversityByDistance
平衡相关性和多样性（推荐系统） → DiversityByMMR
```

---

### 2. 语义哈希的重要性

```sql
-- 数据库 Schema 必须包含语义哈希字段
CREATE TABLE code_vectors (
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    embedding VECTOR(768) NOT NULL,
    semantic_hash VARCHAR(64),  -- ⭐ 关键字段
    language VARCHAR(50),
    
    INDEX idx_semantic_hash (semantic_hash)
);
```

---

### 3. 过度获取因子调优

```
数据相似度高 → factor = 10  (需要更多候选)
数据相似度低 → factor = 3   (候选已足够多样)

默认值 5 适合大多数场景
```

---

### 4. PostgreSQL vs Qdrant 选择

```
开发环境/小规模（< 1M 向量）:
  → PostgreSQL + pgvector
  → 简单部署
  → 多样性在应用层处理

生产环境/大规模（> 10M 向量）:
  → Qdrant
  → 高性能
  → 量化技术节省内存
  → 多样性在应用层处理
```

---

## 🎯 完整示例

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "github.com/x-ream/xb"
)

type CodeVector struct {
    Id           int64  `db:"id"`
    Content      string `db:"content"`
    Embedding    sqlxb.Vector `db:"embedding"`
    Language     string `db:"language"`
    SemanticHash string `db:"semantic_hash"`
}

func (CodeVector) TableName() string {
    return "code_vectors"
}

func main() {
    // 查询向量
    queryVector := sqlxb.Vector{0.1, 0.2, 0.3, 0.4}
    
    // 构建查询（一份代码）
    builder := sqlxb.Of(&CodeVector{}).
        Eq("language", "golang").
        VectorSearch("embedding", queryVector, 20).
        WithHashDiversity("semantic_hash")
    
    built := builder.Build()
    
    // ===== 后端 1: PostgreSQL =====
    fmt.Println("=== PostgreSQL ===")
    sql, args := built.SqlOfVectorSearch()
    fmt.Printf("SQL: %s\n", sql)
    fmt.Printf("Args: %v\n", args)
    
    // 执行查询
    // rows, err := db.Query(sql, args...)
    
    // ===== 后端 2: Qdrant =====
    fmt.Println("\n=== Qdrant ===")
    jsonStr, err := built.ToQdrantJSON()
    if err != nil {
        panic(err)
    }
    fmt.Println(jsonStr)
    
    // HTTP 请求 Qdrant
    // POST http://qdrant:6333/collections/code_vectors/points/search
    // Body: jsonStr
    
    // 应用层去重
    // results := qdrantClient.Search(jsonStr)
    // uniqueResults := deduplicateByHash(results, "semantic_hash", 20)
}
```

**输出**：

```
=== PostgreSQL ===
SQL: SELECT *, embedding <-> ? AS distance FROM code_vectors WHERE language = ? ORDER BY distance LIMIT 20
Args: [[0.1 0.2 0.3 0.4] golang]

=== Qdrant ===
{
  "vector": [0.1, 0.2, 0.3, 0.4],
  "limit": 100,
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}}
    ]
  },
  "with_payload": true,
  "params": {
    "hnsw_ef": 128
  }
}
```

---

## 📖 API 参考

### 多样性方法

```go
// 通用方法
WithDiversity(strategy DiversityStrategy, params ...interface{}) *BuilderX

// 快捷方法
WithHashDiversity(hashField string) *BuilderX
WithMinDistance(minDistance float32) *BuilderX
WithMMR(lambda float32) *BuilderX
```

### Qdrant 方法

```go
// 生成 JSON 字符串
ToQdrantJSON() (string, error)

// 生成请求结构体
ToQdrantRequest() (*QdrantSearchRequest, error)
```

---

## 🎊 总结

**sqlxb 向量多样性查询**：

✅ 解决了查询结果缺乏多样性的问题  
✅ 三种策略满足不同场景  
✅ 优雅降级，PostgreSQL 自动忽略  
✅ 一份代码，多种后端  
✅ AI-First 设计，易于维护

**开始使用**：

```bash
go get github.com/x-ream/xb@v0.9.2
```

**文档**：
- [向量快速开始](./VECTOR_QUICKSTART.md)
- [为什么选择 Qdrant](./WHY_QDRANT.md)
- [QdrantX 使用指南](./QDRANT_X_USAGE.md)

---

**问题反馈**：https://github.com/x-ream/xb/issues

