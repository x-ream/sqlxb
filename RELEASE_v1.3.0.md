# xb v1.3.0 Release Notes

**Release Date**: 2025-11-15

---

## 🎉 Overview

v1.3.0 完成了向量方言的统一入口：所有 Qdrant 查询（包括 Recommend / Discover / Scroll）都通过 `JsonOfSelect()` 输出，再无 `ToQdrant*JSON()` 记忆负担。配套的文档、迁移指南、测试与发布资产也同步更新。

**Core Theme**: **Single Entry Vector Builder** — 一个 API 覆盖全部高级场景。

---

## ✨ What's New

### 1️⃣ JsonOfSelect() 统一入口

- `ToQdrantJSON`, `ToQdrantRecommendJSON`, `ToQdrantDiscoverJSON`, `ToQdrantScrollJSON` 退役为内部函数。
- `JsonOfSelect()` 自动识别 `QdrantCustom` 的 Recommend / Discover / Scroll 配置并输出正确的 JSON。
- 错误信息更明确：缺失 `Custom` 会直接返回提示，避免误用。

### 2️⃣ 高级 API 插件化

- `QdrantCustom.Recommend/Discover/ScrollID` 会自动往条件链注入隐藏字段，不需要手写 `Bb`。
- `applyAdvancedConfig()` 现在在 `Generate()` 入口调用，确保所有 JSON 构建路径都能识别高级配置。
- 推荐/探索/滚动的 JSON 生成器被收拢为私有实现，避免用户侧调用混乱。

### 3️⃣ 回归测试

- 新增 `TestJsonOfSelect_WithRecommendConfig/WithDiscoverConfig/WithScrollConfig`，覆盖三类高级请求。
- 旧测试文件（`qdrant_test.go`, `qdrant_nil_filter_test.go`, `empty_or_and_test.go`）同步切换为 `JsonOfSelect()`。
- `go test ./...` 通过，验证 CTE/UNION/向量/自定义拦截器等历史能力无回归。

### 4️⃣ 文档 & 迁移

- README 顶部 “Latest” 区块更新为 v1.3.0，插图示例改用 `JsonOfSelect()`。
- `MIGRATION.md` 新增 v1.2.x → v1.3.0 小节，附上替换表与示例代码。
- 所有 doc/ai_application 资料批量替换 `ToQdrant*` 为 `JsonOfSelect()`，保持叙述一致。

---

## 🔒 Internal Improvements

- `QdrantCustom.Generate()` 在 SELECT 分支新增推荐/探索/滚动路由。
- `ensureQdrantAdvanced()` 与 `mergeAndSerialize()` 保持复用，避免多份 JSON 拼装逻辑。
- 批量脚本清理文档中的旧 API 名称，减少未来维护成本。

---

## 📚 Documentation & Assets

- `README.md`、`MIGRATION.md`、doc/* 全量更新。
- 新增 `RELEASE_v1.3.0.md`, `TEST_REPORT_v1.3.0.md`, `RELEASE_COMMANDS_v1.3.0.md`。
- `commit_message/CHANGELOG.md` 将在发布提交时追加对应条目。

---

## 🧪 Testing

- `go test ./...` — ✅ Pass（包含新的 Qdrant JsonOfSelect 回归测试）
- **重点验证**
  - Recommend/Discover/Scroll JSON 结构
  - `applyAdvancedConfig` 对 `JsonOfSelect()` 的影响
  - 旧 SQL 构建功能（CTE、UNION、Meta）无回归

---

## 🔄 Migration Guide

1. `go get github.com/fndome/xb@v1.3.0`
2. 将所有 `built.ToQdrant*JSON()` 替换为 `built.JsonOfSelect()`。
3. 保持 `QdrantCustom` 配置不变，`JsonOfSelect()` 会自动识别。

> 详见根目录 `MIGRATION.md`。

---

## 📦 What's Included

- 新增测试：`TestJsonOfSelect_WithRecommendConfig` / `WithDiscoverConfig` / `WithScrollConfig`
- 文档更新：README、MIGRATION、doc/*
- 发布资产：Release Notes、Test Report、Release Commands
- Bugfix：`QdrantCustom.Generate()` 正确路由高级 API

---

## 🎯 Summary

v1.3.0 让 Qdrant + xb 的体验回到“只记住一个 API”的初心：

- ✅ `JsonOfSelect()` 统一所有向量查询
- ✅ Qdrant 高级 API 插件化、免样板
- ✅ 文档与迁移指南全套同步
- ✅ 可靠的回归测试护航

**立即升级，减轻团队记忆负担，同时获得更强的高级检索能力。**


