# xb 设计原则

## 🎯 核心原则

### **"Don't add concepts to solve problems"**

**每个具体的命名都是一个概念。** 概念越多，认知负担越重，学习成本越高。

---

## 📜 黄金法则

### 法则 1：概念守恒定律

```
框架价值 = 功能 / 概念数量

理想：功能增加，概念不增加
现实：功能不变，概念减少 ✅
```

**xb v1.2.0 验证**：
- 删除 8 个概念（预设函数、专用方法）
- 功能不减反增（智能格式检测）
- 价值提升 = ∞

---

### 法则 2：命名成本定律

```
每个公开 API 的成本 = 
    学习成本 + 
    记忆成本 + 
    决策成本 + 
    维护成本（永久）
```

**示例**：

| API | 学习 | 记忆 | 决策 | 维护 | 总成本 |
|-----|------|------|------|------|--------|
| `NewQdrantCustom()` | 低 | 低 | 无 | 低 | ⭐ 低 |
| `QdrantHighPrecision()` | 中 | 中 | 高 | 高 | ❌ 高 |

---

### 法则 3：API 不可逆定律

```
增加 API：1 小时
删除 API：永远不可能（breaking change）

结论：
- 每个 API 都是永久承诺
- 宁可少加，不能多加
- Less is more
```

---

## 🚫 禁止模式（Anti-Patterns）

### ❌ 反模式 1：预设配置函数

```go
// ❌ 禁止
func QdrantHighPrecision() *QdrantCustom { ... }
func QdrantHighSpeed() *QdrantCustom { ... }
func QdrantBalanced() *QdrantCustom { ... }

// ✅ 正确
func NewQdrantCustom() *QdrantCustom { ... }  // 只有这一个

// 用户配置
custom := NewQdrantCustom()
custom.DefaultHnswEf = 512  // 手动配置，清晰
```

**为什么禁止？**
1. 增加概念数量（3 个额外概念）
2. 用户需要决策（该用哪个？）
3. 配置不透明（HnswEf=512 是隐藏的）
4. 永久维护负担

---

### ❌ 反模式 2：专用方法

```go
// ❌ 禁止
func (x *BuilderX) InsertPoint(point interface{}) *BuilderX { ... }
func (x *BuilderX) InsertPoints(points []interface{}) *BuilderX { ... }

// ✅ 正确
func (x *BuilderX) Insert(f func(ib *InsertBuilder)) *BuilderX { ... }
// Custom 内部智能处理不同格式
```

**为什么禁止？**
1. 破坏 API 统一性
2. 用户困惑（Insert vs InsertPoint？）
3. 维护两套逻辑

---

### ❌ 反模式 3：便捷包装

```go
// ❌ 禁止（除非极其常用）
func (x *BuilderX) Delete() *BuilderX { ... }

// ✅ 正确
// 在方法内部处理
func (built *Built) JsonOfDelete() (string, error) {
    built.Delete = true  // 自动设置
    ...
}
```

**为什么禁止？**
1. 增加 API 表面积
2. 不够必要（调用 JsonOfDelete 已经明确意图）

---

## ✅ 推荐模式（Best Practices）

### ✅ 模式 1：单一构造函数 + 公开字段

```go
// ✅ 只有一个构造函数
func NewQdrantCustom() *QdrantCustom { ... }

// ✅ 公开字段供配置
type QdrantCustom struct {
    DefaultHnswEf         int     // 公开
    DefaultScoreThreshold float32 // 公开
    DefaultWithVector     bool    // 公开
}

// ✅ 用户配置
custom := NewQdrantCustom()
custom.DefaultHnswEf = 512
```

**优势**：
- 概念数：1
- 灵活性：无限
- 清晰度：100%

---

### ✅ 模式 2：在已有 API 上扩展

```go
// ✅ 已有的闭包 API
xb.Of(...).QdrantX(func(qx *QdrantBuilderX) {
    qx.HnswEf(512)  // 使用已有的，不新增
})

// ❌ 不要新增
xb.Of(...).QdrantHighPrecision()  // 新的 API
```

---

### ✅ 模式 3：便捷方法（必须极其常用）

```go
// ✅ 允许（极其常用）
func (built *Built) SqlOfUpsert() (string, []interface{})

// 判断标准：
// 1. 是否 90% 的 MySQL 用户都需要？（UPSERT：是）
// 2. 是否无法通过配置实现？（需要专门逻辑：是）
// 3. 是否会增加认知负担？（SqlOfUpsert 名字清晰：否）
```

---

## 🛡️ 守护机制

### 1. **代码注释守护**

在构造函数中加入：
```go
// ⚠️ 设计原则：只提供这一个构造函数！
// 参考：xb v1.1.0 的教训（预设函数 → v1.2.0 全部删除）
```

---

### 2. **文档守护**

在 `DESIGN_PRINCIPLES.md` 中明确：
- 禁止模式
- 历史教训
- 决策流程

---

### 3. **Code Review 守护**

PR Checklist：
- [ ] 是否增加了新的构造函数？（如果是 → 拒绝）
- [ ] 是否可以通过字段配置实现？（如果是 → 拒绝）
- [ ] 是否增加了概念数量？（如果是 → 慎重）

---

### 4. **AI Review 守护**

当 AI（包括我）提议增加 API 时，自动问：

```
🤖 AI: "要不要加 QdrantForRAG()？"

📋 Review Checklist:
1. 用户不用这个能实现吗？
   → 能（设置字段）
   
2. 这会增加概念数量吗？
   → 会（新增 1 个命名）
   
3. 那为什么要加？
   → ...
   
结论：❌ 拒绝
```

---

## 📊 历史教训

### xb v1.1.0 的错误

| 添加的概念 | 原因 | 现状 |
|----------|------|------|
| `QdrantHighPrecision()` | "方便用户" | v1.2.0 删除 |
| `QdrantHighSpeed()` | "方便用户" | v1.2.0 删除 |
| `QdrantBalanced()` | "方便用户" | v1.2.0 删除 |
| `MySQLWithUpsert()` | "方便用户" | v1.2.0 删除 |
| `MySQLWithIgnore()` | "方便用户" | v1.2.0 删除 |
| `InsertPoint()` | "专用 API" | v1.2.0 删除 |
| `InsertPoints()` | "批量操作" | v1.2.0 删除 |
| `Delete()` | "统一风格" | v1.2.0 删除 |

**总计删除：8 个概念**

---

## 🎯 决策流程

### 当想增加新 API 时

```
问题：用户需要 XXX 功能

Step 1: 能否通过现有 API 实现？
├─ 是 → 停止，不增加
└─ 否 → Step 2

Step 2: 能否通过字段配置实现？
├─ 是 → 停止，不增加
└─ 否 → Step 3

Step 3: 是否 90% 用户都需要？
├─ 否 → 停止，不增加
└─ 是 → Step 4

Step 4: 是否会增加概念数量？
├─ 是 → 重新思考设计
└─ 否 → 可以考虑（但要慎重）
```

---

## 💎 成功案例

### xb v1.2.0

**问题**：MySQL UPSERT 很常用

**❌ 错误方案**：
```go
MySQLWithUpsert()  // 新概念
```

**✅ 正确方案**：
```go
built.SqlOfUpsert()  // 便捷方法，名字清晰，不增加认知负担
```

---

### xb Qdrant

**问题**：Qdrant Insert 数据结构不同

**❌ 错误方案**：
```go
InsertPoint(point)  // 新 API，破坏统一性
```

**✅ 正确方案**：
```go
Insert(func(ib) { ... })  // 统一 API，Custom 内部智能处理
```

---

## 🌟 愿景

**让 xb 成为 Go 生态简洁性的标杆**

- ✅ 概念少：新手 5 分钟上手
- ✅ 功能强：覆盖 SQL + Vector DB
- ✅ API 稳定：一次学习，终身受用
- ✅ 影响 Go 生态：证明简洁可以很强大

---

## 📖 推荐阅读

- [Go Proverbs](https://go-proverbs.github.io/)
- [The Zen of Python](https://www.python.org/dev/peps/pep-0020/)
- [UNIX Philosophy](https://en.wikipedia.org/wiki/Unix_philosophy)

**核心思想相通**：
> "Do one thing and do it well"  
> "Worse is better"  
> "Less is exponentially more"

---

## 🔒 承诺

**xb 项目承诺**：

1. ✅ 每个数据库只有一个基础构造函数
2. ✅ 新 API 必须通过"4 步决策流程"
3. ✅ 定期审查：能否删除现有 API
4. ✅ 文档优先于代码

**守护 Go 生态的简洁性！**

---

**版本历史**：
- v1.1.0: 学到了教训（8 个概念删除）
- v1.2.0: 实践了原则（Less is more）
- 未来: 永远坚守（Don't add concepts）

---

**这个文档本身就是守护者。** 🛡️

**当你想增加 API 时，先读这个文档。** 📖

**如果 AI 提议增加概念，拿这个文档反驳我。** 🤖

---

**xb - 简洁的守护者！** ✨

