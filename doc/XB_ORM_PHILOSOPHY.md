# xb 与 ORM 的设计哲学

## 🎯 核心理念

> **xb 不是 ORM，而是各种 ORM 的完美补充**

**xb 的定位**：
- ✅ 提供强大的 SQL/JSON 构建能力
- ✅ 提供必要的 Meta 数据（字段映射）
- ✅ 让用户自己完成 ORM 扫描
- ✅ 作为各种 ORM 框架的底层引擎

---

## 💎 为什么不做完整的 ORM？

### 问题：ORM 框架的局限性

#### 1. **简单查询容易，复杂查询困难**

```go
// ❌ ORM 框架的问题
// 简单查询（容易）
db.Where("age > ?", 18).Find(&users)

// 复杂查询（困难）
// - 多表 JOIN？
// - 子查询？
// - 聚合函数？
// - 窗口函数？
// - CTE（WITH 子句）？
// → 最终还是写 Raw SQL！
```

#### 2. **各种 ORM 框架百花齐放**

| ORM 框架 | 特点 | 用户数 |
|---------|------|--------|
| GORM | 功能丰富 | 最多 |
| XORM | 轻量级 | 很多 |
| sqlx | 半 ORM | 很多 |
| ent | 类型安全 | 增长中 |
| sqlc | 代码生成 | 增长中 |
| 自研 ORM | 公司特定 | 很多 |

**xb 无法取代所有 ORM，也不应该！**

---

## 🚀 xb 的解决方案

### xb = SQL Builder + Meta 数据提供者

```go
// ============================================================================
// xb 提供的能力
// ============================================================================

// 1. 强大的 SQL 构建
built := xb.Of("users").As("u").
    Select("u.id", "u.name", "o.order_id", "o.amount AS total_amount").
    FromX(func(fb *xb.FromBuilder) {
        fb.JOIN(xb.INNER).Of("orders").As("o").
            On("o.user_id = u.id")
    }).
    Eq("u.status", "active").
    Gt("o.amount", 100).
    Paged(func(pb *xb.PageBuilder) {
        pb.Page(2).Rows(10)
    }).
    Build()

// 2. 生成 SQL + Args + Meta
sql, args, meta := built.SqlOfSelect()

// SQL: 
// SELECT u.id AS c0, u.name AS c1, o.order_id AS c2, o.amount AS total_amount
// FROM users u
// INNER JOIN orders o ON o.user_id = u.id
// WHERE u.status = ? AND o.amount > ?
// LIMIT 10 OFFSET 10

// Args: [active, 100]

// Meta（⭐ 关键）: 
// map[c0:u.id c1:u.name c2:o.order_id total_amount:total_amount]

// ============================================================================
// 用户自己完成 ORM 扫描
// ============================================================================

rows, _ := db.Query(sql, args...)
defer rows.Close()

// 方式 1: 使用 sqlx（半 ORM）
var results []map[string]interface{}
sqlx.StructScan(rows, &results)

// 方式 2: 使用 GORM
db.Raw(sql, args...).Scan(&users)

// 方式 3: 手动扫描（使用 Meta 映射）
for rows.Next() {
    var c0, c1, c2 interface{}
    var totalAmount float64
    rows.Scan(&c0, &c1, &c2, &totalAmount)
    
    // ⭐ 使用 Meta 映射字段
    userId := c0      // meta[c0] = u.id
    userName := c1    // meta[c1] = u.name
    orderId := c2     // meta[c2] = o.order_id
    // total_amount 是显式别名，直接使用
}
```

---

## 💡 xb 提供的核心价值

### 1. **Meta Map：ORM 扫描的依据**

```go
// xb 生成的 Meta
meta := map[string]string{
    "c0":           "u.id",         // 自动生成别名
    "c1":           "u.name",       // 自动生成别名
    "c2":           "o.order_id",   // 自动生成别名
    "total_amount": "total_amount", // 显式别名
}

// ORM 使用 Meta 映射结果
type UserOrder struct {
    UserId      int64   `db:"c0" json:"user_id"`       // ⭐ 使用别名
    UserName    string  `db:"c1" json:"user_name"`     // ⭐ 使用别名
    OrderId     int64   `db:"c2" json:"order_id"`      // ⭐ 使用别名
    TotalAmount float64 `db:"total_amount" json:"total_amount"`
}
```

---

### 2. **复杂查询的构建能力**

| 查询类型 | xb 支持 | ORM 框架 |
|---------|---------|---------|
| 简单查询 | ✅ | ✅ |
| 多表 JOIN | ✅ 流畅 API | ⚠️ 困难 |
| 子查询 | ✅ 嵌套 Builder | ⚠️ Raw SQL |
| 聚合函数 | ✅ Agg() | ⚠️ 有限支持 |
| 窗口函数 | ✅ X() 扩展 | ❌ Raw SQL |
| 向量检索 | ✅ VectorSearch() | ❌ 不支持 |
| 分页优化 | ✅ Paged() | ⚠️ 基础支持 |
| 数据库方言 | ✅ Custom 接口 | ⚠️ 有限支持 |

---

### 3. **数据库适配能力**

```go
// ⭐ xb 提供 SQL + Meta，ORM 负责扫描
var custom xb.Custom

switch config.Database {
case "oracle":
    custom = oracle_custom.New()
case "mysql":
    custom = xb.MySQLWithUpsert()
case "postgresql":
    custom = nil  // 默认即可
}

built := xb.Of("users").
    Custom(custom).
    Select("u.id", "u.name").
    Eq("status", "active").
    Build()

sql, args, meta := built.SqlOfSelect()

// ⭐ 使用你喜欢的 ORM
gorm.DB.Raw(sql, args...).Scan(&users)    // GORM
sqlx.Select(&users, sql, args...)         // sqlx
自研ORM.Scan(sql, args, meta, &users)     // 自研 ORM
```

---

## 🎨 设计对比

### 完整 ORM 方案（如 GORM）

```go
// ❌ 问题：复杂查询困难
db.Model(&User{}).
    Select("users.id, users.name, orders.amount").
    Joins("INNER JOIN orders ON orders.user_id = users.id").
    Where("users.status = ?", "active").
    Where("orders.amount > ?", 100).
    Offset(10).
    Limit(10).
    Find(&results)

// 问题：
// 1. JOIN 语法别扭
// 2. 子查询几乎不可能
// 3. 窗口函数不支持
// 4. 向量检索不支持
// 5. 最终还是写 Raw SQL
```

---

### xb + ORM 方案（推荐）⭐

```go
// ✅ xb 构建 SQL
built := xb.Of("users").As("u").
    Select("u.id", "u.name", "o.amount AS total_amount").
    FromX(func(fb *xb.FromBuilder) {
        fb.JOIN(xb.INNER).Of("orders").As("o").
            On("o.user_id = u.id")
    }).
    Eq("u.status", "active").
    Gt("o.amount", 100).
    Paged(func(pb *xb.PageBuilder) {
        pb.Page(2).Rows(10)
    }).
    Build()

sql, args, meta := built.SqlOfSelect()

// ✅ ORM 扫描结果（你喜欢哪个用哪个）
db.Raw(sql, args...).Scan(&results)  // GORM
sqlx.Select(&results, sql, args...)  // sqlx

// 优势：
// 1. ✅ SQL 构建流畅
// 2. ✅ 支持任意复杂查询
// 3. ✅ Meta 提供字段映射
// 4. ✅ ORM 专注扫描（各自优势）
```

---

## 📊 xb 在 ORM 生态中的定位

```
┌─────────────────────────────────────────┐
│          应用层（业务代码）              │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│         ORM 层（结果扫描）              │
│  GORM / sqlx / XORM / ent / 自研       │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│      xb 层（SQL/JSON 构建 + Meta）      │  ⭐
│  - 复杂 SQL 构建                        │
│  - 向量检索支持                         │
│  - 数据库方言适配                       │
│  - Meta 数据提供                        │
└─────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────┐
│       数据库驱动层（database/sql）       │
│  MySQL / PostgreSQL / Oracle ...       │
└─────────────────────────────────────────┘
```

---

## 💡 实战案例

### 案例 1: xb + GORM

```go
package main

import (
    "github.com/fndome/xb"
    "github.com/fndome/xb/oracle_custom"
    "gorm.io/gorm"
)

type UserOrder struct {
    UserId      int64   `db:"c0" json:"user_id"`
    UserName    string  `db:"c1" json:"user_name"`
    OrderId     int64   `db:"c2" json:"order_id"`
    TotalAmount float64 `db:"total_amount" json:"total_amount"`
}

func GetUserOrders(db *gorm.DB, page, rows int) ([]UserOrder, error) {
    // ⭐ xb 构建复杂 SQL
    built := xb.Of("users").As("u").
        Custom(oracle_custom.New()).  // Oracle 分页
        Select("u.id", "u.name", "o.order_id", "o.amount AS total_amount").
        FromX(func(fb *xb.FromBuilder) {
            fb.JOIN(xb.INNER).Of("orders").As("o").
                On("o.user_id = u.id")
        }).
        Eq("u.status", "active").
        Gt("o.amount", 100).
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(uint(page)).Rows(uint(rows))
        }).
        Build()
    
    countSQL, dataSQL, args, meta := built.SqlOfPage()
    
    // ⭐ 使用 Meta 映射（可选，用于调试或动态映射）
    _ = meta  // map[c0:u.id c1:u.name c2:o.order_id total_amount:total_amount]
    
    // ⭐ GORM 扫描结果
    var results []UserOrder
    err := db.Raw(dataSQL, args...).Scan(&results).Error
    
    return results, err
}
```

---

### 案例 2: xb + sqlx

```go
package main

import (
    "github.com/fndome/xb"
    "github.com/jmoiron/sqlx"
)

func GetUserOrders(db *sqlx.DB, page, rows int) ([]UserOrder, error) {
    // ⭐ xb 构建 SQL
    built := xb.Of("users").As("u").
        Select("u.id", "u.name", "o.order_id", "o.amount AS total_amount").
        FromX(func(fb *xb.FromBuilder) {
            fb.JOIN(xb.INNER).Of("orders").As("o").
                On("o.user_id = u.id")
        }).
        Eq("u.status", "active").
        Build()
    
    sql, args, meta := built.SqlOfSelect()
    
    _ = meta  // Meta 数据可用于动态映射
    
    // ⭐ sqlx 扫描结果
    var results []UserOrder
    err := db.Select(&results, sql, args...)
    
    return results, err
}
```

---

### 案例 3: xb + 自研 ORM（使用 Meta）

```go
package main

import (
    "database/sql"
    "github.com/fndome/xb"
)

// 自研 ORM：利用 xb 的 Meta map
type CustomORM struct {
    db *sql.DB
}

func (orm *CustomORM) Query(built *xb.Built) ([]map[string]interface{}, error) {
    sql, args, meta := built.SqlOfSelect()
    
    rows, err := orm.db.Query(sql, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // ⭐ 使用 Meta 映射字段
    columns, _ := rows.Columns()
    results := []map[string]interface{}{}
    
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        rows.Scan(valuePtrs...)
        
        // ⭐ 使用 Meta 映射结果
        result := make(map[string]interface{})
        for i, col := range columns {
            // ⭐ 如果 Meta 中有映射，使用原始字段名
            if originalField, ok := meta[col]; ok {
                result[originalField] = values[i]  // u.id 而不是 c0
            } else {
                result[col] = values[i]
            }
        }
        
        results = append(results, result)
    }
    
    return results, nil
}

// 使用
func main() {
    orm := &CustomORM{db: db}
    
    built := xb.Of("users").As("u").
        Select("u.id", "u.name").
        Build()
    
    results, _ := orm.Query(built)
    
    // results[0]["u.id"]   // ⭐ 使用原始字段名（通过 Meta 映射）
    // results[0]["u.name"] // ⭐ 而不是 c0, c1
}
```

---

## 🌟 Meta Map 的核心价值

### Meta Map 是什么？

**字段别名到原始字段的映射表**

```go
// SQL:
SELECT u.id AS c0, u.name AS c1, o.amount AS total_amount

// Meta Map:
map[string]string{
    "c0":           "u.id",         // 自动别名 → 原始字段
    "c1":           "u.name",       // 自动别名 → 原始字段
    "total_amount": "total_amount", // 显式别名 → 显式别名
}
```

---

### Meta Map 的使用场景

#### 场景 1: 动态字段映射

```go
// 运行时不知道查询了哪些字段
sql, args, meta := built.SqlOfSelect()

// ⭐ 使用 Meta 动态映射
for rows.Next() {
    // 扫描到别名字段 c0, c1, c2
    // 通过 Meta 知道：c0 = u.id, c1 = u.name
}
```

---

#### 场景 2: 跨数据库字段映射

```go
// Oracle: 自动生成 c0, c1, c2
// MySQL: 可能直接使用 u.id, u.name（不同驱动行为）
// ⭐ Meta 统一了映射规则
```

---

#### 场景 3: 自研 ORM 的基石

```go
// 自研 ORM 可以基于 Meta 实现：
// 1. 字段到结构体的映射
// 2. JSON 序列化时的字段名
// 3. 动态查询结果处理
```

---

## 📖 完整示例：xb + 自研 ORM

### 自研 ORM 实现

```go
package myorm

import (
    "database/sql"
    "github.com/fndome/xb"
)

// Scanner ORM 扫描器
type Scanner struct {
    db *sql.DB
}

// ScanToMap 扫描到 map（使用 Meta）
func (s *Scanner) ScanToMap(built *xb.Built) ([]map[string]interface{}, error) {
    sql, args, meta := built.SqlOfSelect()
    
    rows, err := s.db.Query(sql, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    columns, _ := rows.Columns()
    results := []map[string]interface{}{}
    
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        rows.Scan(valuePtrs...)
        
        // ⭐ 使用 Meta 映射
        result := make(map[string]interface{})
        for i, col := range columns {
            if originalField, ok := meta[col]; ok {
                // 使用原始字段名
                result[originalField] = values[i]
            } else {
                result[col] = values[i]
            }
        }
        
        results = append(results, result)
    }
    
    return results, nil
}

// ScanToStruct 扫描到结构体（使用 Meta + 反射）
func (s *Scanner) ScanToStruct(built *xb.Built, dest interface{}) error {
    sql, args, meta := built.SqlOfSelect()
    
    // ⭐ Meta 用于字段到结构体的映射
    // 实现细节省略...
    
    return nil
}
```

---

### 使用自研 ORM

```go
func main() {
    scanner := &myorm.Scanner{db: db}
    
    // ⭐ xb 构建复杂查询
    built := xb.Of("users").As("u").
        Select("u.id", "u.name", "o.order_id", "o.amount AS total").
        FromX(func(fb *xb.FromBuilder) {
            fb.JOIN(xb.INNER).Of("orders").As("o").
                On("o.user_id = u.id")
        }).
        Eq("u.status", "active").
        Build()
    
    // ⭐ 自研 ORM 扫描（利用 Meta）
    results, _ := scanner.ScanToMap(built)
    
    // results[0]["u.id"]      // ✅ 原始字段名
    // results[0]["u.name"]    // ✅ 原始字段名
    // results[0]["o.order_id"] // ✅ 原始字段名
    // results[0]["total"]     // ✅ 显式别名
}
```

---

## 🎯 总结

### xb 的定位

**不是完整的 ORM，而是各种 ORM 的完美补充**

| 职责 | xb | ORM 框架 |
|------|----|---------| 
| SQL 构建 | ✅ 强大 | ⚠️ 简单查询 OK，复杂困难 |
| Meta 数据 | ✅ 提供 | ❌ 不提供 |
| 结果扫描 | ❌ 不做 | ✅ 专注扫描 |
| 数据库方言 | ✅ Custom 接口 | ⚠️ 有限支持 |
| 向量检索 | ✅ 原生支持 | ❌ 不支持 |

---

### xb 的核心价值

1. ✅ **复杂查询构建**：JOIN、子查询、窗口函数、向量检索
2. ✅ **Meta 数据提供**：字段别名映射，ORM 扫描的依据
3. ✅ **数据库适配**：Custom 接口支持 Oracle/MySQL/ClickHouse/Qdrant 等
4. ✅ **ORM 兼容**：与 GORM/sqlx/XORM/自研 ORM 完美配合

---

## 💎 设计哲学

> **"xb 只提供必要的数据，作为各种 ORM 的补充"**

**这就是 xb 的使命**：
- ✅ 让复杂查询变简单
- ✅ 让数据库适配变容易
- ✅ 让 ORM 专注扫描
- ✅ 让用户自由选择 ORM

**xb = SQL Builder + Meta Provider，而不是完整的 ORM！** 🚀✨

---

**版本**: v1.1.0  
**设计**: xb Team  
**理念**: 极简、专注、补充、不替代

