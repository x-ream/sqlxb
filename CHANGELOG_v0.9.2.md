# v0.9.2 更新日志

**发布日期**: 2025-10-27

---

## ✨ 新增功能

### 1. Interceptor 拦截器

独立的 `interceptor/` 包，用于全局观察 SQL 构建过程。

**特点**：
- ✅ 类型安全：`BeforeBuild(meta *Metadata)` 编译时限制
- ✅ 全局性：注册一次，所有查询生效
- ✅ 用途：日志、监控、审计
- ❌ 不用于：业务逻辑（多租户、权限应显式写）

**文件**：
- `interceptor/metadata.go`
- `interceptor/interceptor.go`
- `interceptor/registry.go`
- `interceptor_test.go` (5 个测试)

---

### 2. 扩展指南

新增两份扩展指南：
- `doc/CUSTOM_VECTOR_DB_GUIDE.md` - 如何支持 Milvus, Weaviate 等
- `doc/CUSTOM_JOINS_GUIDE.md` - 如何扩展自定义 JOIN

---

## 🧹 文档清理

删除 7 个过时文档，更新所有 API 示例。

---

## ✅ 测试

- 新增 5 个 Interceptor 测试 ✅
- 所有现有测试通过 ✅
- 100% 向后兼容 ✅

---

**升级**：`go get github.com/fndome/xb@v0.9.2`

