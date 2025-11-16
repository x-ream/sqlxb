# Qdrant API 同步策略

本文档是 `xb/doc/QDRANT_API_SYNC_STRATEGY.md` 的中文版本。它记录了项目如何保持 xb 的 Qdrant 实现与上游更改同步。

---

## 流程

1. **监控发布** – 订阅 Qdrant 发布说明和 API 变更日志。
2. **差异模式** – 每当发布时，将官方 JSON 负载与 xb 快照进行比较。
3. **添加功能标志** – 在 `QdrantCustom` 辅助方法后面控制有风险的添加（例如，payload 投影）。
4. **通过测试验证** – 使用新固定装置扩展 `xb/qdrant_test.go` 和 `xb/qdrant_custom_test.go`。

---

## 版本映射

| Qdrant 版本 | xb 要求 | 说明 |
|------------|--------|------|
| 1.6+ | `JsonOfSelect` 统一 | 支持 Recommend/Discover/Scroll |
| 1.4–1.5 | 某些 API 需要兼容模式 | 如需要，使用 `CustomLegacy()` |
| <1.4 | 未正式测试 | 依赖快照或自定义 fork |

---

## 工具

- `./scripts/qdrant-schema-sync.sh`（内部）收集 JSON 样本。
- 快照测试（`testdata/qdrant/*.json`）确保没有回归。
- 发布检查清单要求在 Qdrant 宣布破坏性更改时更新本文档。

---

## 相关文档

- `doc/cn/QDRANT_GUIDE.md`
- `doc/cn/QDRANT_ADVANCED_API.md`
- `doc/cn/VECTOR_DIVERSITY_QDRANT.md`

