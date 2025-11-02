# Oracle Custom - Oracle 数据库专属支持

## 🎯 概述

`oracle_custom` 包提供了 Oracle 数据库的专属支持，主要解决 Oracle 分页语法与标准 SQL 的差异。

---

## 🚀 快速开始

### 安装

```bash
go get github.com/fndome/xb/oracle_custom
```

### 基础使用

```go
import (
    "github.com/fndome/xb"
    "github.com/fndome/xb/oracle_custom"
)

func main() {
    // Oracle 分页（ROWNUM）
    built := xb.Of("users").
        Custom(oracle_custom.New()).
        Eq("age", 18).
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(2).Rows(10)
        }).
        Build()
    
    countSQL, dataSQL, args, _ := built.SqlOfPage()
    
    // countSQL: SELECT COUNT(*) FROM users WHERE age = ?
    // dataSQL:  SELECT * FROM (
    //             SELECT a.*, ROWNUM rn FROM (
    //               SELECT * FROM users WHERE age = ?
    //             ) a WHERE ROWNUM <= 20
    //           ) WHERE rn > 10
}
```

---

## 📋 支持的分页语法

### 1. ROWNUM（Oracle 11g 及以下）- 默认

```go
custom := oracle_custom.New()  // 或 oracle_custom.WithRowNum()
```

**生成的 SQL**：

```sql
-- Data SQL
SELECT * FROM (
  SELECT a.*, ROWNUM rn FROM (
    SELECT * FROM users WHERE age = ?
  ) a WHERE ROWNUM <= 20
) WHERE rn > 10

-- Count SQL
SELECT COUNT(*) FROM users WHERE age = ?
```

---

### 2. FETCH FIRST（Oracle 12c+）

```go
custom := oracle_custom.WithFetchFirst()
```

**生成的 SQL**：

```sql
-- Data SQL
SELECT * FROM users WHERE age = ?
OFFSET 10 ROWS
FETCH NEXT 10 ROWS ONLY

-- Count SQL
SELECT COUNT(*) FROM users WHERE age = ?
```

---

## 🎨 预设模式

| 方法 | Oracle 版本 | 分页语法 | 说明 |
|------|------------|---------|------|
| `New()` | 11g+ | ROWNUM | 默认，兼容性最好 |
| `WithRowNum()` | 11g+ | ROWNUM | 显式声明 ROWNUM |
| `WithFetchFirst()` | 12c+ | FETCH FIRST | 性能更好，语法更简洁 |
| `Default()` | 11g+ | ROWNUM | 单例，等价于 New() |

---

## 💡 使用示例

### 示例 1: 基础分页

```go
import (
    "github.com/fndome/xb"
    "github.com/fndome/xb/oracle_custom"
)

func GetUsers(page, rows int) ([]User, error) {
    built := xb.Of("users").
        Custom(oracle_custom.New()).
        Gt("age", 18).
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(uint(page)).Rows(uint(rows))
        }).
        Build()
    
    countSQL, dataSQL, args, _ := built.SqlOfPage()
    
    // 执行查询...
}
```

---

### 示例 2: Oracle 12c+（FETCH FIRST）

```go
func GetUsersModern(page, rows int) ([]User, error) {
    built := xb.Of("users").
        Custom(oracle_custom.WithFetchFirst()).  // ⭐ Oracle 12c+
        Eq("status", "active").
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(uint(page)).Rows(uint(rows))
        }).
        Build()
    
    countSQL, dataSQL, args, _ := built.SqlOfPage()
    
    // 更简洁的 SQL（OFFSET/FETCH）
}
```

---

### 示例 3: 非分页查询

```go
// 非分页查询使用默认 SQL
built := xb.Of("users").
    Custom(oracle_custom.New()).
    Eq("age", 18).
    Build()

sql, args, _ := built.SqlOfSelect()
// SELECT * FROM users WHERE age = ?
// ⭐ 不含 ROWNUM（因为没有分页）
```

---

### 示例 4: INSERT/UPDATE（与标准 SQL 一致）

```go
// Oracle 的 INSERT/UPDATE 与标准 SQL 一致
built := xb.Of("users").
    Custom(oracle_custom.New()).
    Insert(func(ib *xb.InsertBuilder) {
        ib.Set("name", "张三").Set("age", 18)
    }).
    Build()

sql, args := built.SqlOfInsert()
// INSERT INTO users (name, age) VALUES (?, ?)
```

---

## 🎯 设计原则

### 1. 只处理差异部分

**Oracle Custom 只处理**：
- ✅ 分页语法（ROWNUM/FETCH FIRST）
- ✅ Count SQL 生成

**其他操作使用默认实现**：
- ✅ SELECT（非分页）
- ✅ INSERT
- ✅ UPDATE
- ✅ DELETE

---

### 2. 多版本兼容

| Oracle 版本 | 推荐方案 |
|------------|---------|
| Oracle 11g 及以下 | `oracle_custom.New()` |
| Oracle 12c+ | `oracle_custom.WithFetchFirst()` |

---

## 📊 性能对比

### ROWNUM vs FETCH FIRST

| 特性 | ROWNUM | FETCH FIRST |
|------|--------|-------------|
| 兼容性 | ✅ 11g+ | ⚠️ 12c+ |
| 性能 | ⚠️ 嵌套查询 | ✅ 优化器友好 |
| 语法 | ⚠️ 复杂 | ✅ 简洁 |
| 推荐 | 需要兼容旧版本 | Oracle 12c+ |

---

## 🔧 高级用法

### 自定义配置

```go
custom := &oracle_custom.OracleCustom{
    UseFetchFirst: true,   // 使用 FETCH FIRST
    Placeholder:   "?",    // 占位符
}

built := xb.Of("users").Custom(custom).Build()
```

---

## 📝 注意事项

### 1. PageCondition 必填

分页查询必须使用 `.Paged()`：

```go
// ✅ 正确
built := xb.Of("users").
    Custom(oracle_custom.New()).
    Paged(func(pb *xb.PageBuilder) {
        pb.Page(2).Rows(10)
    }).
    Build()

// ❌ 错误（不会使用 Oracle 分页语法）
built := xb.Of("users").
    Custom(oracle_custom.New()).
    Build()  // 没有 Paged()
```

---

### 2. 占位符自动转换

Oracle 驱动会自动转换占位符：
- xb 使用：`?`
- Oracle 执行：`:1, :2, :3`

无需手动处理！

---

## 🎯 与其他 Custom 对比

| Custom | 包名 | 用途 |
|--------|------|------|
| QdrantCustom | `xb` | 向量数据库（内置）|
| MySQLCustom | `xb` | MySQL UPSERT/IGNORE（内置）|
| **OracleCustom** | `oracle_custom` | Oracle 分页（独立包）⭐ |
| MilvusCustom | `milvus_custom` | Milvus 向量搜索（用户实现）|

---

## 📖 参考

- [xb Custom 接口设计哲学](../doc/CUSTOM_INTERFACE_PHILOSOPHY.md)
- [Custom 快速开始](../doc/CUSTOM_QUICKSTART.md)
- [xb 主包](https://github.com/fndome/xb)

---

**开始使用 Oracle Custom！** 🚀

