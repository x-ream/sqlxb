# sqlxb 测试策略与回归测试

## 🐛 v0.9.1 Bug 复盘

### 为什么会出现这些 bug？

#### 1. **测试分裂问题** 🔴

```
传统查询 (SqlOfSelect):
  ✅ 有 And()/Or() 测试
  ✅ 使用正确的 toCondSql()
  
向量查询 (SqlOfVectorSearch):
  ❌ 没有 And()/Or() 测试          ⬅️ 测试缺失
  ❌ 使用简化的 buildConditionSql() ⬅️ 临时实现
```

**教训**：
- 新功能（向量查询）只测试了基础场景
- 没有做**跨功能组合测试**
- 临时实现（buildConditionSql）的 TODO 被遗忘

---

#### 2. **数值类型零值测试缺失** 🔴

```go
// ❌ 之前的测试都是 int/string
Gt("rank", 0)     // int
Eq("name", "")    // string

// ✅ 但缺少 float64/float32
Gt("score", 0.0)  // float64 - 未测试！
Lt("rate", 0.0)   // float32 - 未测试！
```

**教训**：
- 测试没有覆盖所有数值类型
- `interface{} == 0` 对 float64 无效，但未被发现

---

#### 3. **API 一致性测试缺失** 🔴

```go
// ❌ 没有对比测试
SqlOfSelect()        // 正确
SqlOfVectorSearch()  // 有 bug

// 两者应该使用相同的条件构建逻辑
// 但没有测试验证一致性
```

**教训**：
- 多个相似 API 应该有一致性测试
- 确保它们对相同查询条件的处理一致

---

## ✅ v0.9.1 回归测试套件

### 新增测试（`regression_test.go`）

1. **TestRegression_README_AndOr**
   - 验证 README 中的 `And()/Or()` 示例
   - 确保文档中的代码真的能用

2. **TestRegression_Float64_ZeroFilter**
   - 测试所有数值类型的零值过滤
   - `float64`, `float32`, `int`, `int64` 等

3. **TestRegression_VectorSearch_WithAndOr**
   - 向量查询 + `And()/Or()` 组合测试
   - 确保 `SqlOfVectorSearch` 正确处理子查询

4. **TestRegression_SqlOfSelect_vs_SqlOfVectorSearch**
   - API 一致性测试
   - 相同条件在两个 API 中应该有一致的行为

5. **TestRegression_EmptyAndOr_AllQueryTypes**
   - 空 `And()/Or()` 在所有查询类型中的过滤
   - `SqlOfSelect`, `SqlOfVectorSearch`, `ToQdrantRequest`

6. **TestRegression_NestedAndOr**
   - 嵌套 `And()/Or()` 测试
   - 确保复杂嵌套也能正确处理

---

## 📋 测试改进策略

### 1. **组合测试矩阵** ⭐

#### 查询类型 × 条件类型

|               | SqlOfSelect | SqlOfVectorSearch | ToQdrantRequest |
|---------------|-------------|-------------------|-----------------|
| 基础条件       | ✅          | ✅                | ✅              |
| And()         | ✅          | ✅ (新增)         | ✅              |
| Or()          | ✅          | ✅ (新增)         | ✅              |
| 嵌套 And/Or   | ✅          | ✅ (新增)         | ✅              |
| 空 And/Or     | ✅          | ✅ (新增)         | ✅              |

#### 数值类型 × 过滤操作

|           | Eq | Gt | Lt | Gte | Lte | In |
|-----------|----|----|----|----|-----|----|
| int       | ✅ | ✅ | ✅ | ✅  | ✅  | ✅ |
| int64     | ✅ | ✅ | ✅ | ✅  | ✅  | ✅ |
| float64   | ✅ | ✅ (新增) | ✅ (新增) | ✅ (新增) | ✅ (新增) | ✅ |
| float32   | ✅ | ✅ (新增) | ✅ (新增) | ✅ (新增) | ✅ (新增) | ✅ |
| string    | ✅ | ❌ | ❌ | ❌  | ❌  | ✅ |
| bool      | ✅ | ❌ | ❌ | ❌  | ❌  | ❌ |

---

### 2. **自动化测试生成** 🤖

#### 使用表格驱动测试

```go
func TestZeroValueFiltering(t *testing.T) {
    tests := []struct {
        name     string
        value    interface{}
        shouldFilter bool
    }{
        {"int_zero", 0, true},
        {"int_nonzero", 100, false},
        {"float64_zero", 0.0, true},
        {"float64_nonzero", 0.5, false},
        {"float32_zero", float32(0.0), true},
        {"float32_nonzero", float32(0.5), false},
        {"string_empty", "", true},
        {"string_nonempty", "test", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            builder := Of(&TestType{}).Gt("field", tt.value).Build()
            sql, args, _ := builder.SqlOfSelect()
            
            if tt.shouldFilter {
                if containsString(sql, "field") {
                    t.Errorf("Expected field to be filtered for %v", tt.value)
                }
            } else {
                if !containsString(sql, "field") {
                    t.Errorf("Expected field to exist for %v", tt.value)
                }
            }
        })
    }
}
```

---

### 3. **API 一致性断言** 🔍

#### 创建通用测试辅助函数

```go
// 验证多个 API 对相同查询的处理一致性
func AssertAPIConsistency(t *testing.T, builder *Built) {
    t.Helper()
    
    // SqlOfSelect
    sql1, args1, _ := builder.SqlOfSelect()
    
    // SqlOfVectorSearch (如果有向量检索)
    sql2, args2 := builder.SqlOfVectorSearch()
    
    // 验证：Or/And 子查询数量应该一致
    orCount1 := strings.Count(sql1, "OR")
    orCount2 := strings.Count(sql2, "OR")
    
    if orCount1 != orCount2 {
        t.Errorf("OR count mismatch: SqlOfSelect=%d, SqlOfVectorSearch=%d", 
            orCount1, orCount2)
    }
    
    // 验证：标量条件数量应该一致（除了向量参数）
    // ...
}
```

---

### 4. **代码覆盖率监控** 📊

#### 关键目标

- **总体覆盖率**: ≥ 85%
- **核心模块覆盖率**: ≥ 95%
  - `cond_builder.go`: 95%
  - `to_sql.go`: 95%
  - `to_vector_sql.go`: 95%
  - `to_qdrant_json.go`: 95%

#### 运行覆盖率测试

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

### 5. **持续集成检查** 🔄

#### GitHub Actions 工作流

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      
      # 所有测试
      - name: Run all tests
        run: go test -v ./...
      
      # 回归测试
      - name: Run regression tests
        run: go test -v -run TestRegression ./...
      
      # 覆盖率
      - name: Coverage
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      # 确保覆盖率 >= 85%
      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 85" | bc -l) )); then
            echo "Coverage $coverage% is below 85%"
            exit 1
          fi
```

---

## 🎯 测试原则

### 1. **文档即测试**
- README 中的所有示例代码都应该有对应的测试
- 确保文档中的代码真的能用

### 2. **组合优先**
- 不仅测试单个功能，更要测试功能组合
- 例如：向量查询 + And/Or + 多样性 + Qdrant

### 3. **边界条件**
- nil, 0, "", []
- 空 And/Or
- 单条件、多条件、嵌套条件

### 4. **类型全覆盖**
- 所有数值类型（int, int64, float32, float64, ...）
- 所有操作符（Eq, Gt, Lt, In, Like, ...）
- 所有组合（And, Or, 嵌套）

### 5. **一致性验证**
- 相似 API 应该有一致的行为
- `SqlOfSelect` vs `SqlOfVectorSearch`
- PostgreSQL vs Qdrant

---

## 📝 测试检查清单

### 新功能开发时

- [ ] 基础功能测试
- [ ] 与现有功能的组合测试
- [ ] 边界条件测试
- [ ] 错误处理测试
- [ ] README 示例测试
- [ ] API 一致性测试

### 代码评审时

- [ ] 是否有足够的测试？
- [ ] 是否测试了组合场景？
- [ ] 是否测试了所有类型？
- [ ] 是否更新了 README？
- [ ] 是否添加了回归测试？

---

## 🚀 未来改进

1. **性能测试 (Benchmark)**
   - 测试大数据量下的性能
   - 对比不同 API 的性能差异

2. **并发测试**
   - 测试 Builder 的线程安全性
   - 测试连接池的并发访问

3. **Fuzzing 测试**
   - 使用 Go 1.18+ 的 Fuzzing
   - 自动发现边界条件 bug

4. **集成测试**
   - 真实数据库连接测试
   - PostgreSQL pgvector 集成测试
   - Qdrant 集成测试

---

**结论**：

测试不仅仅是验证代码正确性，更重要的是**预防回归**。

v0.9.1 的 bug 提醒我们：
- ✅ 新功能必须有组合测试
- ✅ 临时实现必须有 TODO 追踪
- ✅ 所有类型都要覆盖
- ✅ API 一致性必须验证

**这套回归测试将确保这些 bug 永不再现！** 🎯

