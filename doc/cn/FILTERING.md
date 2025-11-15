# 过滤与校验（中文）

xb 会自动跳过空输入，避免到处写 `if value == ""`。本文列出哪些值会被忽略、如何覆盖默认行为，以及在何处加入校验逻辑。

---

## 1. 单值条件

适用方法：`Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `Between`, `Like`, `LikeLeft`, `LikeRight`

| 类型 | 被跳过的值 |
|------|------------|
| `string` | `""` |
| 数值（int/uint/float） | `0` |
| `bool` | `false` |
| 指针 | `nil` |
| `time.Time` | 不跳过（会格式化为字符串） |

示例：

```go
builder := xb.Of("t_user").
    Eq("status", 0).   // 被忽略
    Eq("status", 1).   // 生效
    Like("name", "").  // 被忽略
    Like("name", "ai") // 生效
```

---

## 2. `IN / NOT IN`

`In`, `NotIn`, `InRequired`

- 如果参数列表为空或全部为空值，整个调用被跳过。
- nil 指针、零值会从列表里移除。
- `InRequired` 会在结果为空时 panic，用于强制 guard。

```go
builder := xb.Of("t").
    In("id", 0, nil, 9, 10) // 实际生成 IN (9, 10)
```

---

## 3. 组合块

- `Or(func(cb *CondBuilder))`
- `And(func(cb *CondBuilder))`
- `Cond(func(cb *CondBuilder))`

若内部没有有效条件，则整个块被丢弃（不会产生空括号）。

---

## 4. 校验钩子

使用拦截器或 `Meta(func(meta *interceptor.Metadata))` 注入自定义校验。

```go
xb.RegisterBeforeBuild(func(built *xb.Built) error {
    if built.HasInRequiredViolation() {
        return errors.New("missing required IDs")
    }
    return nil
})
```

可用于：

- 确保租户/组织 ID 必填
- 强制分页限制
- 审计 metadata

---

## 5. 排障

| 现象 | 解决办法 |
|------|----------|
| SQL 中缺少条件 | 检查是否被自动跳过 |
| IN 为空 | 使用 `InRequired` 捕获 |
| OR 块消失 | 内部条件都被跳过 |
| 校验无效 | 确保在 `Build()` 前注册拦截器 |

可通过 `fmt.Printf("%#v\n", built.Conds)` 查看内部状态。

---

## 6. 相关文档

- `QUICKSTART.md`：链式语法
- `CUSTOM_INTERFACE.md`：方言自定义校验
- `QDRANT_GUIDE.md`：过滤如何映射到 Qdrant JSON

若发现某类场景不该被跳过，欢迎提交 issue 讨论。

