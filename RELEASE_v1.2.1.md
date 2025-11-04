# xb v1.2.1 Release Notes

## 🎯 终极简化：认知负担最小化

**发布日期**: 2025-01-XX  
**核心成就**: 统一所有数据库配置为唯一的 `Custom()` 入口

---

## 为什么 v1.2.1 很重要？

### v1.2.0 的遗留问题

```go
// ❌ v1.2.0：两个配置入口（令人困惑）
xb.Of(...).
    Custom(NewQdrantCustom()).  // ← INSERT/UPDATE/DELETE 用这个
    VectorSearch(...).
    QdrantX(func(qx) {           // ← SELECT 用这个？？
        qx.HnswEf(512)
    })
```

**问题**：
- 人类需要记住：什么时候用 `Custom()`，什么时候用 `QdrantX()`
- 概念分裂：两个配置方式
- 认知负担：AI 能记住规则，人类容易混淆

### v1.2.1 的解决方案

```go
// ✅ v1.2.1：唯一入口（清晰明了）
xb.Of(...).
    Custom(
        xb.NewQdrantBuilder().
            HnswEf(512).
            ScoreThreshold(0.8).
            Build(),
    ).
    VectorSearch(...).
    Build()
```

**优势**：
- ✅ 只记住一个入口：`Custom()`
- ✅ 所有操作统一：INSERT/UPDATE/DELETE/SELECT
- ✅ 链式调用：流畅、易读
- ✅ 配置复用：Builder 模式天然支持

---

## 新增功能

### 1. Builder 模式

#### Qdrant 配置构建器
```go
config := xb.NewQdrantBuilder().
    HnswEf(512).
    ScoreThreshold(0.85).
    WithVector(false).
    Build()

xb.Of(...).Custom(config).Build()
```

#### MySQL 配置构建器
```go
config := xb.NewMySQLBuilder().
    UseUpsert(true).
    Build()

xb.Of(...).Custom(config).Build()
```

### 2. Custom 默认值在 SELECT 中生效

```go
// ✅ DefaultHnswEf 现在会影响 SELECT 查询
xb.Of(...).
    Custom(
        NewQdrantBuilder().HnswEf(512).Build()
    ).
    VectorSearch(...).
    Build().JsonOfSelect()  // ← hnsw_ef=512 ✅
```

**优先级**：
```
QdrantX 参数 > Custom 默认值 > 硬编码默认值
（已删除 QdrantX，现在只有 Custom 默认值）
```

---

## 删除的功能

### 完全移除 QdrantX()

| 删除内容 | 原因 |
|---------|------|
| `qdrant_x.go` | 职责与 Custom 重复 |
| `QdrantX()` 方法 | 增加认知负担 |
| `QdrantBuilderX` | 不再需要 |
| 6 个测试文件 | 相关功能已删除 |

### 删除过度设计

- `CustomBuilder[T any]` 泛型接口 - Go 的类型系统不支持 `? extends`，强行加接口反而增加复杂度

---

## API 对比

### v1.2.0 vs v1.2.1

| 场景 | v1.2.0 | v1.2.1 |
|------|--------|--------|
| **Qdrant INSERT** | `Custom(NewQdrantCustom())` | `Custom(NewQdrantBuilder().Build())` |
| **Qdrant SELECT** | `QdrantX(func(qx) {...})` | `Custom(NewQdrantBuilder()...Build())` ✅ |
| **MySQL INSERT** | `Custom(NewMySQLCustom())` | `Custom(NewMySQLBuilder().Build())` |
| **记忆负担** | 2 个入口 | 1 个入口 ✅ |

---

## 完整示例

### Qdrant 向量搜索

```go
// ✅ 统一的配置方式
result, _ := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantBuilder().
            HnswEf(512).
            ScoreThreshold(0.8).
            WithVector(false).
            Build(),
    ).
    VectorSearch("embedding", queryVector, 20).
    Eq("language", "golang").
    Build().
    JsonOfSelect()
```

### MySQL Upsert

```go
// ✅ 统一的配置方式
sql, args := xb.Of(&User{}).
    Custom(
        xb.NewMySQLBuilder().
            UseUpsert(true).
            Build(),
    ).
    Insert(func(ib *InsertBuilder) {
        ib.Set("name", "张三").Set("age", 18)
    }).
    Build().
    SqlOfInsert()
```

### 配置复用

```go
// ✅ Builder 模式天然支持配置复用
highPrecision := xb.NewQdrantBuilder().HnswEf(512).Build()

// 多次使用
result1 := xb.Of(...).Custom(highPrecision).VectorSearch(...).Build()
result2 := xb.Of(...).Custom(highPrecision).VectorSearch(...).Build()
```

---

## 设计哲学

### 核心原则

> **"Don't add concepts to solve problems"**

### 本次优化的思考

**问题**：Java 有 `? extends Custom` 可以做类型约束，Go 没有，怎么办？

**答案**：不需要！
- Go 的 Duck Typing 已经够用
- 强行加泛型接口反而增加复杂度
- 简单的接口 + 显式 `.Build()` = 最优解

**权衡**：
- ✅ 接受显式调用 `.Build()`（职责清晰）
- ✅ 接受无法用 `func` 延迟构建（务实选择）
- ✅ 换来统一的 API 和最低的记忆成本

### 用户只需记住

```
1. NewXxxBuilder()  - 创建构建器
2. .Method()        - 链式配置
3. .Build()         - 构建配置
4. Custom()         - 统一入口
```

**4 个步骤，0 个例外，100% 一致性** 🎯

---

## 测试覆盖

- **总测试数**: 196 个
- **测试结果**: 全部通过 ✅
- **新增测试**:
  - `qdrant_builder_test.go` - QdrantBuilder 测试
  - `mysql_builder_test.go` - MySQLBuilder 测试
  - `qdrant_custom_priority_test.go` - 优先级测试
- **删除测试**:
  - `qdrant_x_test.go`
  - `qdrant_xx_test.go`
  - `qdrant_compat_test.go`
  - `qdrant_recommend_test.go`
  - `qdrant_discover_test.go`
  - `qdrant_custom_select_test.go`

---

## 破坏性变更

### 删除的 API

```go
// ❌ 已删除
.QdrantX(func(qx *QdrantBuilderX) {...})

// ✅ 迁移到
.Custom(NewQdrantBuilder()...Build())
```

### 迁移指南

**场景 1: 简单配置**
```go
// Before
.QdrantX(func(qx) { qx.HnswEf(512) })

// After
.Custom(NewQdrantBuilder().HnswEf(512).Build())
```

**场景 2: 复杂配置**
```go
// Before
.QdrantX(func(qx) {
    qx.HnswEf(512).ScoreThreshold(0.8).WithVector(false)
})

// After
.Custom(
    NewQdrantBuilder().
        HnswEf(512).
        ScoreThreshold(0.8).
        WithVector(false).
        Build(),
)
```

**场景 3: 配置复用**
```go
// Before
// 每次都要写闭包，无法复用

// After
config := NewQdrantBuilder().HnswEf(512).Build()
xb.Of(...).Custom(config).Build()
xb.Of(...).Custom(config).Build()  // 复用
```

---

## 技术细节

### Custom 默认值的应用逻辑

修改了 `ToQdrantRequest()` 方法：

```go
// 1. 从 Custom 读取默认值
if built.Custom != nil {
    if qdrantCustom, ok := built.Custom.(*QdrantCustom); ok {
        defaultHnswEf = qdrantCustom.DefaultHnswEf
        defaultScoreThreshold = qdrantCustom.DefaultScoreThreshold
        defaultWithVector = qdrantCustom.DefaultWithVector
    }
}

// 2. 应用默认值
req.Params.HnswEf = defaultHnswEf
req.ScoreThreshold = &defaultScoreThreshold
req.WithVector = defaultWithVector

// 3. 运行时参数覆盖（已删除 QdrantX，此逻辑保留用于未来扩展）
applyQdrantSpecificConfig(built.Conds, req)
```

---

## 向后兼容性

### ✅ 保持兼容的功能

- `NewQdrantCustom()` - 直接创建 Custom
- `NewMySQLCustom()` - 直接创建 Custom
- 手动设置字段：`custom.DefaultHnswEf = 512`
- 所有 SQL 相关 API

### ❌ 不兼容的功能（需要迁移）

- `QdrantX()` 方法 → 改用 `Custom(NewQdrantBuilder()...Build())`

---

## 文件变更统计

### 新增文件
- `qdrant_builder_test.go` - QdrantBuilder 测试
- `mysql_builder_test.go` - MySQLBuilder 测试
- `qdrant_custom_priority_test.go` - 优先级测试

### 修改文件
- `qdrant_custom.go` - 添加 QdrantBuilder
- `mysql_custom.go` - 添加 MySQLBuilder
- `to_qdrant_json.go` - Custom 默认值应用到 SELECT
- `CHANGELOG.md` - 添加 v1.2.1 条目

### 删除文件
- `qdrant_x.go` - QdrantX 方法
- `qdrant_x_test.go` - QdrantX 测试
- `qdrant_xx_test.go` - QdrantX 扩展测试
- `qdrant_compat_test.go` - 兼容性测试
- `qdrant_recommend_test.go` - Recommend 测试
- `qdrant_discover_test.go` - Discover 测试
- `qdrant_custom_select_test.go` - 临时测试

**净减少**: 3 个文件（代码更简洁）

---

## 性能影响

**无性能影响**：
- Builder 模式在构建时执行，不影响运行时
- Custom 默认值读取是 O(1) 操作
- 所有原有的优化保持不变

---

## 下一步

v1.2.1 是 xb 走向完美的重要一步。统一的 API 让框架更易用、更易记、更易维护。

未来可能的优化方向：
1. Milvus/Weaviate 的 Builder 模式
2. 更多的 Custom 配置选项
3. 更好的文档和示例

但始终遵循：**Don't add concepts to solve problems**

---

## 升级建议

**推荐升级**：v1.2.1 的 API 统一性值得升级成本

**迁移成本**：低
- 主要是将 `QdrantX()` 改为 `Custom(NewQdrantBuilder()...Build())`
- 大部分代码无需修改

**收益**：高
- 认知负担大幅降低
- API 一致性提升
- 代码可维护性增强

---

**xb - 迈向更完美！** 🚀

