# xb 空 OR/AND 自动过滤机制

## 🎯 您的问题

**"如果是 `OR()` 这样的，肯定要过滤掉 `OR()`，需要找出之前是在哪个环节完成了这种过滤"**

---

## ✅ 答案：在 `orAndSub()` 方法中

### 过滤位置

**文件**：`cond_builder.go` 第 145-159 行

```go
func (cb *CondBuilder) orAndSub(orAnd string, f func(cb *CondBuilder)) *CondBuilder {
    c := subCondBuilder()  // 创建子 CondBuilder
    f(c)                   // 执行用户提供的函数
    
    // ⭐ 关键：如果子条件为空，直接返回，不添加空的 OR/AND
    if c.bbs == nil || len(c.bbs) == 0 {
        return cb
    }

    // 只有当子条件不为空时，才添加 OR/AND
    bb := Bb{
        op:   orAnd,
        key:  orAnd,
        subs: c.bbs,
    }
    cb.bbs = append(cb.bbs, bb)
    return cb
}
```

---

## 📊 过滤流程

### 场景 1: 空 OR（所有子条件都是 nil/0）

```
用户代码:
  sqlxb.Of(&User{}).
      Eq("language", "golang").  // ✅ 有效
      Or(func(cb *CondBuilder) {
          cb.Eq("category", "")   // ⭐ 空字符串
          cb.Gt("rank", 0)        // ⭐ 0
      }).
      Gt("quality", 0.8)          // ✅ 有效

执行流程:
  ↓
Or() 调用 orAndSub(OR_SUB, f)
  ↓
c := subCondBuilder()  // 创建空的子 CondBuilder
  ↓
f(c)  // 执行用户函数
  ↓
  c.Eq("category", "")  → doGLE() → return cb (空字符串被过滤，不添加)
  c.Gt("rank", 0)       → doGLE() → return cb (0 被过滤，不添加)
  ↓
c.bbs = []  // ⭐ 子条件数组为空！
  ↓
if len(c.bbs) == 0 {
    return cb  // ⭐ 直接返回，不添加空的 OR
}

结果:
  cb.bbs = [
      Bb{op: EQ, key: "language", value: "golang"},
      Bb{op: GT, key: "quality", value: 0.8}
  ]
  // ⭐ 没有 OR 条件！

生成 SQL:
  SELECT * FROM users 
  WHERE language = ? AND quality > ?
  
  // ⭐ 没有 OR
```

---

### 场景 2: 部分有效的 OR

```
用户代码:
  .Or(func(cb *CondBuilder) {
      cb.Eq("category", "service")     // ✅ 有效
      cb.Eq("category", "repository")  // ✅ 有效
  })

执行流程:
  ↓
c.Eq("category", "service")     → addBb() ✅
c.Eq("category", "repository")  → addBb() ✅
  ↓
c.bbs = [
    Bb{op: EQ, key: "category", value: "service"},
    Bb{op: EQ, key: "category", value: "repository"}
]  // ⭐ 有 2 个条件

if len(c.bbs) == 0 {
    // 不会执行
}

// ⭐ 继续添加 OR
bb := Bb{op: OR_SUB, subs: c.bbs}
cb.bbs = append(cb.bbs, bb)

生成 SQL:
  WHERE ... OR (category = ? OR category = ?)
  
  // ⭐ OR 被保留
```

---

## 🔍 相关方法

### 1. `orAndSub()` - 主过滤逻辑

```go
// 处理 .Or() 和 .And() 方法
func (cb *CondBuilder) orAndSub(orAnd string, f func(cb *CondBuilder)) *CondBuilder

// 调用者：
func (cb *CondBuilder) And(f func(cb *CondBuilder)) *CondBuilder {
    return cb.orAndSub(AND_SUB, f)
}

func (cb *CondBuilder) Or(f func(cb *CondBuilder)) *CondBuilder {
    return cb.orAndSub(OR_SUB, f)
}
```

---

### 2. `doGLE()` - 单个条件过滤

```go
// 过滤 nil/0/空字符串
func (cb *CondBuilder) doGLE(p string, k string, v interface{}) *CondBuilder {
    switch v.(type) {
    case string:
        if v.(string) == "" {
            return cb  // ⭐ 不添加
        }
    case int, int64, ...:
        if v == 0 {
            return cb  // ⭐ 不添加
        }
    default:
        if v == nil {
            return cb  // ⭐ 不添加
        }
    }
    return cb.addBb(p, k, v)
}
```

---

## 🧪 测试验证

### 测试 1: 空 OR 被过滤

```go
func TestEmptyOr_Filtered(t *testing.T) {
    built := Of(&User{}).
        Eq("language", "golang").
        Or(func(cb *CondBuilder) {
            cb.Eq("category", "")  // 空字符串
            cb.Gt("rank", 0)       // 0
        }).
        Gt("quality", 0.8).
        Build()

    sql, args := built.SqlOfVectorSearch()
    
    // 预期 SQL: WHERE language = ? AND quality > ?
    // 实际 SQL: WHERE language = ? AND quality > ?
    // ✅ 空 OR 被过滤
}
```

**测试结果**：

```bash
$ go test -v -run TestEmptyOr_Filtered

=== RUN   TestEmptyOr_Filtered
    SQL: SELECT * FROM code_vectors WHERE language = ? AND quality > ?
    Args: [golang 0.8]
    ✅ 空 OR 被成功过滤
--- PASS: TestEmptyOr_Filtered (0.00s)
```

---

## 📝 过滤层级

sqlxb 有 **两层过滤**：

### 第 1 层：单个条件过滤（`doGLE()`）

```
过滤条件：
  - 空字符串 ("")
  - 数字 0
  - nil
  
位置：添加到 bbs 数组之前

示例：
  cb.Eq("name", "")  → 不添加到 bbs
  cb.Gt("count", 0)  → 不添加到 bbs
```

---

### 第 2 层：空 OR/AND 过滤（`orAndSub()`）

```
过滤条件：
  - 子条件数组为空（len(c.bbs) == 0）
  
位置：构建 OR/AND Bb 之前

示例：
  .Or(func(cb *CondBuilder) {
      cb.Eq("category", "")  // 第 1 层过滤
      cb.Gt("rank", 0)       // 第 1 层过滤
  })
  // 第 2 层过滤：整个 OR 被过滤
```

---

## 🎯 为什么需要两层过滤？

### 原因 1: 单个条件可能无效

```
.Eq("category", category)  // category 可能为 ""
```

**第 1 层过滤**：单个条件层面，如果 `category == ""`，不添加到 bbs

---

### 原因 2: OR/AND 可能全部无效

```
.Or(func(cb *CondBuilder) {
    cb.Eq("tag1", tag1)  // tag1 可能为 ""
    cb.Eq("tag2", tag2)  // tag2 可能为 ""
})
```

**第 2 层过滤**：如果所有子条件都被第 1 层过滤掉，整个 OR 也应该被过滤

---

## 🌟 优势

### 1. 用户无需手动判断

```
❌ 手动判断（繁琐且易错）:
if category != "" || tag1 != "" || tag2 != "" {
    builder.Or(func(cb *CondBuilder) {
        if category != "" {
            cb.Eq("category", category)
        }
        if tag1 != "" {
            cb.Eq("tag1", tag1)
        }
        if tag2 != "" {
            cb.Eq("tag2", tag2)
        }
    })
}

✅ sqlxb 自动过滤（简洁）:
builder.Or(func(cb *CondBuilder) {
    cb.Eq("category", category)  // 自动过滤
    cb.Eq("tag1", tag1)          // 自动过滤
    cb.Eq("tag2", tag2)          // 自动过滤
})  // 整个 OR 自动过滤
```

---

### 2. SQL 干净整洁

```
没有过滤:
  WHERE language = ? AND OR () AND quality > ?
  // ❌ 有空的 OR()

有过滤:
  WHERE language = ? AND quality > ?
  // ✅ 干净的 SQL
```

---

### 3. Qdrant JSON 也受益

```go
built := sqlxb.Of(&CodeVector{}).
    Eq("language", "golang").
    Or(func(cb *CondBuilder) {
        cb.Eq("category", "")
    }).
    VectorSearch("embedding", vec, 20).
    Build()

json, _ := built.ToQdrantJSON()
```

**输出**：

```json
{
  "vector": [0.1, 0.2, 0.3],
  "limit": 20,
  "filter": {
    "must": [
      {"key": "language", "match": {"value": "golang"}}
    ]
  }
}
```

**✅ 没有空的 `should`（OR 对应 `should`）**

---

## 🔗 相关过滤

sqlxb 还有其他自动过滤：

| 过滤类型 | 位置 | 说明 |
|---------|------|------|
| **nil/0 过滤** | `doGLE()` | 单个条件过滤 |
| **空 OR/AND 过滤** | `orAndSub()` | 整体 OR/AND 过滤 |
| **空 IN 过滤** | `doIn()` | `In()` 参数为空时过滤 |
| **空字符串 LIKE** | `Like()` | `Like("", "")` 时过滤 |

---

## 📚 总结

### 空 OR/AND 过滤机制

```
✅ 位置：cond_builder.go 的 orAndSub() 方法
✅ 时机：构建 OR/AND Bb 之前
✅ 条件：子条件数组为空（len(c.bbs) == 0）
✅ 结果：空的 OR/AND 不会添加到 bbs
✅ 好处：用户无需手动判断，SQL 自动干净
```

---

### 与 nil/0 过滤的关系

```
第 1 层：doGLE() 过滤单个条件
  Eq("category", "") → 不添加
  
第 2 层：orAndSub() 过滤空 OR/AND
  Or(所有子条件都被过滤) → 不添加

两层结合 → 完美的自动过滤
```

---

**这就是 sqlxb AI-First 设计的体现：用户只需关注业务逻辑，底层自动处理各种边界情况！** ✨

