# xb 常见错误和解决方法

## 📋 概述

本文档列出使用 xb 时可能遇到的常见错误及其解决方法。

---

## 🚨 编译时错误

### 1. `No 'func (* Po) TableName() string' of interface Po`

**错误原因**：
```go
type User struct {
    ID   int64
    Name string
}

// ❌ 没有实现 TableName() 方法
builder := xb.Of(&User{})
```

**解决方法**：
```go
type User struct {
    ID   int64
    Name string
}

// ✅ 实现 TableName() 方法
func (*User) TableName() string {
    return "users"
}

builder := xb.Of(&User{})
```

---

### 2. `xb.Builder is nil`

**错误原因**：
```go
var builder *xb.BuilderX // nil
builder.Build() // panic
```

**解决方法**：
```go
// ✅ 正确初始化
builder := xb.Of(&User{})
builder.Build()
```

---

## ⚠️ 运行时错误

### 3. `page.rows must be greater than 0`

**错误原因**：
```go
builder.Paged(func(pb *xb.PageBuilder) {
    pb.Page(1).Rows(0) // ❌ rows 不能为 0
})
```

**解决方法**：
```go
builder.Paged(func(pb *xb.PageBuilder) {
    pb.Page(1).Rows(10) // ✅ rows 必须 > 0
})
```

---

### 4. `last > 0, Numeric sorts[0] required`

**错误原因**：
```go
builder.Paged(func(pb *xb.PageBuilder) {
    pb.Last(12345) // ❌ 使用 Last 但没有设置 Sort
})
```

**解决方法**：
```go
builder.Sort("id", xb.ASC). // ✅ 必须先设置数值字段排序
    Paged(func(pb *xb.PageBuilder) {
        pb.Last(12345)
    })
```

---

### 5. `call Cond(on *ON) after ON(onStr)`

**错误原因**：
```go
fb.Cond(func(on *xb.ON) {
    // ❌ 没有先调用 ON()
})
```

**解决方法**：
```go
fb.JOIN(xb.INNER).Of("orders").As("o").
   ON("o.user_id = u.id"). // ✅ 必须先调用 ON()
   Cond(func(on *xb.ON) {
       on.Gt("o.amount", 100)
   })
```

---

### 6. `USING.key can not blank`

**错误原因**：
```go
fb.Using("") // ❌ key 不能为空
```

**解决方法**：
```go
fb.Using("user_id") // ✅ 提供有效的字段名
```

---

### 7. `join, on can not nil`

**错误原因**：
```go
fb.JOIN(nil) // ❌ join 不能为 nil
```

**解决方法**：
```go
fb.JOIN(xb.INNER) // ✅ 使用预定义的 JOIN 类型
// 或
fb.JOIN(xb.LEFT)
fb.JOIN(xb.RIGHT)
```

---

## 🔍 逻辑错误

### 8. 复用 Builder 导致条件累积

**问题**：
```go
var baseBuilder = xb.Of(&User{})

func GetUser1() {
    sql, _, _ := baseBuilder.Eq("id", 1).Build().SqlOfSelect()
    // WHERE id = ?
}

func GetUser2() {
    sql, _, _ := baseBuilder.Eq("id", 2).Build().SqlOfSelect()
    // WHERE id = ? AND id = ? ❌ 条件累积了！
}
```

**解决方法**：
```go
// ✅ 每次创建新的 Builder
func GetUser1() {
    sql, _, _ := xb.Of(&User{}).Eq("id", 1).Build().SqlOfSelect()
}

func GetUser2() {
    sql, _, _ := xb.Of(&User{}).Eq("id", 2).Build().SqlOfSelect()
}
```

详见 [Builder Best Practices](./BUILDER_BEST_PRACTICES.md)

---

### 9. Like 重复添加通配符

**问题**：
```go
username := "john"
builder.Like("username", "%"+username+"%") // ❌ 会变成 %%john%%
```

**解决方法**：
```go
builder.Like("username", username) // ✅ 自动添加 %，变成 %john%
```

---

### 10. 不必要的 nil/0 检查

**问题**：
```go
// ❌ 不需要手动检查
if username != "" {
    builder.Like("username", username)
}
if age > 0 {
    builder.Gte("age", age)
}
```

**解决方法**：
```go
// ✅ 直接传递，自动过滤
builder.Like("username", username).
        Gte("age", age)
```

sqlxb 会自动忽略空字符串、nil 和 0 值。

---

## 🐛 类型错误

### 11. Vector 维度不匹配

**错误原因**：
```go
vec1 := xb.Vector{1, 2, 3}
vec2 := xb.Vector{1, 2, 3, 4, 5}
distance := vec1.Distance(vec2, xb.DistanceCosine) // panic: vectors must have same dimension
```

**解决方法**：
```go
// ✅ 确保向量维度相同
vec1 := xb.Vector{1, 2, 3}
vec2 := xb.Vector{4, 5, 6}
distance := vec1.Distance(vec2, xb.DistanceCosine)
```

---

### 12. Interceptor 错误

**错误原因**：
```go
type BadInterceptor struct{}

func (i *BadInterceptor) Name() string { return "bad" }

func (i *BadInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
    return fmt.Errorf("something wrong") // ❌ 返回错误
}

func (i *BadInterceptor) AfterBuild(built *xb.Built) error {
    return nil
}
```

**解决方法**：
```go
// ✅ 确保 Interceptor 不返回错误，或正确处理错误
func (i *GoodInterceptor) BeforeBuild(meta *interceptor.Metadata) error {
    meta.Set("trace_id", generateTraceID())
    return nil // ✅ 成功返回 nil
}
```

---

## 💡 性能问题

### 13. 未使用 Limit 导致大结果集

**问题**：
```go
builder := xb.Of(&User{}) // ❌ 可能返回数百万条记录
```

**解决方法**：
```go
// ✅ 使用 Limit
builder := xb.Of(&User{}).Limit(100)

// 或使用 Paged
builder := xb.Of(&User{}).
    Paged(func(pb *xb.PageBuilder) {
        pb.Page(1).Rows(10)
    })
```

---

### 14. 过度 over-fetch

**问题**：
```go
// ❌ 为了多样性，over-fetch 10 倍
builder.VectorSearch("embedding", vec, 1000).
    WithHashDiversity("category")
// 实际只需要 100 条
```

**解决方法**：
```go
// ✅ 合理的 over-fetch（2-3 倍）
builder.VectorSearch("embedding", vec, 300).
    WithHashDiversity("category")
// 在应用层过滤到 100 条
```

---

## 📚 相关文档

- [Builder Best Practices](./BUILDER_BEST_PRACTICES.md) - Builder 使用最佳实践
- [FAQ](./ai_application/FAQ.md) - 常见问题
- [All Filtering Mechanisms](./ALL_FILTERING_MECHANISMS.md) - 自动过滤机制

---

**最后更新**: 2025-02-27  
**版本**: v0.10.3

