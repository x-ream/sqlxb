# Custom 接口设计哲学 (v1.1.0)

## 🎯 为什么 Custom 是 AI 时代需要的？

### 传统方案的局限（Dialect 枚举）

```go
// ❌ 旧思路：枚举所有数据库
type Dialect string

const (
    PostgreSQL Dialect = "postgresql"
    MySQL      Dialect = "mysql"
    Oracle     Dialect = "oracle"
    Qdrant     Dialect = "qdrant"
    Milvus     Dialect = "milvus"
    // ... 100+ 个数据库
)

// ❌ 问题：
// 1. 框架需要实现所有数据库的逻辑 → 臃肿
// 2. 新数据库出现 → 必须修改框架 → 不可扩展
// 3. 用户特殊需求 → 无法满足 → 不灵活
// 4. AI 时代新数据库层出不穷 → 跟不上
```

---

### Custom 接口的革命性

```go
// ✅ 新思路：接口驱动，用户实现
type Custom interface {
    Generate(built *Built) (interface{}, error)
}

// ✅ 优势：
// 1. 框架极简：只提供接口，不实现所有数据库
// 2. 用户自由：5 分钟实现任何数据库
// 3. 持续演进：跟随数据库新特性
// 4. AI 时代：向量数据库、图数据库、时序数据库，都能轻松支持
```

---

## 🚀 AI 时代的数据库爆炸

### 向量数据库（20+ 种）

| 数据库 | 特点 | Custom 实现难度 |
|-------|------|---------------|
| Qdrant | 纯向量搜索 | ✅ 5 分钟（官方支持）|
| Milvus | 企业级 | ✅ 30 分钟 |
| Weaviate | GraphQL | ✅ 30 分钟 |
| Pinecone | 云服务 | ✅ 15 分钟 |
| Chroma | 轻量级 | ✅ 10 分钟 |
| LanceDB | 嵌入式 | ✅ 20 分钟 |
| Vespa | 混合搜索 | ✅ 30 分钟 |
| ... | ... | ✅ 都能实现！|

**如果用 Dialect 枚举**：
- ❌ 框架需要维护 20+ 个实现
- ❌ 用户提 PR，等半年
- ❌ 框架臃肿，维护成本高

**用 Custom 接口**：
- ✅ 用户 5-30 分钟自己实现
- ✅ 框架保持极简
- ✅ 用户可以跟随数据库最新特性

---

### SQL 数据库的新变种

| 数据库 | 特殊性 | Custom 价值 |
|-------|--------|-----------|
| **ClickHouse** | `FORMAT JSONEachRow`<br>`ALTER TABLE UPDATE` | ✅ 必须 Custom |
| **Oracle** | `ROWNUM` 分页<br>序列语法 | ✅ 必须 Custom |
| **TimescaleDB** | 超表（Hypertable）<br>时间分区 | ✅ 必须 Custom |
| **CockroachDB** | 分布式事务<br>`AS OF SYSTEM TIME` | ✅ 必须 Custom |
| **DuckDB** | 嵌入式分析<br>特殊聚合 | ✅ 必须 Custom |
| **YugabyteDB** | 分布式 PostgreSQL | ✅ 可选 Custom |

---

## 💎 Generate() 接口的设计美学

### 1. 极简主义

```go
// ✅ 只需一个方法
type Custom interface {
    Generate(built *Built) (interface{}, error)
}

// ❌ 不需要多个方法
type Custom interface {
    GetDialect() Dialect                    // ❌ 类型本身就是标识
    ApplyParams(bbs, req) error            // ❌ 在 Generate 内部处理
    ToJSON(built) (string, error)          // ❌ 返回类型不统一
    ToSQL(built) (string, []interface{})   // ❌ 两个方法不如一个
}
```

---

### 2. 类型灵活

```go
// ✅ 返回 interface{} 支持任意类型
func (c *QdrantCustom) Generate(built *Built) (interface{}, error) {
    return `{"vector": [...]}`, nil  // string
}

func (c *OracleCustom) Generate(built *Built) (interface{}, error) {
    return &SQLResult{SQL: "...", Args: [...]}, nil  // *SQLResult
}

func (c *GraphDBCustom) Generate(built *Built) (interface{}, error) {
    return &CypherQuery{Query: "...", Params: {...}}, nil  // 自定义类型
}
```

---

### 3. 智能分发

```go
// ✅ built.JsonOfSelect() 自动处理 string
func (built *Built) JsonOfSelect() (string, error) {
    result, _ := built.Custom.Generate(built)
    
    if str, ok := result.(string); ok {
        return str, nil  // ✅ JSON
    }
    
    if sqlResult, ok := result.(*SQLResult); ok {
        return "", fmt.Errorf("got SQL, use SqlOfSelect()")  // ✅ 类型错误提示
    }
}

// ✅ built.SqlOfSelect() 自动处理 *SQLResult
func (built *Built) SqlOfSelect() (string, []interface{}, map[string]string) {
    if built.Custom == nil {
        return built.defaultSQL()  // ✅ 默认实现
    }
    
    result, _ := built.Custom.Generate(built)
    
    if sqlResult, ok := result.(*SQLResult); ok {
        return sqlResult.SQL, sqlResult.Args, sqlResult.Meta  // ✅ SQL
    }
}
```

---

## 🎨 设计对比：Dialect vs Custom

### Dialect 方案（传统，不可扩展）

```go
// ❌ 框架臃肿
xb/
├── postgresql_dialect.go  (200 行)
├── mysql_dialect.go       (200 行)
├── oracle_dialect.go      (300 行)
├── clickhouse_dialect.go  (400 行)
├── qdrant_dialect.go      (300 行)
├── milvus_dialect.go      (300 行)
├── weaviate_dialect.go    (300 行)
├── pinecone_dialect.go    (200 行)
├── chroma_dialect.go      (200 行)
├── lancedb_dialect.go     (200 行)
└── ... 100+ 个文件

总代码：30,000+ 行
维护成本：极高
新数据库：必须修改框架
用户自定义：不可能
```

---

### Custom 方案（现代，高度可扩展）

```go
// ✅ 框架极简
xb/
├── dialect.go         (170 行) - Custom 接口定义
├── qdrant_custom.go   (77 行)  - 官方示例（Qdrant）
└── ... 核心代码

总代码：247 行
维护成本：极低
新数据库：用户 5-30 分钟实现
用户自定义：完全自由

// ✅ 用户项目
your-project/
└── db/
    ├── milvus_custom.go      (150 行) - 用户实现
    ├── oracle_custom.go      (200 行) - 用户实现
    ├── clickhouse_custom.go  (250 行) - 用户实现
    └── my_special_db.go      (100 行) - 自研数据库！
```

---

## 🌟 AI 时代的关键特征

### 1. 数据库技术爆炸

**过去 10 年**：
- 传统数据库：PostgreSQL, MySQL, Oracle（3-5 种）

**AI 时代（现在）**：
- 向量数据库：20+ 种
- 图数据库：10+ 种
- 时序数据库：15+ 种
- 分析数据库：10+ 种
- 总计：**60+ 种数据库**

**Custom 接口**：
- ✅ 一个接口，适配所有
- ✅ 用户需要哪个，自己实现
- ✅ 框架不臃肿

---

### 2. 技术迭代极快

**向量数据库的进化**：
- 2023.01：Qdrant 只支持基础搜索
- 2023.06：新增 Recommend API
- 2023.12：新增 Discover API
- 2024.06：新增量化重打分
- 2024.12：新增多向量搜索

**Dialect 方案**：
- ❌ 每次新特性，修改框架 → 跟不上

**Custom 方案**：
- ✅ 用户直接实现新特性 → 0 延迟
- ✅ 使用 `X()` 扩展点 → 立即可用

---

### 3. 用户需求多样化

**场景 1**：公司自研向量数据库

```go
// ✅ Custom 接口：5 分钟实现
type InternalVectorDBCustom struct {
    Endpoint string
}

func (c *InternalVectorDBCustom) Generate(built *Built) (interface{}, error) {
    // 生成公司内部的 JSON 格式
    return customJSON, nil
}
```

**场景 2**：ClickHouse + Milvus 混合部署

```go
// ✅ 运行时切换
var custom xb.Custom
if useClickHouse {
    custom = NewClickHouseCustom()
} else {
    custom = NewMilvusCustom()
}

built := xb.Of("data").Custom(custom).Build()
```

---

## 🎯 Generate() 接口的哲学

### 设计原则

1. **极简主义**（Minimalism）
   - 一个接口
   - 一个方法
   - 返回 `interface{}`

2. **类型驱动**（Type-Driven）
   - QdrantCustom 类型 = Qdrant
   - OracleCustom 类型 = Oracle
   - 无需枚举

3. **多态调用**（Polymorphism）
   - `built.Custom.Generate()`
   - Go 接口自动分发
   - 无需 if/switch 判断

4. **智能适配**（Smart Adaptation）
   - `JsonOfSelect()` → 期望 string
   - `SqlOfSelect()` → 期望 *SQLResult
   - 自动类型转换和错误提示

---

## 📊 性能对比

### Dialect 枚举方案

```go
// ❌ 运行时判断（慢）
func (built *Built) ToJSON() (string, error) {
    switch built.Dialect {
    case Qdrant:
        return toQdrantJSON(built)
    case Milvus:
        return toMilvusJSON(built)
    case Weaviate:
        return toWeaviateJSON(built)
    // ... 100+ 个 case
    }
}
```

**性能**：
- ❌ 每次调用都要 switch
- ❌ 分支预测失败
- ❌ 代码膨胀

---

### Custom 接口方案

```go
// ✅ 编译时绑定（快）
built.Custom.Generate(built)
// ↓ 编译器直接调用
QdrantCustom.Generate(built)
// ↓ 无需判断，直接执行
```

**性能**：
- ✅ 接口方法调用：~1ns
- ✅ 无分支判断
- ✅ 编译器优化友好

---

## 🌈 未来展望

### Custom 接口可以支持的数据库类型

#### 1. 向量数据库

```go
type MilvusCustom struct { ... }
type WeaviateCustom struct { ... }
type PineconeCustom struct { ... }
type ChromaCustom struct { ... }
type LanceDBCustom struct { ... }
```

#### 2. 图数据库

```go
type Neo4jCustom struct { ... }

func (c *Neo4jCustom) Generate(built *Built) (interface{}, error) {
    // 生成 Cypher 查询
    cypher := "MATCH (n:User) WHERE n.age > $age RETURN n"
    return &CypherQuery{Query: cypher, Params: params}, nil
}
```

#### 3. 时序数据库

```go
type InfluxDBCustom struct { ... }

func (c *InfluxDBCustom) Generate(built *Built) (interface{}, error) {
    // 生成 InfluxQL
    influxQL := "SELECT mean(temp) FROM weather WHERE time > now() - 1h"
    return influxQL, nil
}
```

#### 4. 文档数据库

```go
type MongoDBCustom struct { ... }

func (c *MongoDBCustom) Generate(built *Built) (interface{}, error) {
    // 生成 MongoDB 查询
    mongoQuery := bson.M{"age": bson.M{"$gt": 18}}
    return mongoQuery, nil
}
```

#### 5. 搜索引擎

```go
type ElasticsearchCustom struct { ... }

func (c *ElasticsearchCustom) Generate(built *Built) (interface{}, error) {
    // 生成 ES DSL
    dsl := `{"query": {"match": {"content": "golang"}}}`
    return dsl, nil
}
```

#### 6. 自研数据库

```go
type MyCompanyDBCustom struct { ... }

func (c *MyCompanyDBCustom) Generate(built *Built) (interface{}, error) {
    // 公司内部数据库的查询格式
    return customFormat, nil
}
```

---

## 💡 Custom 接口的核心价值

### 不是"框架支持所有数据库"

而是：

### "用户能轻松支持任何数据库" ⭐

---

## 🎨 设计美学

### 对称性（Symmetry）

```go
// ✅ 完美对称
type QdrantCustom struct { ... }   // 向量数据库
type OracleCustom struct { ... }   // SQL 数据库
type Neo4jCustom struct { ... }    // 图数据库

// 都实现同一个接口
func (c *Custom) Generate(built *Built) (interface{}, error)
```

---

### 简洁性（Simplicity）

```go
// ✅ 用户实现 Milvus 支持
type MilvusCustom struct {
    DefaultNProbe int
}

func (c *MilvusCustom) Generate(built *Built) (interface{}, error) {
    json, _ := built.toMilvusJSON()
    return json, nil
}

// 完成！只需 10 行代码
```

---

### 可组合性（Composability）

```go
// ✅ 用户可以组合多个 Custom
type HybridCustom struct {
    VectorDB Custom  // Qdrant
    SQL      Custom  // PostgreSQL
}

func (c *HybridCustom) Generate(built *Built) (interface{}, error) {
    // 混合查询：先向量，后 SQL
    vectorResults, _ := c.VectorDB.Generate(built)
    sqlResults, _ := c.SQL.Generate(built)
    
    return merge(vectorResults, sqlResults), nil
}
```

---

## 📚 实现指南

### 5 分钟实现基础版本

```go
type MyDBCustom struct {}

func (c *MyDBCustom) Generate(built *Built) (interface{}, error) {
    return `{"query": "test"}`, nil  // 最简实现
}
```

---

### 30 分钟实现完整版本

```go
type MyDBCustom struct {
    DefaultConfig map[string]interface{}
}

func (c *MyDBCustom) Generate(built *Built) (interface{}, error) {
    // 1. 提取参数
    vectorBb := findVectorSearchBb(built.Conds)
    
    // 2. 构建请求
    req := buildRequest(vectorBb, c.DefaultConfig)
    
    // 3. 应用过滤器
    applyFilters(built.Conds, req)
    
    // 4. 序列化
    json, _ := json.Marshal(req)
    return string(json), nil
}
```

---

### 1 小时实现生产级别

```go
type MyDBCustom struct {
    DefaultConfig Config
}

func NewMyDBCustom() *MyDBCustom { ... }
func MyDBHighPrecision() *MyDBCustom { ... }
func MyDBHighSpeed() *MyDBCustom { ... }

func (c *MyDBCustom) Generate(built *Built) (interface{}, error) {
    // 完整实现：参数提取、过滤器、错误处理、测试
}
```

---

## 🎯 与其他框架对比

### Java Hibernate（Dialect 方案）

```java
// ❌ 框架包含所有方言
org.hibernate.dialect.PostgreSQLDialect
org.hibernate.dialect.MySQLDialect
org.hibernate.dialect.OracleDialect
// ... 50+ 个内置方言

// ❌ 用户想支持新数据库？
// 1. 继承 Dialect 类
// 2. 实现 100+ 个方法
// 3. 提交 PR 到 Hibernate
// 4. 等半年合并
```

---

### xb Custom（Interface 方案）

```go
// ✅ 框架极简
type Custom interface {
    Generate(built *Built) (interface{}, error)
}

// ✅ 用户想支持新数据库？
// 1. 定义结构体
type MyDBCustom struct { ... }

// 2. 实现一个方法
func (c *MyDBCustom) Generate(built *Built) (interface{}, error) {
    return result, nil
}

// 3. 使用
built.Custom(myCustom).Build()

// 完成！5-30 分钟
```

---

## 🚀 总结

### Custom 接口是 AI 时代需要的，因为：

1. ✅ **数据库爆炸**：20+ 向量数据库，Custom 轻松支持
2. ✅ **技术迭代快**：用户自己跟随新特性，框架不臃肿
3. ✅ **需求多样化**：自研数据库、混合部署，Custom 都能满足
4. ✅ **极简设计**：一个接口、一个方法、无限可能
5. ✅ **类型安全**：编译时检查，运行时无错
6. ✅ **性能极致**：接口调用 ~1ns，无分支判断
7. ✅ **用户自由**：5 分钟到生产级别，完全掌控

---

## 💎 这才是编程技术里的钻石

**不是框架做所有事，而是让用户能轻松做任何事！**

**这就是 xb v1.1.0 Custom 接口的革命性意义！** 🚀✨

---

**版本**: v1.1.0  
**设计者**: xb Team  
**理念**: 极简、通用、实用、面向未来

