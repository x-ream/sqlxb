# 构建器最佳实践

本文档是 `xb/doc/BUILDER_BEST_PRACTICES.md` 的中文版本。它总结了团队在实践中使用的约定，以保持构建器的可预测性和可读性。

---

## 1. 建模输入

- **使用请求 DTO** – 只暴露你需要的字段（例如，`ListUserRequest`）。
- **可选过滤器优先使用指针** – `*uint64` 或 `*string` 可以区分"未设置"和"零值"。
- **规范化枚举** – 在链式调用 `Eq` 之前，将用户字符串转换为类型化常量。

---

## 2. 链式风格

| 技巧 | 为什么有帮助 |
|------|------------|
| 将相关条件分组 | 更容易扫描，可通过辅助函数复用 |
| 使用 `Cond(func(cb *CondBuilder))` | 比手动括号堆叠 `Or()` 调用更好 |
| 保持方法顺序一致 | `Select → From → Join → Where → Sort → Limit` 镜像 SQL |
| 避免在闭包内产生副作用 | 构建器是纯数据；将 IO 保持在外部 |

---

## 3. 可复用块

```go
func addTenantGuard(b *xb.Builder, tenantID uint64) *xb.Builder {
    return b.Eq("tenant_id", tenantID)
}

func addPagination(b *xb.Builder, req *PageRequest) *xb.Builder {
    return b.Limit(req.Limit()).Offset(req.Offset())
}
```

- 将常见片段（如租户守卫、软删除过滤器或默认排序）打包。
- 尽早组合它们，避免忘记必需的条件。

---

## 4. 可观测性

- 使用 `Meta(func(*interceptor.Metadata))` 嵌入 `TraceID`、`UserID` 或 `RequestID`。
- 注册全局拦截器以在执行前记录 SQL/JSON。
- 在单元测试中包含 `built.Raw()` 转储，以便快速进行回归诊断。

---

## 5. 安全护栏

- `InRequired`、`Bool`、`X` 和 `Sub` 在自动过滤不够时提供精确控制。
- 对破坏性操作使用特定上下文的辅助方法，确保永远不会意外运行 `DELETE FROM table`。
- 用返回类型化 DTO 而不是原始 map 的仓库函数包装构建器使用。

---

## 6. 延伸阅读

- `doc/cn/FILTERING.md`
- `doc/cn/CUSTOM_INTERFACE.md`
- `doc/cn/TESTING_STRATEGY.md`

