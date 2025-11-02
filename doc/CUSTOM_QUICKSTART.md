# Custom 接口快速开始（5 分钟上手）

## 🎯 快速理解

**一句话**：`Custom` 接口让你用 5-30 分钟就能支持任何数据库！

---

## 🚀 最简示例（5 分钟）

### Step 1: 定义 Custom 结构体

```go
package main

import "github.com/fndome/xb"

type MyDBCustom struct {
    // 配置参数（可选）
    Timeout int
}
```

---

### Step 2: 实现 Generate 方法

```go
func (c *MyDBCustom) Generate(built *xb.Built) (interface{}, error) {
    // 生成你的数据库需要的查询格式
    return `{"query": "hello world"}`, nil
}
```

---

### Step 3: 使用

```go
func main() {
    custom := &MyDBCustom{Timeout: 30}
    
    built := xb.Of("users").
        Custom(custom).  // ⭐ 设置 Custom
        Eq("age", 18).
        Build()
    
    json, _ := built.JsonOfSelect()
    println(json)  // {"query": "hello world"}
}
```

**完成！** ✅

---

## 💎 实战示例 1：Milvus 向量数据库（30 分钟）

### 完整实现

```go
package xb

import (
    "encoding/json"
    "fmt"
)

// ============================================================================
// 1. 定义 Milvus Custom
// ============================================================================

type MilvusCustom struct {
    DefaultNProbe     int
    DefaultRoundDec   int
    DefaultMetricType string
}

func NewMilvusCustom() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:     64,
        DefaultRoundDec:   4,
        DefaultMetricType: "L2",
    }
}

// ============================================================================
// 2. 实现 Generate 方法
// ============================================================================

func (c *MilvusCustom) Generate(built *Built) (interface{}, error) {
    return built.toMilvusJSON()
}

// ============================================================================
// 3. 实现 JSON 生成逻辑
// ============================================================================

type MilvusSearchRequest struct {
    CollectionName string             `json:"collection_name"`
    Data           [][]float32        `json:"data"`
    Limit          int                `json:"limit"`
    SearchParams   MilvusSearchParams `json:"search_params"`
    Expr           string             `json:"expr,omitempty"`
}

type MilvusSearchParams struct {
    MetricType string                 `json:"metric_type"`
    Params     map[string]interface{} `json:"params"`
}

func (built *Built) toMilvusJSON() (string, error) {
    // 提取向量检索参数
    var vectorBb *Bb
    for i := range built.Conds {
        if built.Conds[i].Op == VECTOR_SEARCH {
            vectorBb = &built.Conds[i]
            break
        }
    }
    
    if vectorBb == nil {
        return "", fmt.Errorf("no vector search found")
    }
    
    params := vectorBb.Value.(VectorSearchParams)
    
    // 构建 Milvus 请求
    req := &MilvusSearchRequest{
        CollectionName: params.TableName,
        Data:           [][]float32{params.Vector},
        Limit:          params.Limit,
        SearchParams: MilvusSearchParams{
            MetricType: "L2",
            Params:     map[string]interface{}{"nprobe": 64},
        },
    }
    
    // 应用过滤器
    if len(built.Conds) > 1 {
        // 构建 expr（如：age > 18 AND city == "Beijing"）
        req.Expr = buildMilvusExpr(built.Conds)
    }
    
    // 序列化为 JSON
    bytes, err := json.MarshalIndent(req, "", "  ")
    if err != nil {
        return "", err
    }
    
    return string(bytes), nil
}

func buildMilvusExpr(conds []Bb) string {
    // 简化实现
    return "age > 18"
}

// ============================================================================
// 4. 使用
// ============================================================================

func main() {
    queryVector := xb.Vector{0.1, 0.2, 0.3}
    
    built := xb.Of("code_vectors").
        Custom(NewMilvusCustom()).  // ⭐ Milvus
        VectorSearch("embedding", queryVector, 20).
        Eq("language", "golang").
        Build()
    
    json, _ := built.JsonOfSelect()
    println(json)
    // {
    //   "collection_name": "code_vectors",
    //   "data": [[0.1, 0.2, 0.3]],
    //   "limit": 20,
    //   "search_params": {
    //     "metric_type": "L2",
    //     "params": {"nprobe": 64}
    //   },
    //   "expr": "language == 'golang'"
    // }
}
```

---

## 💎 实战示例 2：Oracle 分页（30 分钟）

### 完整实现

```go
package xb

import "fmt"

// ============================================================================
// 1. 定义 Oracle Custom
// ============================================================================

type OracleCustom struct {
    UseRowNum bool  // true: ROWNUM, false: FETCH FIRST (12c+)
}

func NewOracleCustom() *OracleCustom {
    return &OracleCustom{UseRowNum: true}
}

// ============================================================================
// 2. 实现 Generate 方法
// ============================================================================

func (c *OracleCustom) Generate(built *Built) (interface{}, error) {
    // 检查是否是分页查询
    if built.PageCondition != nil {
        return c.generatePageSQL(built)
    }
    
    // 普通查询使用默认实现
    sql, args, meta := built.defaultSQL()
    return &SQLResult{SQL: sql, Args: args, Meta: meta}, nil
}

// ============================================================================
// 3. 实现分页 SQL 生成
// ============================================================================

func (c *OracleCustom) generatePageSQL(built *Built) (*SQLResult, error) {
    page := built.PageCondition.Page
    rows := built.PageCondition.Rows
    
    offset := (page - 1) * rows
    limit := rows
    
    // 提取基础查询
    baseSQL, args := c.buildBaseSQL(built)
    
    var dataSQL string
    var countSQL string
    
    if c.UseRowNum {
        // ROWNUM 方式（Oracle 11g 及以下）
        dataSQL = fmt.Sprintf(`
SELECT * FROM (
  SELECT a.*, ROWNUM rn FROM (
    %s
  ) a WHERE ROWNUM <= %d
) WHERE rn > %d`, baseSQL, offset+limit, offset)
        
        countSQL = fmt.Sprintf("SELECT COUNT(*) FROM (%s)", baseSQL)
    } else {
        // FETCH FIRST 方式（Oracle 12c+）
        dataSQL = fmt.Sprintf(`%s
OFFSET %d ROWS
FETCH NEXT %d ROWS ONLY`, baseSQL, offset, limit)
        
        countSQL = fmt.Sprintf("SELECT COUNT(*) FROM (%s)", baseSQL)
    }
    
    return &SQLResult{
        SQL:      dataSQL,
        CountSQL: countSQL,  // ⭐ 提供独立的 Count SQL
        Args:     args,
    }, nil
}

func (c *OracleCustom) buildBaseSQL(built *Built) (string, []interface{}) {
    // 简化：生成 SELECT * FROM users WHERE age > ?
    return "SELECT * FROM users WHERE age > ?", []interface{}{18}
}

// ============================================================================
// 4. 使用
// ============================================================================

func main() {
    built := xb.Of("users").
        Custom(NewOracleCustom()).  // ⭐ Oracle
        Eq("age", 18).
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(3).Rows(20)  // 第 3 页，每页 20 条
        }).
        Build()
    
    countSQL, dataSQL, args, _ := built.SqlOfPage()
    
    fmt.Println("Count SQL:", countSQL)
    // SELECT COUNT(*) FROM (SELECT * FROM users WHERE age > ?)
    
    fmt.Println("Data SQL:", dataSQL)
    // SELECT * FROM (
    //   SELECT a.*, ROWNUM rn FROM (
    //     SELECT * FROM users WHERE age > ?
    //   ) a WHERE ROWNUM <= 60
    // ) WHERE rn > 40
    
    fmt.Println("Args:", args)
    // [18]
}
```

---

## 💎 实战示例 3：ClickHouse 批量插入（30 分钟）

```go
package xb

import "fmt"

// ============================================================================
// 1. 定义 ClickHouse Custom
// ============================================================================

type ClickHouseCustom struct {
    UseJSONFormat bool
}

func NewClickHouseCustom() *ClickHouseCustom {
    return &ClickHouseCustom{UseJSONFormat: true}
}

// ============================================================================
// 2. 实现 Generate 方法
// ============================================================================

func (c *ClickHouseCustom) Generate(built *Built) (interface{}, error) {
    // Insert
    if built.Inserts != nil {
        return c.generateInsertSQL(built)
    }
    
    // Update（ClickHouse 特殊语法）
    if built.Updates != nil {
        return c.generateUpdateSQL(built)
    }
    
    // Delete（ClickHouse 特殊语法）
    if built.Delete {
        return c.generateDeleteSQL(built)
    }
    
    // Select（默认）
    sql, args, meta := built.defaultSQL()
    return &SQLResult{SQL: sql, Args: args, Meta: meta}, nil
}

// ============================================================================
// 3. 实现 ClickHouse Insert（批量 JSON）
// ============================================================================

func (c *ClickHouseCustom) generateInsertSQL(built *Built) (*SQLResult, error) {
    if c.UseJSONFormat {
        // FORMAT JSONEachRow（高性能批量插入）
        sql := fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow\n", built.Table)
        
        // 假设 built.Inserts 是 []map[string]interface{}
        for _, row := range built.Inserts {
            json := toJSON(row)
            sql += json + "\n"
        }
        
        return &SQLResult{
            SQL:  sql,
            Args: nil,  // ⭐ ClickHouse JSONEachRow 不需要占位符
        }, nil
    }
    
    // 标准 INSERT
    sql, args := built.defaultInsertSQL()
    return &SQLResult{SQL: sql, Args: args}, nil
}

// ============================================================================
// 4. 实现 ClickHouse Update（ALTER TABLE）
// ============================================================================

func (c *ClickHouseCustom) generateUpdateSQL(built *Built) (*SQLResult, error) {
    // ClickHouse 的 UPDATE 是 ALTER TABLE UPDATE
    sql := fmt.Sprintf("ALTER TABLE %s UPDATE name = ?, age = ? WHERE id = ?",
        built.Table)
    
    args := []interface{}{"张三", 18, 123}
    
    return &SQLResult{SQL: sql, Args: args}, nil
}

// ============================================================================
// 5. 使用
// ============================================================================

func main() {
    // ClickHouse 批量插入
    built := xb.Of("users").
        Custom(NewClickHouseCustom()).
        // 假设有批量插入数据
        Build()
    
    sql, args := built.SqlOfInsert()
    
    fmt.Println(sql)
    // INSERT INTO users FORMAT JSONEachRow
    // {"id": 1, "name": "张三", "age": 18}
    // {"id": 2, "name": "李四", "age": 20}
    
    fmt.Println(args)
    // []  (ClickHouse JSONEachRow 不需要参数)
}

func toJSON(data interface{}) string {
    // 简化实现
    return `{"id": 1, "name": "test"}`
}
```

---

## 🎯 进阶：预设模式

### 为什么需要预设模式？

用户不想每次都配置参数，希望开箱即用！

```go
// ============================================================================
// 预设模式（Milvus）
// ============================================================================

func NewMilvusCustom() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:   64,   // 默认
        DefaultRoundDec: 4,
    }
}

func MilvusHighPrecision() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:   256,  // 高精度
        DefaultRoundDec: 6,
    }
}

func MilvusHighSpeed() *MilvusCustom {
    return &MilvusCustom{
        DefaultNProbe:   16,   // 高速
        DefaultRoundDec: 2,
    }
}

// ============================================================================
// 使用
// ============================================================================

// 默认模式
built := xb.Of("t").Custom(NewMilvusCustom()).Build()

// 高精度模式
built := xb.Of("t").Custom(MilvusHighPrecision()).Build()

// 高速模式
built := xb.Of("t").Custom(MilvusHighSpeed()).Build()
```

---

## 🎯 进阶：运行时切换

### 多数据库部署

```go
package main

import (
    "github.com/fndome/xb"
    "os"
)

func main() {
    // 根据环境变量选择数据库
    var custom xb.Custom
    
    switch os.Getenv("VECTOR_DB") {
    case "qdrant":
        custom = xb.NewQdrantCustom()
    case "milvus":
        custom = NewMilvusCustom()
    case "weaviate":
        custom = NewWeaviateCustom()
    default:
        custom = xb.NewQdrantCustom()
    }
    
    // ⭐ 统一的查询构建
    built := xb.Of("documents").
        Custom(custom).  // ⭐ 运行时切换
        VectorSearch("embedding", queryVec, 20).
        Eq("status", "published").
        Build()
    
    // ⭐ 统一的接口
    json, _ := built.JsonOfSelect()
    
    // 根据配置调用不同的客户端
    switch os.Getenv("VECTOR_DB") {
    case "qdrant":
        results, _ := qdrantClient.Search(json)
    case "milvus":
        results, _ := milvusClient.Search(json)
    case "weaviate":
        results, _ := weaviateClient.Search(json)
    }
}
```

---

## 📝 检查清单

实现一个 Custom，需要：

- [ ] **定义结构体**：包含默认配置参数
- [ ] **实现 Generate()**：一个方法
- [ ] **返回正确类型**：string（JSON）或 *SQLResult（SQL）
- [ ] **提供预设模式**：`NewXxxCustom()`、`XxxHighPrecision()` 等
- [ ] **编写测试**：验证生成的 JSON/SQL 正确
- [ ] **文档注释**：说明使用方法

---

## 🎯 常见问题

### Q1: Generate() 应该返回什么？

**A**: 
- 向量数据库：返回 `string`（JSON）
- SQL 数据库：返回 `*xb.SQLResult`

### Q2: 如何提取 Built 中的参数？

**A**:
```go
// 提取向量检索参数
for _, bb := range built.Conds {
    if bb.Op == xb.VECTOR_SEARCH {
        params := bb.Value.(xb.VectorSearchParams)
        // 使用 params.Vector, params.Limit 等
    }
}

// 提取标量过滤器
for _, bb := range built.Conds {
    if bb.Op == xb.EQ {
        // bb.Key, bb.Value
    }
}
```

### Q3: 如何处理分页？

**A**:
```go
if built.PageCondition != nil {
    page := built.PageCondition.Page
    rows := built.PageCondition.Rows
    offset := (page - 1) * rows
    
    // 生成分页 SQL
    return &xb.SQLResult{
        SQL:      dataSQL,
        CountSQL: countSQL,  // ⭐ 提供 CountSQL
        Args:     args,
    }, nil
}
```

### Q4: 如何支持 Insert/Update/Delete？

**A**:
```go
func (c *Custom) Generate(built *xb.Built) (interface{}, error) {
    if built.Inserts != nil {
        return c.generateInsert(built)
    }
    
    if built.Updates != nil {
        return c.generateUpdate(built)
    }
    
    if built.Delete {
        return c.generateDelete(built)
    }
    
    // Select
    return c.generateSelect(built)
}
```

---

## 🚀 下一步

1. ✅ 阅读 `xb/qdrant_custom.go`（官方示例）
2. ✅ 参考 `xb/doc/CUSTOM_VECTOR_DB_GUIDE.md`（完整指南）
3. ✅ 实现你的 Custom（5-30 分钟）
4. ✅ 编写测试
5. ✅ 享受 Custom 接口的强大！

---

**开始实现你的第一个 Custom 吧！** 🎉✨

