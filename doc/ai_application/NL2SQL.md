# 自然语言查询转换 (实验性)

## ⚠️ 免责声明

本功能处于**实验性阶段**，不建议用于生产环境。自然语言到 SQL 的转换存在以下风险：
- 语义理解可能不准确
- 生成的查询可能不安全
- 性能可能不理想

**推荐**: 对于生产应用，请使用预定义的查询模板或 [AGENT_TOOLKIT.md](./AGENT_TOOLKIT.md) 中的结构化方法。

## 📋 概述

NL2SQL 允许用户用自然语言描述查询需求，自动转换为 sqlxb 查询代码。

## 🎯 基础实现

### 使用 GPT-4 生成查询

```go
package nl2sql

import (
    "context"
    "encoding/json"
    openai "github.com/sashabaranov/go-openai"
)

type QueryGenerator struct {
    client *openai.Client
    schema SchemaInfo
}

type SchemaInfo struct {
    TableName string
    Fields    []FieldInfo
}

type FieldInfo struct {
    Name        string
    Type        string
    Description string
    Enum        []string
}

func NewQueryGenerator(apiKey string, schema SchemaInfo) *QueryGenerator {
    return &QueryGenerator{
        client: openai.NewClient(apiKey),
        schema: schema,
    }
}

func (g *QueryGenerator) GenerateQuery(ctx context.Context, naturalQuery string) (string, error) {
    prompt := g.buildPrompt(naturalQuery)
    
    resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem,
                Content: systemPrompt,
            },
            {
                Role:    openai.ChatMessageRoleUser,
                Content: prompt,
            },
        },
        Temperature: 0,
    })
    
    if err != nil {
        return "", err
    }
    
    return resp.Choices[0].Message.Content, nil
}

func (g *QueryGenerator) buildPrompt(query string) string {
    schemaJSON, _ := json.MarshalIndent(g.schema, "", "  ")
    
    return fmt.Sprintf(`根据以下数据库表结构，将自然语言查询转换为 sqlxb 查询代码。

表结构:
%s

自然语言查询: %s

请生成 Go 代码（只包含 sqlxb 查询部分）:`, schemaJSON, query)
}

const systemPrompt = `你是一个数据库查询专家。你的任务是将自然语言查询转换为 sqlxb 查询代码。

规则:
1. 只生成 sqlxb 查询代码，不要包含其他内容
2. 使用正确的字段名和类型
3. 对于模糊匹配使用 Like()
4. 对于精确匹配使用 Eq()
5. 对于范围查询使用 Gte()/Lte()
6. 对于多值匹配使用 In()
7. 始终添加适当的 Limit()

示例:
输入: "查找所有活跃用户"
输出: sqlxb.Of(&User{}).Eq("status", "active").Limit(100).Build()

输入: "查找年龄在18到30岁之间的用户"
输出: sqlxb.Of(&User{}).Gte("age", 18).Lte("age", 30).Limit(100).Build()
`
```

### 使用示例

```go
func main() {
    // 定义表结构
    schema := nl2sql.SchemaInfo{
        TableName: "users",
        Fields: []nl2sql.FieldInfo{
            {
                Name:        "id",
                Type:        "int64",
                Description: "用户 ID",
            },
            {
                Name:        "username",
                Type:        "string",
                Description: "用户名",
            },
            {
                Name:        "status",
                Type:        "string",
                Description: "账户状态",
                Enum:        []string{"active", "inactive", "banned"},
            },
            {
                Name:        "age",
                Type:        "int",
                Description: "年龄",
            },
        },
    }
    
    generator := nl2sql.NewQueryGenerator("your-api-key", schema)
    
    // 自然语言查询
    queries := []string{
        "查找所有活跃用户",
        "找出年龄大于25岁的用户",
        "搜索用户名包含 admin 的账户",
    }
    
    for _, q := range queries {
        code, err := generator.GenerateQuery(context.Background(), q)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("查询: %s\n", q)
        fmt.Printf("代码: %s\n\n", code)
    }
}
```

## 🎯 RAG 查询生成

### 向量检索查询生成

```go
func (g *QueryGenerator) GenerateVectorQuery(ctx context.Context, naturalQuery string) (string, error) {
    prompt := fmt.Sprintf(`将自然语言查询转换为 sqlxb 向量检索查询。

查询: %s

生成包含以下步骤的代码:
1. 调用 embedding 函数获取查询向量
2. 使用 VectorSearch() 进行向量检索
3. 添加适当的标量过滤条件
4. 设置 Top-K 和分数阈值

示例输出:
queryVector, _ := embedText(query)
built := sqlxb.Of(&DocumentChunk{}).
    VectorSearch("embedding", queryVector, 10).
    Eq("language", "zh").
    Build()
result, _ := built.ToQdrantJSON()
`, naturalQuery)
    
    // 调用 LLM...
}
```

## 🔒 安全控制

### 查询验证

```go
type QueryValidator struct {
    allowedOperations []string
    maxLimit          int
    allowedFields     []string
}

func (v *QueryValidator) Validate(generatedCode string) error {
    // 1. 检查是否包含危险操作
    dangerousOps := []string{"Delete(", "Drop(", "Truncate("}
    for _, op := range dangerousOps {
        if strings.Contains(generatedCode, op) {
            return fmt.Errorf("dangerous operation detected: %s", op)
        }
    }
    
    // 2. 检查 Limit
    if !strings.Contains(generatedCode, "Limit(") {
        return fmt.Errorf("missing Limit() call")
    }
    
    // 3. 提取并验证字段名
    fields := extractFields(generatedCode)
    for _, field := range fields {
        if !contains(v.allowedFields, field) {
            return fmt.Errorf("field not allowed: %s", field)
        }
    }
    
    return nil
}
```

### 沙箱执行

```go
// 在隔离环境中执行生成的查询
func ExecuteInSandbox(generatedCode string, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    // 创建临时文件
    tmpFile := "/tmp/query_" + generateID() + ".go"
    ioutil.WriteFile(tmpFile, []byte(wrapCode(generatedCode)), 0644)
    defer os.Remove(tmpFile)
    
    // 编译
    cmd := exec.CommandContext(ctx, "go", "build", "-o", "/tmp/query", tmpFile)
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("compilation failed: %w", err)
    }
    
    // 执行
    cmd = exec.CommandContext(ctx, "/tmp/query")
    output, err := cmd.CombinedOutput()
    
    return string(output), err
}
```

## 📊 实际示例

### 完整流程

```go
func NL2SQLDemo() {
    // 1. 用户输入自然语言
    userInput := "查找最近7天发布的关于人工智能的中文文章，按相关性排序"
    
    // 2. 生成查询代码
    generator := nl2sql.NewQueryGenerator(apiKey, schema)
    code, _ := generator.GenerateVectorQuery(context.Background(), userInput)
    
    fmt.Println("生成的代码:")
    fmt.Println(code)
    // 输出:
    // queryVector, _ := embedText("人工智能")
    // sevenDaysAgo := time.Now().AddDate(0, 0, -7)
    // result, _ := sqlxb.Of(&Article{}).
    //     VectorSearch("embedding", queryVector).
    //     Eq("language", "zh").
    //     Eq("category", "tech").
    //     Gte("published_at", sevenDaysAgo).
    //     QdrantX(func(qx *sqlxb.QdrantBuilderX) {
    //         qx.ScoreThreshold(0.7)
    //     }).
    //     Build().ToQdrantJSON()
    
    // 3. 验证查询
    validator := &QueryValidator{
        allowedFields: []string{"language", "category", "published_at"},
        maxLimit:      100,
    }
    
    if err := validator.Validate(code); err != nil {
        log.Fatal("查询验证失败:", err)
    }
    
    // 4. 执行查询（在实际应用中）
    // results := executeQuery(code)
}
```

## 🎓 最佳实践

1. **限制使用场景**
   - 仅用于内部工具或演示
   - 不要暴露给最终用户
   - 总是人工审核生成的查询

2. **强制安全检查**
   - 验证所有生成的代码
   - 限制可用字段和操作
   - 设置查询超时

3. **提供回退方案**
   - 准备预定义查询模板
   - 生成失败时使用模板
   - 记录失败案例用于改进

4. **持续改进**
   - 收集用户反馈
   - 优化 Prompt
   - 扩展示例库

## 🚀 未来方向

- [ ] 支持更复杂的 JOIN 查询
- [ ] 自动索引建议
- [ ] 查询优化建议
- [ ] 多轮对话式查询构建

---

**警告**: 请勿在生产环境中直接使用未经验证的自动生成代码。

