# xb v1.2.3 Release Notes

**Release Date**: 2025-11-09

---

## 🎉 Overview

v1.2.3 turns `xb` into a full-featured analytical query builder. Common Table Expressions, recursive hierarchies, and UNION pipelines are now first-class citizens—still within the “don’t add concepts” philosophy.

**Core Theme**: **Composable SQL Pipelines** — build complex statements without dropping to raw strings.

---

## ✨ What's New

### 1️⃣ Fluent CTE Builders

- `With(name, fn)` schedules a CTE block using the normal builder chain.
- `WithRecursive(name, fn)` adds recursive CTE support with a single keyword.
- Works seamlessly with existing condition helpers, sorting, paging, and metadata.

```go
report := xb.Of("recent_orders").As("ro").
    With("recent_orders", func(sb *xb.BuilderX) {
        sb.From("orders o").
            Select("o.id", "o.user_id").
            Gt("o.created_at", since30Days)
    }).
    WithRecursive("team_hierarchy", func(sb *xb.BuilderX) {
        sb.From("users u").
            Select("u.id", "u.manager_id").
            Eq("u.active", true)
    }).
    Build()
```

### 2️⃣ UNION Chaining

- `UNION(kind, fn)` merges result sets without writing SQL manually.
- `ALL()` helper emits `UNION ALL`; omitting `kind` defaults to `UNION` (DISTINCT).
- UNION blocks execute after the primary SELECT core, preserving ORDER BY / LIMIT semantics.

### 3️⃣ Metadata Injection

- `Meta(func(meta *interceptor.Metadata))` lets you attach TraceID, tenant info, or custom labels inline.
- Interceptors receive the populated metadata before executing `BeforeBuild`.

### 4️⃣ SQL Safety & Quality

- Alias normalization ensures `From("cte").As("alias")` always generates valid `FROM` clauses.
- Internal constants renamed (`DISTINCT` → `DISTINCT_SCRIPT`) to prevent naming collisions with new helpers.

---

## 🔒 Internal Improvements

- Shared SELECT writer (`writeSelectCore`) centralizes projection, FROM, WHERE, GROUP BY, and HAVING generation.
- CTE and UNION rendering functions reuse argument slices, guaranteeing parameter ordering correctness.
- Builder state now caches built CTE/UNION clauses to avoid duplicate Build() calls.

---

## 📚 Documentation & Assets

- `README.md` refreshed with v1.2.3 hero section, CTE + UNION examples, and observability tips.
- `CHANGELOG`, release commands, and test report updated for the new release.
- New regression tests: `with_cte_test.go`, `union_test.go`.

---

## 🧪 Testing

- **go test ./...** — ✅ Pass (approx. 240 unit tests including new suites)
- **Key Focus Areas**
  - CTE default/recursive pipelines
  - UNION DISTINCT vs UNION ALL composition
  - Alias normalization and argument ordering
  - Metadata pass-through during interception

---

## 🔄 Migration Guide

No breaking changes. Existing APIs continue to work as before.

**Recommended actions**
1. Adopt `With()` / `WithRecursive()` for complex reports.
2. Replace ad-hoc SQL unions with the new `UNION()` helper when possible.
3. Use `Meta(func)` to enrich downstream logging/tracing.

---

## 📦 What's Included

- New APIs: `With`, `WithRecursive`, `UNION`, `ALL`, `Meta(func)`.
- Struct additions: `WithClause`, `UnionClause`, alias preservation on `Built`.
- Updated helpers: renamed constants, shared SQL writers, README updates.

---

## 🎯 Design Philosophy

- **Extend without bloat**: CTE/UNION integrate into the existing fluent builder—no new structs for callers to memorize.
- **Observability first**: Metadata DSL keeps middleware hooks type-safe.
- **Zero-breaking changes**: All upgrades are opt-in, existing projects keep running untouched.

---

## 🚀 Summary

v1.2.3 unlocks enterprise-grade SQL composition:

- ✅ CTE + Recursive pipelines
- ✅ UNION DISTINCT / ALL chaining
- ✅ Metadata & interceptor synergy
- ✅ Fully tested, fully documented

**Upgrade now and build richer analytics with the same, simple API.** 🚀

