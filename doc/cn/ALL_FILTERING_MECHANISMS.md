# 所有过滤机制

本文档是 `xb/doc/ALL_FILTERING_MECHANISMS.md` 的中文版本。它解释了 xb 如何自动跳过空输入，`InRequired` 如何保护破坏性操作，以及当内部条件全部失效时，高级块的行为。

---

## 1. 单值守卫

- `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`,  `Like`, `LikeLeft`
- 空字符串、零值数字、`false` 布尔值和 `nil` 指针会被忽略。
- `time.Time` 值总是会被序列化（转换为 `YYYY-MM-DD HH:MM:SS`）。

```go
xb.Of("t_user").
    Eq("status", 0).      // 被跳过
    Eq("status", 1).      // 被包含
    Like("name", "").     // 被跳过
    Like("name", "ai").   // 被包含
    Build()
```

---

## 2. `IN` / `NOT IN`

- 如果所有参数都是零值、空值或 `nil`，整个子句会被移除。
- 混合列表在渲染前会被清理。
- `InRequired` 在剩余列表为空时会 panic，这对于管理员批量操作非常有用。

```go
xb.Of("orders").
    In("id", 0, nil, 9, 10).         // 渲染为 IN (9,10)

xb.Of("orders").
    InRequired("id", ids...).        // 如果 ids 被清理后为空则 panic
```

---

## 3. 复合块

- `Or(func(*CondBuilder))`, `And(func(*CondBuilder))`, `Cond(func(*CondBuilder))`
- 如果嵌套构建器产生零个有效条件，整个块会被丢弃，避免出现 `WHERE ()`。

---

## 4. LIKE 辅助方法

- `Like`, `NotLike`, `LikeLeft` 都会跳过空字符串。
- `doLike` 会注入正确的通配符位置（`%foo%`, `foo%`）。

---

## 5. 调试技巧

| 症状 | 解释 |
|------|------|
| 条件缺失 | 值被自动过滤；检查 `built.Conds` |
| IN 消失 | 输入被清理为零个元素 |
| OR 块缺失 | 每个嵌套条件都被跳过 |
| 需要严格强制执行 | 切换到 `InRequired` 或添加 `X()` 原始表达式 |

在测试中使用 `fmt.Printf("%#v\n", built.Conds)` 来检查最终的 AST。

---

## 6. 相关材料

- `doc/cn/FILTERING.md` – 精简概述
- `xb/doc/EMPTY_OR_AND_FILTERING.md` – 旧版中文参考
- `doc/cn/TESTING_STRATEGY.md` – 在测试中断言自动过滤的想法

