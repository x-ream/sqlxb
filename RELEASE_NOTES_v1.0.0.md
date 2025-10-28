# xb v1.0.0 Release Notes 🎉

**发布日期**: 2025-10-28  
**里程碑**: 首个生产就绪版本

---

## 🎯 版本定位

`xb v1.0.0` 是一个**生产就绪**的 Go SQL Builder + Vector DB 客户端库，经过：
- ✅ **4 个生产服务**集成测试验证
- ✅ **95%+ 单元测试覆盖率**
- ✅ **Fuzz 测试**加固
- ✅ **完整文档**和最佳实践

---

## 📦 核心特性

### 1. SQL Builder（关系数据库）

```go
// PostgreSQL / MySQL 查询
sql, args, _ := xb.Of(&User{}).
    Eq("status", 1).
    Like("username", "alice").  // 自动添加 %
    Gte("created_at", time.Now().AddDate(0, -1, 0)).
    Sort("id", xb.DESC).
    Limit(10).
    Build().
    SqlOfSelect()
```

**支持的操作**:
- ✅ CRUD（Insert / Update / Delete / Select）
- ✅ 条件查询（Eq / In / Like / Gte / Lte / IsNull ...）
- ✅ 聚合（GroupBy / Having / Agg）
- ✅ JOIN（InnerJoin / LeftJoin）
- ✅ 分页（Paged / Limit / Offset）
- ✅ 排序（Sort with ASC/DESC）
- ✅ 自动过滤（nil / 0 / 空字符串 / 空切片 / zero time.Time）

---

### 2. Vector DB（Qdrant / pgvector）

```go
// Qdrant 向量搜索
json := xb.QdrantX("products", embedding).
    WithFilter(func(b *xb.CondBuilder) {
        b.Eq("category", "electronics").
          Gte("price", 100)
    }).
    WithScoreThreshold(0.7).
    Limit(20).
    Build().
    ToQdrantJSON()
```

**支持的操作**:
- ✅ VectorSearch（向量搜索）
- ✅ Recommend（推荐）
- ✅ Discover（发现）
- ✅ Scroll（滚动查询）
- ✅ Hybrid Search（混合搜索：向量 + 过滤）
- ✅ Score Threshold（相似度阈值）
- ✅ HNSW 配置（ANN 算法参数）

---

### 3. 混合查询（AI 应用场景）

```go
// RAG: 向量搜索 + 关系数据库过滤
type Document struct {
    ID        int64     `db:"id"`
    Content   string    `db:"content"`
    DocID     *int64    `db:"doc_id"`      // 文档 ID（可选）
    Category  string    `db:"category"`
    CreatedAt time.Time `db:"created_at"`
    Embedding []float32 `db:"embedding"`   // pgvector
}

// pgvector 查询
sql, args := xb.Of(&Document{}).
    Eq("category", "tech").              // 关系数据库过滤
    VectorSearch("embedding", queryEmb, 10, xb.L2).  // 向量搜索
    Build().
    SqlOfVectorSearch()
```

---

## 🆕 v1.0.0 新特性

### 从 v0.11.1 以来

1. **集成测试验证** ✅
   - 在 4 个生产服务中验证兼容性
   - 100% API 向后兼容

2. **完整测试覆盖** ✅
   - 单元测试覆盖率：**95.x%**
   - Fuzz 测试：4 个关键函数
   - 集成测试：server-g 项目

3. **文档完善** ✅
   - 最佳实践指南
   - 常见错误排查
   - 4 个应用示例（pgvector / Qdrant / RAG / PageIndex）
   - AI 应用集成指南（LangChain / LlamaIndex）

---

## 📚 从 v0.7.4 升级

### 升级步骤

1. **更新依赖**:
```bash
# 旧版本
github.com/x-ream/sqlxb v0.7.4

# 新版本
github.com/fndome/xb v1.0.0
```

2. **更新 import**:
```go
// 旧
import "github.com/x-ream/sqlxb"

// 新
import "github.com/fndome/xb"
```

3. **更新包名**:
```go
// 旧
sqlxb.Of(&User{})

// 新
xb.Of(&User{})
```

### ✅ 100% 兼容

所有 API 保持向后兼容，无需修改业务逻辑！

详见：[MIGRATION.md](./MIGRATION.md)

---

## 🔧 技术亮点

### 1. 自动过滤机制

```go
// 无需手动检查 nil/0/空值
user := &User{
    Status: 0,           // ❌ 被忽略（零值）
    Email:  "",          // ❌ 被忽略（空字符串）
    Name:   "Alice",     // ✅ 生效
}

xb.Of(user).
    Eq("status", user.Status).   // ❌ 自动忽略
    Eq("name", user.Name).        // ✅ 生成 WHERE name = ?
    Build()
```

### 2. Builder 模式

```go
// 链式调用，类型安全
builder := xb.Of(&Order{}).
    Eq("user_id", userId).
    In("status", []int{1, 2, 3}).
    Gte("created_at", startTime).
    Sort("id", xb.DESC).
    Limit(20)

// 按需生成 SQL
sql1, args1 := builder.Build().SqlOfSelect()
sql2, args2 := builder.Build().SqlOfDelete()
```

### 3. 性能优化

- ✅ `strings.Builder` 预分配（减少内存分配）
- ✅ 延迟求值（Build() 前不生成 SQL）
- ✅ 零反射（使用 `db` tag）

---

## 📖 文档资源

### 快速开始
- [README.md](./README.md) - 项目概览和快速入门
- [MIGRATION.md](./MIGRATION.md) - 从 sqlxb 迁移指南

### 最佳实践
- [BUILDER_BEST_PRACTICES.md](./doc/BUILDER_BEST_PRACTICES.md) - Builder 使用指南
- [COMMON_ERRORS.md](./doc/COMMON_ERRORS.md) - 常见错误排查

### 应用示例
- [examples/pgvector-app](./examples/pgvector-app) - PostgreSQL + pgvector
- [examples/qdrant-app](./examples/qdrant-app) - Qdrant 集成
- [examples/rag-app](./examples/rag-app) - RAG 应用
- [examples/pageindex-app](./examples/pageindex-app) - PageIndex 结构化检索

### AI 应用集成
- [doc/ai_application/LANGCHAIN_INTEGRATION.md](./doc/ai_application/LANGCHAIN_INTEGRATION.md)
- [doc/ai_application/LLAMAINDEX_INTEGRATION.md](./doc/ai_application/LLAMAINDEX_INTEGRATION.md)
- [doc/ai_application/RAG_BEST_PRACTICES.md](./doc/ai_application/RAG_BEST_PRACTICES.md)

---

## 🛣️ Roadmap

### v1.1.0（计划中）
- 性能基准测试（Benchmark）
- 更多集成测试
- 连接池最佳实践

### v1.2.0（计划中）
- 批量插入优化
- 事务辅助函数
- SQL 日志拦截器示例

详见：[ROADMAP_v1.0.0.md](./doc/ROADMAP_v1.0.0.md)

---

## 📊 项目统计

- **代码行数**: ~5000+ lines
- **测试覆盖率**: 95%+
- **文档页数**: 20+ 文档
- **示例应用**: 4 个完整示例
- **生产验证**: 4 个服务集成测试

---

## 🙏 致谢

感谢所有贡献者和早期采用者！

特别感谢：
- **x-ream** 组织（原 sqlxb 项目）
- **server-g** 项目团队（集成测试）
- **fndome** 社区

---

## 📝 许可证

Apache License 2.0

---

## 🚀 快速安装

```bash
go get github.com/fndome/xb@v1.0.0
```

```go
import "github.com/fndome/xb"

// 开始使用
sql, args, _ := xb.Of(&YourModel{}).
    Eq("field", value).
    Build().
    SqlOfSelect()
```

---

**Happy Coding with xb! 🎉**

