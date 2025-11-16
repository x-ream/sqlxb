# xb (可扩展查询构建器)
[![OSCS Status](https://www.oscs1024.com/platform/badge/fndome/xb.svg?size=small)](https://www.oscs1024.com/project/fndome/xb?ref=badge_small)
![workflow build](https://github.com/fndome/xb/actions/workflows/go.yml/badge.svg)
[![GitHub tag](https://img.shields.io/github/tag/fndome/xb.svg?style=flat)](https://github.com/fndome/xb/tags)
[![Go Report Card](https://goreportcard.com/badge/github.com/fndome/xb)](https://goreportcard.com/report/github.com/fndome/xb)

> 文档语言：**中文** | [English](../en/README.md)

`xb` 是面向 AI 时代的 SQL / JSON 链式构建器，统一服务关系型数据库与向量数据库。一套 API 可以：

- 为 `database/sql`、`sqlx`、`gorm` 等生成 SQL
- 为 Qdrant（Recommend / Discover / Scroll）及其他向量存储生成 JSON
- 组合 SQL 与向量检索，构建混合查询流水线

所有能力都收敛到 `Custom()` + `Build()`，让可扩展性与记忆成本同时最小化。

> **提示**
> - 持久化 Struct 可直接镜像数据库字段（如自增主键），不必全部设置为指针。
> - 请求 / 过滤 DTO 建议将非主键的数值、布尔字段声明为指针，以区分“未填写”与“填写 0/false”，也便于 xb 自动跳过无效条件。
> - 若需要手动控制，可使用 `X("...")` 注入原始 SQL（不会被自动过滤），或在 JOIN 场景中选择特定的 JOIN helper（例如 `JOIN(NON_JOIN)` / 自定义 JOIN）来关闭自动裁剪；对 `BuilderX` 可调用 `WithoutOptimization()` 完全关闭 JOIN/CTE 优化。
> - 链式调用里需要临时控制流时，可使用 `Any(func(*BuilderX))` 包裹循环或辅助函数；布尔条件需要复用时，可使用 `Bool(func() bool, func(*CondBuilder))`。

---

## 亮点

- **统一向量入口**：`JsonOfSelect()` 覆盖 Qdrant 的搜索 / 推荐 / 发现 / Scroll，彻底移除 `ToQdrant*JSON()`。
- **组合式 SQL**：`With/WithRecursive`、`UNION(kind, fn)` 用 Go 代码描述复杂 CTE 与报表。
- **智能条件 DSL**：自动过滤 `nil/0/""`，提供 `InRequired` / `X()` / `CondBuilderX.Sub()` 等守卫与逃生舱。
- **自适应 JOIN 规划**：`FromX` + `JOIN(kind)` 可自动跳过空 ON 条件或无意义的联表，让 SQL 更精简。
- **可观测设计**：`Meta(func)` 搭配全局拦截器，TraceID 与用户上下文贯穿各阶段。
- **AI 辅助维护**：每次发版都由 AI + 人类共同编写代码、测试与文档。

📦 **最新版本**：[v1.3.0](../../RELEASE_v1.3.0.md) – 统一 JsonOfSelect + Qdrant 高级 API。

---

## 快速开始

### 构建 SQL
```go
type Cat struct {
    ID    uint64   `db:"id"`
    Name  string   `db:"name"`
    Age   *uint    `db:"age"`
    Price *float64 `db:"price"`
}

built := xb.Of(&Cat{}).
    Eq("status", 1).
    Gte("age", 3).
    Build()

sql, args, _ := built.SqlOfSelect()
// SELECT * FROM t_cat WHERE status = ? AND age >= ?
```

### 构建 Qdrant Recommend 请求
```go
json, _ := xb.Of(&CodeVector{}).
    Custom(
        xb.NewQdrantCustom().
            Recommend(func(rb *xb.RecommendBuilder) {
                rb.Positive(123, 456).Negative(789).Limit(20)
            }),
    ).
    Eq("language", "golang").
    VectorSearch("embedding", queryVector, 10).
    Build().
    JsonOfSelect()
```

### JOIN + 子查询
```go
builder := xb.X().
    Select("p.id", "p.weight").
    FromX(func(fb *xb.FromBuilder) {
        fb.Sub(sub).As("p").
            JOIN(xb.INNER).Of("t_dog").As("d").On("d.pet_id = p.id").
            JOIN(xb.LEFT).Of("t_cat").As("c").On("c.pet_id = p.id")
    }).
    Ne("p.type", "PIG")

sql, args, _ := builder.Build().SqlOfSelect()
```

更多示例请参考 `doc/cn/QUICKSTART.md` 与 `doc/cn/VECTOR_GUIDE.md`。

---

## 文档索引

| 主题 | 中文 | 英文 |
|------|------|------|
| 入口 & 索引 | `README.md`（本文） | [doc/en/README.md](../en/README.md) |
| 快速上手 | [QUICKSTART.md](./QUICKSTART.md) | [doc/en/QUICKSTART.md](../en/QUICKSTART.md) |
| 向量指南 | [VECTOR_GUIDE.md](./VECTOR_GUIDE.md) | [doc/en/VECTOR_GUIDE.md](../en/VECTOR_GUIDE.md) |
| Qdrant 指南 | [QDRANT_GUIDE.md](./QDRANT_GUIDE.md) | [doc/en/QDRANT_GUIDE.md](../en/QDRANT_GUIDE.md) |
| Custom 接口 | [CUSTOM_INTERFACE.md](./CUSTOM_INTERFACE.md) | [doc/en/CUSTOM_INTERFACE.md](../en/CUSTOM_INTERFACE.md) |
| 自动过滤 / 守卫 | [FILTERING.md](./FILTERING.md) | [doc/en/ALL_FILTERING_MECHANISMS.md](../en/ALL_FILTERING_MECHANISMS.md) |
| JOIN 优化 | [CUSTOM_JOINS_GUIDE.md](./CUSTOM_JOINS_GUIDE.md) | [doc/en/CUSTOM_JOINS_GUIDE.md](../en/CUSTOM_JOINS_GUIDE.md) |
| AI 应用指南 | [AI_APPLICATION.md](./AI_APPLICATION.md) | [doc/en/AI_APPLICATION.md](../en/AI_APPLICATION.md) |

### 核心文档

| 主题 | 中文 | 英文 |
|------|------|------|
| 所有过滤机制 | [ALL_FILTERING_MECHANISMS.md](./ALL_FILTERING_MECHANISMS.md) | [doc/en/ALL_FILTERING_MECHANISMS.md](../en/ALL_FILTERING_MECHANISMS.md) |
| 构建器最佳实践 | [BUILDER_BEST_PRACTICES.md](./BUILDER_BEST_PRACTICES.md) | [doc/en/BUILDER_BEST_PRACTICES.md](../en/BUILDER_BEST_PRACTICES.md) |
| 常见错误 | [COMMON_ERRORS.md](./COMMON_ERRORS.md) | [doc/en/COMMON_ERRORS.md](../en/COMMON_ERRORS.md) |
| 测试策略 | [TESTING_STRATEGY.md](./TESTING_STRATEGY.md) | [doc/en/TESTING_STRATEGY.md](../en/TESTING_STRATEGY.md) |
| 命名约定 | [NAMING_CONVENTION.md](./NAMING_CONVENTION.md) | [doc/en/NAMING_CONVENTION.md](../en/NAMING_CONVENTION.md) |

### Qdrant 相关

| 主题 | 中文 | 英文 |
|------|------|------|
| Qdrant 高级 API | [QDRANT_ADVANCED_API.md](./QDRANT_ADVANCED_API.md) | [doc/en/QDRANT_ADVANCED_API.md](../en/QDRANT_ADVANCED_API.md) |
| Qdrant 优化总结 | [QDRANT_OPTIMIZATION_SUMMARY.md](./QDRANT_OPTIMIZATION_SUMMARY.md) | [doc/en/QDRANT_OPTIMIZATION_SUMMARY.md](../en/QDRANT_OPTIMIZATION_SUMMARY.md) |
| Qdrant X 使用 | [QDRANT_X_USAGE.md](./QDRANT_X_USAGE.md) | [doc/en/QDRANT_X_USAGE.md](../en/QDRANT_X_USAGE.md) |
| 为什么选择 Qdrant | [WHY_QDRANT.md](./WHY_QDRANT.md) | [doc/en/WHY_QDRANT.md](../en/WHY_QDRANT.md) |
| 向量多样性与 Qdrant | [VECTOR_DIVERSITY_QDRANT.md](./VECTOR_DIVERSITY_QDRANT.md) | [doc/en/VECTOR_DIVERSITY_QDRANT.md](../en/VECTOR_DIVERSITY_QDRANT.md) |

### Custom 接口相关

| 主题 | 中文 | 英文 |
|------|------|------|
| Custom 接口哲学 | [CUSTOM_INTERFACE_PHILOSOPHY.md](./CUSTOM_INTERFACE_PHILOSOPHY.md) | [doc/en/CUSTOM_INTERFACE_PHILOSOPHY.md](../en/CUSTOM_INTERFACE_PHILOSOPHY.md) |
| Custom 快速开始 | [CUSTOM_QUICKSTART.md](./CUSTOM_QUICKSTART.md) | [doc/en/CUSTOM_QUICKSTART.md](../en/CUSTOM_QUICKSTART.md) |
| Custom 向量数据库指南 | [CUSTOM_VECTOR_DB_GUIDE.md](./CUSTOM_VECTOR_DB_GUIDE.md) | [doc/en/CUSTOM_VECTOR_DB_GUIDE.md](../en/CUSTOM_VECTOR_DB_GUIDE.md) |
| Custom 架构验证 | [CUSTOM_ARCHITECTURE_VALIDATION.md](./CUSTOM_ARCHITECTURE_VALIDATION.md) | [doc/en/CUSTOM_ARCHITECTURE_VALIDATION.md](../en/CUSTOM_ARCHITECTURE_VALIDATION.md) |
| Dialect vs Custom 设计 | [DIALECT_CUSTOM_DESIGN.md](./DIALECT_CUSTOM_DESIGN.md) | [doc/en/DIALECT_CUSTOM_DESIGN.md](../en/DIALECT_CUSTOM_DESIGN.md) |

### 向量数据库相关

| 主题 | 中文 | 英文 |
|------|------|------|
| 向量快速开始 | [VECTOR_QUICKSTART.md](./VECTOR_QUICKSTART.md) | [doc/en/VECTOR_QUICKSTART.md](../en/VECTOR_QUICKSTART.md) |
| 向量数据库扩展指南 | [VECTOR_DB_EXTENSION_GUIDE.md](./VECTOR_DB_EXTENSION_GUIDE.md) | [doc/en/VECTOR_DB_EXTENSION_GUIDE.md](../en/VECTOR_DB_EXTENSION_GUIDE.md) |
| 向量数据库接口设计 | [VECTOR_DB_INTERFACE_DESIGN.md](./VECTOR_DB_INTERFACE_DESIGN.md) | [doc/en/VECTOR_DB_INTERFACE_DESIGN.md](../en/VECTOR_DB_INTERFACE_DESIGN.md) |

### AI 应用相关

| 主题 | 中文 | 英文 |
|------|------|------|
| AI 应用索引 | [ai_application/README.md](./ai_application/README.md) | [doc/en/ai_application/README.md](../en/ai_application/README.md) |
| RAG 最佳实践 | [ai_application/RAG_BEST_PRACTICES.md](./ai_application/RAG_BEST_PRACTICES.md) | [doc/en/ai_application/RAG_BEST_PRACTICES.md](../en/ai_application/RAG_BEST_PRACTICES.md) |
| 混合搜索 | [ai_application/HYBRID_SEARCH.md](./ai_application/HYBRID_SEARCH.md) | [doc/en/ai_application/HYBRID_SEARCH.md](../en/ai_application/HYBRID_SEARCH.md) |
| LangChain 集成 | [ai_application/LANGCHAIN_INTEGRATION.md](./ai_application/LANGCHAIN_INTEGRATION.md) | [doc/en/ai_application/LANGCHAIN_INTEGRATION.md](../en/ai_application/LANGCHAIN_INTEGRATION.md) |
| LlamaIndex 集成 | [ai_application/LLAMAINDEX_INTEGRATION.md](./ai_application/LLAMAINDEX_INTEGRATION.md) | [doc/en/ai_application/LLAMAINDEX_INTEGRATION.md](../en/ai_application/LLAMAINDEX_INTEGRATION.md) |

### 贡献相关

| 主题 | 中文 | 英文 |
|------|------|------|
| 贡献指南 | [CONTRIBUTING.md](./CONTRIBUTING.md) | [doc/en/CONTRIBUTING.md](../en/CONTRIBUTING.md) |
| 贡献者 | [CONTRIBUTORS.md](./CONTRIBUTORS.md) | [doc/en/CONTRIBUTORS.md](../en/CONTRIBUTORS.md) |

> 旧版文档仍保留在 `xb/doc/`，待逐步迁移后统一指向 `doc/en` / `doc/cn`。

---

## 发布 & 迁移

- [Release Notes v1.3.0](../../RELEASE_v1.3.0.md)
- [Release Commands v1.3.0](../../RELEASE_COMMANDS_v1.3.0.md)
- [Test Report v1.3.0](../../TEST_REPORT_v1.3.0.md)
- [Migration Guide](../../MIGRATION.md)

---

## 贡献方式

- 在 [GitHub Issues](https://github.com/fndome/xb/issues) 提交问题 / 功能建议
- 参与 [Discussions](https://github.com/fndome/xb/discussions) 分享想法
- 阅读 [CONTRIBUTING](./CONTRIBUTING.md) 了解提 PR 流程
- 运行 `go test ./...` 并同步更新相关文档 / 测试

我们的目标是让中文与英文文档保持同步，如发现遗漏，欢迎直接提 PR 或讨论。谢谢支持！
