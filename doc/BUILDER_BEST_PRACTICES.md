# xb Builder 使用最佳实践

## 📋 概述

本文档提供 xb Builder 的最佳使用实践，帮助开发者避免常见错误，编写高效、可维护的查询代码。

---

## 🎯 核心原则

### 1. Builder 是一次性的

```go
// ✅ 正确：每次查询创建新的 Builder
func GetUser(id int64) (*User, error) {
    sql, args, _ := xb.Of(&User{}).
        Eq("id", id).
        Build().
        SqlOfSelect()
    
    var user User
    err := db.Get(&user, sql, args...)
    return &user, err
}

// ❌ 错误：不要复用 Builder
var userBuilder = xb.Of(&User{}) // 不要这样做！

func GetUser1() {
    userBuilder.Eq("id", 1).Build() // 危险！
}

func GetUser2() {
    userBuilder.Eq("id", 2).Build() // 危险！会包含之前的条件
}
```

**原因**：Builder 会累积条件，复用会导致条件叠加。

---

### 2. 不要在多个 goroutine 间共享 Builder

```go
// ✅ 正确：每个 goroutine 创建自己的 Builder
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    
    // 每次请求创建新的 Builder
    sql, args, _ := xb.Of(&User{}).
        Eq("id", id).
        Build().
        SqlOfSelect()
    
    // ...
}

// ❌ 错误：不要共享 Builder
var sharedBuilder = xb.Of(&User{})

func Handler1() {
    sharedBuilder.Eq("status", "active").Build() // 危险！
}

func Handler2() {
    sharedBuilder.Eq("status", "pending").Build() // 危险！
}
```

**原因**：Builder 不是并发安全的，这是设计上的选择，以保持简洁高效。

---

### 3. 充分利用自动过滤

```go
// ✅ 正确：直接传递参数，让 xb 自动过滤
func SearchUsers(username string, minAge int, status string) {
    builder := xb.Of(&User{}).
        Like("username", username).  // username 为空时自动忽略
        Gte("age", minAge).          // minAge 为 0 时自动忽略
        Eq("status", status)         // status 为空时自动忽略
    
    sql, args, _ := builder.Build().SqlOfSelect()
    // ...
}

// ❌ 错误：不要手动检查 nil/0/空字符串
func SearchUsers(username string, minAge int, status string) {
    builder := xb.Of(&User{})
    
    // 不需要这些判断！
    if username != "" {
        builder.Like("username", username)
    }
    if minAge > 0 {
        builder.Gte("age", minAge)
    }
    if status != "" {
        builder.Eq("status", status)
    }
    
    // ...
}
```

**原因**：sqlxb 有 9 层自动过滤机制，会自动忽略 nil/0/空字符串。

---

## 🔧 常见场景最佳实践

### 1. 简单查询

```go
// 获取单条记录
func GetUser(id int64) (*User, error) {
    sql, args, _ := xb.Of(&User{}).
        Eq("id", id).
        Build().
        SqlOfSelect()
    
    var user User
    err := db.Get(&user, sql, args...)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// 获取列表
func ListUsers(status string, limit int) ([]*User, error) {
    sql, args, _ := xb.Of(&User{}).
        Eq("status", status).
        Limit(limit).
        Build().
        SqlOfSelect()
    
    var users []*User
    err := db.Select(&users, sql, args...)
    return users, err
}
```

---

### 2. 复杂条件查询

```go
// OR 条件
func SearchUsers(keyword string) ([]*User, error) {
    sql, args, _ := xb.Of(&User{}).
        Or(func(cb *xb.CondBuilder) {
            cb.Like("username", keyword).
               OR().
               Like("email", keyword)
        }).
        Build().
        SqlOfSelect()
    
    var users []*User
    err := db.Select(&users, sql, args...)
    return users, err
}

// 复杂嵌套条件
func AdvancedSearch(params SearchParams) ([]*User, error) {
    builder := xb.Of(&User{})
    
    // 基础条件
    builder.Eq("status", params.Status)
    
    // 年龄范围
    if params.MinAge > 0 || params.MaxAge > 0 {
        builder.And(func(cb *xb.CondBuilder) {
            cb.Gte("age", params.MinAge).
               Lte("age", params.MaxAge)
        })
    }
    
    // 关键词搜索
    if params.Keyword != "" {
        builder.Or(func(cb *xb.CondBuilder) {
            cb.Like("username", params.Keyword).
               OR().
               Like("email", params.Keyword)
        })
    }
    
    sql, args, _ := builder.Build().SqlOfSelect()
    var users []*User
    err := db.Select(&users, sql, args...)
    return users, err
}
```

---

### 3. 分页查询

```go
// Web 分页（带 COUNT）
func PagedUsers(page, rows int) ([]*User, int64, error) {
    builder := xb.Of(&User{}).
        Eq("status", "active").
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(int64(page)).Rows(int64(rows))
        })
    
    countSql, dataSql, args, _ := builder.Build().SqlOfPage()
    
    // 获取总数
    var total int64
    if countSql != "" {
        db.Get(&total, countSql)
    }
    
    // 获取数据
    var users []*User
    err := db.Select(&users, dataSql, args...)
    
    return users, total, err
}

// 简单分页（无 COUNT）
func ListUsers(limit, offset int) ([]*User, error) {
    sql, args, _ := xb.Of(&User{}).
        Limit(limit).
        Offset(offset).
        Build().
        SqlOfSelect()
    
    var users []*User
    err := db.Select(&users, sql, args...)
    return users, err
}
```

---

### 4. 向量检索

```go
// 基础向量检索
func SearchSimilarDocs(queryVector []float32, limit int) ([]*Document, error) {
    sql, args, _ := xb.Of(&Document{}).
        VectorSearch("embedding", queryVector, limit).
        Build().
        SqlOfVectorSearch()
    
    var docs []*Document
    err := db.Select(&docs, sql, args...)
    return docs, err
}

// 混合检索（向量 + 标量过滤）
func HybridSearch(queryVector []float32, docType string, limit int) ([]*Document, error) {
    sql, args, _ := xb.Of(&Document{}).
        VectorSearch("embedding", queryVector, limit).
        Eq("doc_type", docType).
        Ne("status", "deleted").
        Build().
        SqlOfVectorSearch()
    
    var docs []*Document
    err := db.Select(&docs, sql, args...)
    return docs, err
}
```

---

### 5. Qdrant 查询

```go
// 基础 Qdrant 查询
func QdrantSearch(queryVector []float32) (string, error) {
    built := xb.Of(&Document{}).
        VectorSearch("embedding", queryVector, 20).
        Eq("doc_type", "article").
        Build()
    
    jsonBytes, err := built.ToQdrantJSON()
    if err != nil {
        return "", err
    }
    
    return string(jsonBytes), nil
}

// 高级 Qdrant 查询
func QdrantAdvancedSearch(queryVector []float32) (string, error) {
    built := xb.Of(&Document{}).
        VectorSearch("embedding", queryVector, 20).
        Eq("language", "zh").
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.ScoreThreshold(0.8).
               HnswEf(128).
               WithVector(true)
        }).
        Build()
    
    jsonBytes, err := built.ToQdrantJSON()
    if err != nil {
        return "", err
    }
    
    return string(jsonBytes), nil
}
```

---

## ⚠️ 常见错误

### 1. 复用 Builder 导致条件累积

```go
// ❌ 错误
var baseBuilder = xb.Of(&User{}).Eq("status", "active")

func GetUser1() {
    sql, _, _ := baseBuilder.Eq("id", 1).Build().SqlOfSelect()
    // WHERE status = ? AND id = ?
}

func GetUser2() {
    sql, _, _ := baseBuilder.Eq("id", 2).Build().SqlOfSelect()
    // WHERE status = ? AND id = ? AND id = ? ❌ 条件累积了！
}

// ✅ 正确
func GetUser1() {
    sql, _, _ := xb.Of(&User{}).
        Eq("status", "active").
        Eq("id", 1).
        Build().
        SqlOfSelect()
}

func GetUser2() {
    sql, _, _ := xb.Of(&User{}).
        Eq("status", "active").
        Eq("id", 2).
        Build().
        SqlOfSelect()
}
```

---

### 2. 手动添加 Like 通配符

```go
// ❌ 错误
builder.Like("username", "%"+username+"%") // 会变成 %%username%%

// ✅ 正确
builder.Like("username", username) // 自动添加 %，变成 %username%

// 前缀匹配
builder.LikeLeft("username", username) // 变成 username%
```

---

### 3. 不必要的 nil/0 检查

```go
// ❌ 错误：不需要手动检查
if username != "" {
    builder.Like("username", username)
}
if age > 0 {
    builder.Gte("age", age)
}

// ✅ 正确：直接传递，自动过滤
builder.Like("username", username).
        Gte("age", age)
```

---

### 4. 在事务中错误地使用 Builder

```go
// ✅ 正确：在事务中使用 Builder
func TransferBalance(fromID, toID int64, amount float64) error {
    tx, err := db.Beginx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 每个操作创建新的 Builder
    sql1, args1, _ := xb.Of(&Account{}).
        Update(func(ub *xb.UpdateBuilder) {
            ub.Set("balance", "balance - ?", amount)
        }).
        Eq("id", fromID).
        Build().
        SqlOfUpdate()
    
    _, err = tx.Exec(sql1, args1...)
    if err != nil {
        return err
    }
    
    sql2, args2, _ := xb.Of(&Account{}).
        Update(func(ub *xb.UpdateBuilder) {
            ub.Set("balance", "balance + ?", amount)
        }).
        Eq("id", toID).
        Build().
        SqlOfUpdate()
    
    _, err = tx.Exec(sql2, args2...)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

---

## 💡 性能优化建议

### 1. 只查询需要的字段

```go
// 查询所有字段（默认）
builder := xb.Of(&User{})

// 只查询部分字段
builder := xb.Of(&User{}).
    Select("id", "username", "email")
```

---

### 2. 使用 Limit 避免大结果集

```go
// ✅ 好
builder.Limit(100)

// ❌ 不好：可能返回数百万条记录
builder // 没有 Limit
```

---

### 3. 为高频查询创建辅助函数

```go
// 封装常用查询
func ActiveUsers() *xb.BuilderX {
    return xb.Of(&User{}).Eq("status", "active")
}

// 使用
func GetActiveUser(id int64) (*User, error) {
    sql, args, _ := ActiveUsers().
        Eq("id", id).
        Build().
        SqlOfSelect()
    
    var user User
    err := db.Get(&user, sql, args...)
    return &user, err
}
```

---

## 📝 代码组织建议

### 1. 将查询逻辑放在 Repository 层

```go
type UserRepository struct {
    db *sqlx.DB
}

func (r *UserRepository) GetByID(id int64) (*User, error) {
    sql, args, _ := xb.Of(&User{}).
        Eq("id", id).
        Build().
        SqlOfSelect()
    
    var user User
    err := r.db.Get(&user, sql, args...)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Search(params SearchParams) ([]*User, error) {
    builder := xb.Of(&User{}).
        Like("username", params.Username).
        Gte("age", params.MinAge).
        Eq("status", params.Status)
    
    sql, args, _ := builder.Build().SqlOfSelect()
    
    var users []*User
    err := r.db.Select(&users, sql, args...)
    return users, err
}
```

---

### 2. 使用参数对象而非多个参数

```go
// ✅ 好
type SearchParams struct {
    Username string
    MinAge   int
    MaxAge   int
    Status   string
    Page     int
    Rows     int
}

func SearchUsers(params SearchParams) ([]*User, error) {
    // ...
}

// ❌ 不好
func SearchUsers(username string, minAge, maxAge int, status string, page, rows int) ([]*User, error) {
    // 参数太多
}
```

---

## 🔍 调试技巧

### 1. 打印生成的 SQL

```go
sql, args, _ := builder.Build().SqlOfSelect()

// 调试时打印
fmt.Printf("SQL: %s\n", sql)
fmt.Printf("Args: %v\n", args)

// 生产环境使用日志
log.Printf("SQL: %s, Args: %v", sql, args)
```

---

### 2. 检查参数绑定

```go
sql, args, _ := builder.Build().SqlOfSelect()

// 确保参数数量匹配占位符数量
placeholders := strings.Count(sql, "?")
if len(args) != placeholders {
    log.Printf("Warning: %d placeholders but %d args", placeholders, len(args))
}
```

---

## 📚 相关文档

- [README](../README.md) - xb 基础用法
- [VECTOR_QUICKSTART](./VECTOR_QUICKSTART.md) - 向量数据库快速开始
- [QDRANT_X_USAGE](./QDRANT_X_USAGE.md) - QdrantX 使用指南
- [AI Application Docs](./ai_application/README.md) - AI 应用集成

---

**最后更新**: 2025-02-27  
**版本**: v0.10.3

