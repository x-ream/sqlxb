# 自定义 JOIN 扩展指南

## 🎯 概述

`xb` 提供了基础的 JOIN 类型，但你可以扩展自定义 JOIN 以支持：
- 特定数据库的 JOIN 语法（如 ClickHouse 的 `GLOBAL JOIN`, `ASOF JOIN`）
- 业务特定的 JOIN 逻辑
- 性能优化的 JOIN 变体

---

## 📚 内置 JOIN 类型

### xb 已支持的 JOIN

```go
// sqlxb/joins.go

const (
    inner_join      = "INNER JOIN"
    left_join       = "LEFT JOIN"
    right_join      = "RIGHT JOIN"
    cross_join      = "CROSS JOIN"
    asof_join       = "ASOF JOIN"        // ClickHouse
    global_join     = "GLOBAL JOIN"      // ClickHouse 分布式
    full_outer_join = "FULL OUTER JOIN"
)

// JOIN 函数类型
type JOIN func() string

// 内置 JOIN 函数
func NON_JOIN() string { return ", " }
func INNER() string    { return inner_join }
func LEFT() string     { return left_join }
func RIGHT() string    { return right_join }
func CROSS() string    { return cross_join }
func ASOF() string     { return asof_join }
func GLOBAL() string   { return global_join }
func FULL_OUTER() string { return full_outer_join }
```

---

## 🔧 扩展自定义 JOIN

### 方式 1: 简单字符串 JOIN

```go
// your_project/sqlx_ext/custom_joins.go
package sqlx_ext

// LATERAL_JOIN 横向 JOIN（PostgreSQL）
func LATERAL_JOIN() string {
    return "LATERAL JOIN"
}

// ANTI_JOIN 反连接（排除匹配的记录）
func ANTI_JOIN() string {
    return "LEFT JOIN ... WHERE ... IS NULL"
}

// 使用
import (
    "github.com/fndome/xb"
    "your-project/sqlx_ext"
)

// ⭐ 自定义 JOIN 可以直接使用
xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, sqlx_ext.LATERAL_JOIN)
    })
```

---

### 方式 2: 条件 JOIN（参数化）

```go
// your_project/sqlx_ext/conditional_joins.go
package sqlx_ext

// HASH_JOIN 哈希连接（可指定算法）
func HASH_JOIN(algorithm string) xb.JOIN {
    return func() string {
        return fmt.Sprintf("/*+ HASH_JOIN(%s) */ INNER JOIN", algorithm)
    }
}

// INDEX_JOIN 索引连接（指定索引）
func INDEX_JOIN(indexName string) xb.JOIN {
    return func() string {
        return fmt.Sprintf("/*+ INDEX_JOIN(%s) */ INNER JOIN", indexName)
    }
}

// 使用
xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, sqlx_ext.HASH_JOIN("user_idx"))
    })
```

---

### 方式 3: 智能 JOIN（动态选择）

```go
// your_project/sqlx_ext/smart_joins.go
package sqlx_ext

// SmartJoin 根据数据量自动选择 JOIN 类型
func SmartJoin(leftSize, rightSize int64) xb.JOIN {
    return func() string {
        // 小表驱动
        if leftSize < 1000 && rightSize > 1000000 {
            return "INNER JOIN /*+ USE_NL(right_table) */"
        }
        
        // 大表 JOIN 大表
        if leftSize > 1000000 && rightSize > 1000000 {
            return "INNER JOIN /*+ USE_HASH */"
        }
        
        // 默认
        return "INNER JOIN"
    }
}

// 使用
leftCount := getOrderCount()
rightCount := getUserCount()

xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, sqlx_ext.SmartJoin(leftCount, rightCount))
    })
```

---

## 💡 实际案例

### 案例 1: ClickHouse ASOF JOIN

**场景**：时序数据，按时间戳匹配最接近的记录

```go
// ClickHouse 专属 JOIN
package clickhouse_ext

import "github.com/fndome/xb"

// ASOF_LEFT ClickHouse ASOF LEFT JOIN
// 用于时序数据：找到时间戳最接近且不晚于的记录
func ASOF_LEFT() string {
    return "ASOF LEFT JOIN"
}

// ASOF_INNER ClickHouse ASOF INNER JOIN
func ASOF_INNER() string {
    return "ASOF INNER JOIN"
}

// 使用示例：股票交易和订单匹配
type Trade struct {
    ID        int64     `db:"id"`
    Symbol    string    `db:"symbol"`
    Price     float64   `db:"price"`
    Timestamp time.Time `db:"timestamp"`
}

type Order struct {
    ID        int64     `db:"id"`
    Symbol    string    `db:"symbol"`
    OrderTime time.Time `db:"order_time"`
}

func (Trade) TableName() string { return "trades" }
func (Order) TableName() string { return "orders" }

// 查询：找到每个订单时刻最接近的交易价格
sql, args := xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&Trade{}, clickhouse_ext.ASOF_LEFT).
            On(&Order{}, "symbol", &Trade{}, "symbol").         // 连接条件 1
            On(&Order{}, "order_time", &Trade{}, "timestamp")   // 连接条件 2（时间）
    }).
    Select("orders.id, orders.symbol, trades.price").
    Build().
    SqlOfSelect()

// 生成 SQL:
// SELECT orders.id, orders.symbol, trades.price
// FROM orders
// ASOF LEFT JOIN trades
//   ON orders.symbol = trades.symbol
//   AND orders.order_time = trades.timestamp
```

---

### 案例 2: PostgreSQL LATERAL JOIN

**场景**：相关子查询，每行都执行

```go
package postgres_ext

// LATERAL PostgreSQL 横向 JOIN
func LATERAL() string {
    return "LATERAL"
}

// 使用示例：获取每个用户的最近 3 个订单
func getRecentOrders(userIDs []int64) {
    // PostgreSQL LATERAL JOIN 示例
    sql := `
    SELECT u.id, u.name, recent_orders.*
    FROM users u
    LATERAL (
        SELECT o.id, o.amount, o.created_at
        FROM orders o
        WHERE o.user_id = u.id
        ORDER BY o.created_at DESC
        LIMIT 3
    ) AS recent_orders
    WHERE u.id IN (?)
    `
    
    // xb 可能的未来支持：
    // xb.Of(&User{}).
    //     SourceBuilder.From(func(fb *xb.FromBuilder) {
    //         fb.SubQuery(&Order{}, postgres_ext.LATERAL, func(sb *SubQueryBuilder) {
    //             sb.Eq("user_id", fb.Field("id")).
    //                 OrderBy("created_at", DESC).
    //                 Limit(3)
    //         })
    //     })
}
```

---

### 案例 3: 分布式 JOIN（GLOBAL JOIN）

**场景**：ClickHouse 集群，全局 JOIN

```go
package clickhouse_ext

// GLOBAL_INNER ClickHouse 全局 INNER JOIN
// 在分布式环境中，先在每个节点本地 JOIN，再合并
func GLOBAL_INNER() string {
    return "GLOBAL INNER JOIN"
}

// GLOBAL_LEFT ClickHouse 全局 LEFT JOIN
func GLOBAL_LEFT() string {
    return "GLOBAL LEFT JOIN"
}

// 使用
sql, args := xb.Of(&DistributedOrder{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, clickhouse_ext.GLOBAL_INNER).
            On(&DistributedOrder{}, "user_id", &User{}, "id")
    }).
    Build().
    SqlOfSelect()

// 生成 SQL:
// SELECT *
// FROM distributed_orders
// GLOBAL INNER JOIN users
//   ON distributed_orders.user_id = users.id
```

---

## 🎨 高级扩展：JOIN Builder

### 创建 JOIN 构建器

```go
// your_project/sqlx_ext/join_builder.go
package sqlx_ext

import "github.com/fndome/xb"

// JoinBuilderX JOIN 专属构建器
type JoinBuilderX struct {
    joinType string
    hints    []string
}

// NewJoinBuilder 创建 JOIN 构建器
func NewJoinBuilder() *JoinBuilderX {
    return &JoinBuilderX{
        joinType: "INNER JOIN",
        hints:    []string{},
    }
}

// WithHint 添加 JOIN 提示（优化器提示）
func (jb *JoinBuilderX) WithHint(hint string) *JoinBuilderX {
    jb.hints = append(jb.hints, hint)
    return jb
}

// UseHash 使用哈希 JOIN
func (jb *JoinBuilderX) UseHash() *JoinBuilderX {
    return jb.WithHint("USE_HASH")
}

// UseNL 使用嵌套循环 JOIN
func (jb *JoinBuilderX) UseNL() *JoinBuilderX {
    return jb.WithHint("USE_NL")
}

// UseMerge 使用归并 JOIN
func (jb *JoinBuilderX) UseMerge() *JoinBuilderX {
    return jb.WithHint("USE_MERGE")
}

// Build 构建 JOIN 函数
func (jb *JoinBuilderX) Build() xb.JOIN {
    return func() string {
        if len(jb.hints) > 0 {
            hints := strings.Join(jb.hints, ", ")
            return fmt.Sprintf("/*+ %s */ %s", hints, jb.joinType)
        }
        return jb.joinType
    }
}

// 使用示例
joinFunc := NewJoinBuilder().
    UseHash().              // ⭐ 使用哈希 JOIN
    WithHint("PARALLEL").   // ⭐ 并行执行
    Build()

xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, joinFunc)
    })

// 生成 SQL:
// ... FROM orders
// /*+ USE_HASH, PARALLEL */ INNER JOIN users ...
```

---

## 📖 完整示例

### 示例 1: 业务自定义 JOIN

```go
// your_project/business/order_joins.go
package business

import "github.com/fndome/xb"

// ORDER_DETAIL_JOIN 订单详情 JOIN（业务特定）
// 自动过滤已删除的详情
func ORDER_DETAIL_JOIN() xb.JOIN {
    return func() string {
        return `LEFT JOIN order_details 
                ON orders.id = order_details.order_id 
                AND order_details.deleted_at IS NULL`
    }
}

// WITH_VALID_USER 只连接有效用户
func WITH_VALID_USER() xb.JOIN {
    return func() string {
        return `INNER JOIN users 
                ON orders.user_id = users.id 
                AND users.status = 'active'`
    }
}

// 使用
sql, args := xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, business.WITH_VALID_USER).
            From(&OrderDetail{}, business.ORDER_DETAIL_JOIN)
    }).
    Build().
    SqlOfSelect()
```

---

### 示例 2: 性能优化 JOIN

```go
// your_project/performance/optimized_joins.go
package performance

import (
    "github.com/fndome/xb"
    "time"
)

// TimeBasedJoin 根据时间智能选择 JOIN 策略
func TimeBasedJoin(isPeakHour bool) xb.JOIN {
    return func() string {
        if isPeakHour {
            // 高峰期：使用索引 JOIN，减少锁
            return "/*+ INDEX_JOIN */ INNER JOIN"
        } else {
            // 非高峰期：使用哈希 JOIN，更快
            return "/*+ HASH_JOIN */ INNER JOIN"
        }
    }
}

// 使用
isPeak := time.Now().Hour() >= 18 && time.Now().Hour() <= 22

sql, args := xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, performance.TimeBasedJoin(isPeak))
    }).
    Build().
    SqlOfSelect()
```

---

### 示例 3: 数据库特定 JOIN

```go
// your_project/database/mysql_joins.go
package database

import "github.com/fndome/xb"

// STRAIGHT_JOIN MySQL 强制按顺序 JOIN
func STRAIGHT_JOIN() string {
    return "STRAIGHT_JOIN"
}

// FORCE_INDEX MySQL 强制使用索引
func FORCE_INDEX(indexName string) xb.JOIN {
    return func() string {
        return fmt.Sprintf("INNER JOIN FORCE INDEX (%s)", indexName)
    }
}

// 使用
sql, args := xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, database.FORCE_INDEX("idx_user_id"))
    }).
    Build().
    SqlOfSelect()

// 生成 SQL:
// SELECT * FROM orders
// INNER JOIN FORCE INDEX (idx_user_id) users
//   ON orders.user_id = users.id
```

---

## 🏗️ 高级扩展：JOIN 构建器

### 完整的 JOIN 构建器实现

```go
// your_project/sqlx_ext/join_builder_x.go
package sqlx_ext

import (
    "fmt"
    "github.com/fndome/xb"
    "strings"
)

// JoinBuilderX JOIN 配置构建器
type JoinBuilderX struct {
    joinType   string
    hints      []string
    conditions []string
    indexName  string
}

// NewJoin 创建 JOIN 构建器
func NewJoin() *JoinBuilderX {
    return &JoinBuilderX{
        joinType:   "INNER JOIN",
        hints:      []string{},
        conditions: []string{},
    }
}

// Inner 内连接
func (jb *JoinBuilderX) Inner() *JoinBuilderX {
    jb.joinType = "INNER JOIN"
    return jb
}

// Left 左连接
func (jb *JoinBuilderX) Left() *JoinBuilderX {
    jb.joinType = "LEFT JOIN"
    return jb
}

// UseHash 使用哈希 JOIN
func (jb *JoinBuilderX) UseHash() *JoinBuilderX {
    jb.hints = append(jb.hints, "USE_HASH")
    return jb
}

// UseIndex 强制使用索引
func (jb *JoinBuilderX) UseIndex(indexName string) *JoinBuilderX {
    jb.indexName = indexName
    return jb
}

// Parallel 并行执行
func (jb *JoinBuilderX) Parallel(degree int) *JoinBuilderX {
    jb.hints = append(jb.hints, fmt.Sprintf("PARALLEL(%d)", degree))
    return jb
}

// WithCondition 添加额外 JOIN 条件
func (jb *JoinBuilderX) WithCondition(condition string) *JoinBuilderX {
    jb.conditions = append(jb.conditions, condition)
    return jb
}

// Build 构建 JOIN 函数
func (jb *JoinBuilderX) Build() xb.JOIN {
    return func() string {
        var parts []string
        
        // 添加提示
        if len(jb.hints) > 0 {
            parts = append(parts, fmt.Sprintf("/*+ %s */", strings.Join(jb.hints, ", ")))
        }
        
        // JOIN 类型
        parts = append(parts, jb.joinType)
        
        // 索引提示
        if jb.indexName != "" {
            parts = append(parts, fmt.Sprintf("FORCE INDEX (%s)", jb.indexName))
        }
        
        return strings.Join(parts, " ")
    }
}

// 使用示例
customJoin := NewJoin().
    Inner().
    UseHash().
    Parallel(4).
    UseIndex("idx_user_id").
    Build()

sql, args := xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, customJoin)
    }).
    Build().
    SqlOfSelect()

// 生成 SQL:
// SELECT * FROM orders
// /*+ USE_HASH, PARALLEL(4) */ INNER JOIN FORCE INDEX (idx_user_id) users
//   ON orders.user_id = users.id
```

---

## 🎯 最佳实践

### 1. 使用常量定义常用 JOIN

```go
// your_project/constants/joins.go
package constants

import "github.com/fndome/xb"

// 业务特定的 JOIN 常量
var (
    // 订单和用户的标准 JOIN（只连接有效用户）
    ORDER_USER_JOIN = func() string {
        return "INNER JOIN users ON orders.user_id = users.id AND users.deleted_at IS NULL"
    }
    
    // 订单和商品 JOIN（包含软删除商品）
    ORDER_GOODS_JOIN_WITH_DELETED = func() string {
        return "LEFT JOIN goods ON order_items.goods_id = goods.id"
    }
)

// 使用
sql, args := xb.Of(&Order{}).
    SourceBuilder.From(func(fb *xb.FromBuilder) {
        fb.From(&User{}, constants.ORDER_USER_JOIN)
    })
```

---

### 2. 命名规范

```go
// ✅ 推荐：描述性命名
LATERAL_JOIN()           ✅
ASOF_LEFT_JOIN()         ✅
GLOBAL_INNER_JOIN()      ✅
WITH_VALID_USER_JOIN()   ✅

// ❌ 避免：模糊命名
JOIN1()                  ❌
CUSTOM_JOIN()            ❌
MY_JOIN()                ❌
```

---

### 3. 文档注释

```go
// ✅ 好的注释
// ASOF_LEFT_JOIN ClickHouse ASOF LEFT JOIN
// 用于时序数据：找到时间戳最接近且不晚于的记录
//
// 示例:
//   fb.From(&Trade{}, clickhouse.ASOF_LEFT_JOIN).
//       On(&Order{}, "order_time", &Trade{}, "timestamp")
//
// 生成 SQL:
//   FROM orders ASOF LEFT JOIN trades
//   ON orders.order_time = trades.timestamp
func ASOF_LEFT_JOIN() string {
    return "ASOF LEFT JOIN"
}
```

---

## 🔧 测试建议

### 测试自定义 JOIN

```go
// your_project/sqlx_ext/joins_test.go
package sqlx_ext

import (
    "testing"
    "github.com/fndome/xb"
)

func TestCustomJoin_LATERAL(t *testing.T) {
    sql, args := xb.Of(&User{}).
        SourceBuilder.From(func(fb *xb.FromBuilder) {
            fb.From(&Order{}, LATERAL_JOIN)
        }).
        Build().
        SqlOfSelect()
    
    expected := "... LATERAL JOIN ..."
    if !strings.Contains(sql, "LATERAL JOIN") {
        t.Errorf("Expected LATERAL JOIN in SQL, got: %s", sql)
    }
    
    t.Logf("SQL: %s", sql)
}

func TestSmartJoin_LargeTable(t *testing.T) {
    joinFunc := SmartJoin(1000000, 5000000)
    joinStr := joinFunc()
    
    if !strings.Contains(joinStr, "USE_HASH") {
        t.Errorf("Large tables should use HASH JOIN, got: %s", joinStr)
    }
}
```

---

## 📊 扩展场景对比

| 场景 | 方式 | 复杂度 | 推荐度 |
|------|------|--------|--------|
| **简单 JOIN 变体** | 字符串常量 | ⭐ | ⭐⭐⭐⭐⭐ |
| **参数化 JOIN** | 闭包函数 | ⭐⭐ | ⭐⭐⭐⭐ |
| **智能 JOIN** | 动态逻辑 | ⭐⭐⭐ | ⭐⭐⭐ |
| **JOIN 构建器** | 完整 Builder | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## 🎊 总结

### 扩展 JOIN 的 3 种方式

#### 1. 简单字符串（最常用）⭐

```go
func LATERAL_JOIN() string {
    return "LATERAL JOIN"
}
```

**适用**：大多数标准 JOIN 变体

---

#### 2. 参数化闭包（中等）

```go
func HASH_JOIN(indexName string) xb.JOIN {
    return func() string {
        return fmt.Sprintf("/*+ HASH_JOIN(%s) */ INNER JOIN", indexName)
    }
}
```

**适用**：需要参数的 JOIN

---

#### 3. 构建器模式（复杂）

```go
NewJoin().
    Inner().
    UseHash().
    Parallel(4).
    Build()
```

**适用**：复杂的 JOIN 配置

---

### 核心原则

```
1. ✅ 不修改 xb 核心代码
2. ✅ 在自己的包内扩展
3. ✅ 遵循 xb 的函数式风格
4. ✅ 提供清晰的文档和示例
5. ✅ 编写完整的测试
```

---

## 🔗 相关资源

- **sqlxb JOIN 源码**: [joins.go](../joins.go)
- **sqlxb FROM 构建器**: [from_builder.go](../from_builder.go)
- **ClickHouse JOIN 文档**: https://clickhouse.com/docs/en/sql-reference/statements/select/join
- **PostgreSQL LATERAL**: https://www.postgresql.org/docs/current/queries-table-expressions.html#QUERIES-LATERAL

---

**通过扩展而非修改，让 xb 适应你的业务场景！** 🚀


