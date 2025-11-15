# Builder Best Practices (English)

English rewrite of `xb/doc/BUILDER_BEST_PRACTICES.md`. It captures conventions field teams use to keep builders predictable and readable.

---

## 1. Modeling inputs

- **Use request DTOs** – expose only the fields you need (e.g., `ListUserRequest`).
- **Prefer pointers for optional filters** – `*uint64` or `*string` distinguishes “unset” from “zero”.
- **Normalize enums** – convert user strings to typed constants before chaining `Eq`.

---

## 2. Chaining style

| Tip | Why it helps |
|-----|--------------|
| Group related conditions | Easier to scan, reuse via helper functions |
| Use `Cond(func(cb *CondBuilder))` | Better than stacking `Or()` calls with manual parentheses |
| Keep method ordering consistent | `Select → From → Join → Where → Sort → Limit` mirrors SQL |
| Avoid side effects inside closures | Builders are pure data; keep IO outside |

---

## 3. Reusable blocks

```go
func addTenantGuard(b *xb.Builder, tenantID uint64) *xb.Builder {
    return b.Eq("tenant_id", tenantID)
}

func addPagination(b *xb.Builder, req *PageRequest) *xb.Builder {
    return b.Limit(req.Limit()).Offset(req.Offset())
}
```

- Package common fragments such as tenant guards, soft-delete filters, or default sorts.
- Compose them early to avoid forgetting mandatory conditions.

---

## 4. Observability

- Use `Meta(func(*interceptor.Metadata))` to embed `TraceID`, `UserID`, or `RequestID`.
- Register global interceptors to log SQL/JSON before execution.
- Include `built.Raw()` dumps in unit tests for fast regression diagnosis.

---

## 5. Safety rails

- `InRequired`, `Bool`, `X`, and `Sub` give you precise control when auto-filtering is not enough.
- Use context-specific helper methods for destructive operations to ensure you never run `DELETE FROM table` accidentally.
- Wrap builder usage with repository functions that return typed DTOs, not raw maps.

---

## 6. Further reading

- `doc/en/FILTERING.md`
- `doc/en/CUSTOM_INTERFACE.md`
- `doc/en/TESTING_STRATEGY.md`

