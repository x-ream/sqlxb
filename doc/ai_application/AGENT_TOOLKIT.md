# AI Agent 工具链集成指南

## 📋 概述

本文档介绍如何将 xb 集成到 AI Agent 系统中，使 AI 能够安全、高效地查询和操作数据库。

## 🎯 核心特性

- **JSON Schema 生成**: 为 Function Calling 提供类型定义
- **参数验证**: 自动验证 AI 生成的参数
- **安全控制**: 防止 SQL 注入和危险操作
- **OpenAPI 规范**: 标准化 API 定义

## 🛠️ JSON Schema 生成

### 基础用法

```go
package main

import (
    "encoding/json"
    "github.com/fndome/xb"
)

type User struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Status   string `json:"status"`
    Age      int    `json:"age"`
}

// 生成查询工具的 JSON Schema
func GenerateSearchUserSchema() map[string]interface{} {
    return map[string]interface{}{
        "name": "search_users",
        "description": "搜索用户数据库，支持按用户名、邮箱、状态、年龄等条件过滤",
        "parameters": map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "username": map[string]interface{}{
                    "type":        "string",
                    "description": "用户名（模糊匹配）",
                },
                "email": map[string]interface{}{
                    "type":        "string",
                    "description": "邮箱地址（精确匹配）",
                },
                "status": map[string]interface{}{
                    "type":        "string",
                    "enum":        []string{"active", "inactive", "banned"},
                    "description": "账户状态",
                },
                "min_age": map[string]interface{}{
                    "type":        "integer",
                    "description": "最小年龄",
                    "minimum":     0,
                },
                "max_age": map[string]interface{}{
                    "type":        "integer",
                    "description": "最大年龄",
                    "maximum":     150,
                },
                "limit": map[string]interface{}{
                    "type":        "integer",
                    "description": "返回结果数量（默认10）",
                    "default":     10,
                    "minimum":     1,
                    "maximum":     100,
                },
            },
        },
    }
}

// 执行 AI Agent 的查询请求
func ExecuteSearchUsers(params map[string]interface{}) (string, []interface{}, error) {
    builder := xb.Of(&User{})

    // ⭐ xb 自动过滤 nil/0/空字符串，无需判断
    username, _ := params["username"].(string)
    email, _ := params["email"].(string)
    status, _ := params["status"].(string)
    minAge, _ := params["min_age"].(float64)
    maxAge, _ := params["max_age"].(float64)
    limit, _ := params["limit"].(float64)
    
    if limit == 0 {
        limit = 10  // 默认值
    }

    built := builder.
        Like("username", username).  // ⭐ xb 自动添加 %username%
        Eq("email", email).
        Eq("status", status).
        Gte("age", int(minAge)).
        Lte("age", int(maxAge)).
        Limit(int(limit)).
        Build()
    
    return built.SqlOfSelect()
}
```

### OpenAI Function Calling 集成

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    openai "github.com/sashabaranov/go-openai"
)

func SetupAIAgent(apiKey string) (*openai.Client, error) {
    client := openai.NewClient(apiKey)

    // 注册工具
    tools := []openai.Tool{
        {
            Type: openai.ToolTypeFunction,
            Function: &openai.FunctionDefinition{
                Name:        "search_users",
                Description: "搜索用户数据库，支持按用户名、邮箱、状态、年龄等条件过滤",
                Parameters: json.RawMessage(`{
                    "type": "object",
                    "properties": {
                        "username": {
                            "type": "string",
                            "description": "用户名（模糊匹配）"
                        },
                        "email": {
                            "type": "string",
                            "description": "邮箱地址（精确匹配）"
                        },
                        "status": {
                            "type": "string",
                            "enum": ["active", "inactive", "banned"],
                            "description": "账户状态"
                        },
                        "min_age": {
                            "type": "integer",
                            "description": "最小年龄",
                            "minimum": 0
                        },
                        "max_age": {
                            "type": "integer",
                            "description": "最大年龄",
                            "maximum": 150
                        },
                        "limit": {
                            "type": "integer",
                            "description": "返回结果数量（默认10）",
                            "default": 10,
                            "minimum": 1,
                            "maximum": 100
                        }
                    }
                }`),
            },
        },
    }

    return client, nil
}

// 完整的 AI Agent 对话循环
func RunAIAgentQuery(client *openai.Client, userQuery string) (string, error) {
    ctx := context.Background()

    messages := []openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleSystem,
            Content: "你是一个数据库查询助手，可以帮助用户查询用户信息。",
        },
        {
            Role:    openai.ChatMessageRoleUser,
            Content: userQuery,
        },
    }

    // 第一次调用：AI 决定是否使用工具
    resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:    openai.GPT4,
        Messages: messages,
        Tools:    GetTools(),
    })
    if err != nil {
        return "", err
    }

    // 检查是否需要调用工具
    if len(resp.Choices[0].Message.ToolCalls) > 0 {
        toolCall := resp.Choices[0].Message.ToolCalls[0]

        // 解析参数
        var params map[string]interface{}
        if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
            return "", err
        }

        // 执行查询
        sql, args, err := ExecuteSearchUsers(params)
        if err != nil {
            return "", err
        }

        // 执行 SQL（这里假设你有 db 连接）
        var users []User
        // db.Select(&users, sql, args...)

        // 格式化结果返回给 AI
        resultJSON, _ := json.Marshal(users)

        // 第二次调用：让 AI 总结结果
        messages = append(messages, openai.ChatCompletionMessage{
            Role:       openai.ChatMessageRoleTool,
            Content:    string(resultJSON),
            ToolCallID: toolCall.ID,
        })

        finalResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
            Model:    openai.GPT4,
            Messages: messages,
        })
        if err != nil {
            return "", err
        }

        return finalResp.Choices[0].Message.Content, nil
    }

    return resp.Choices[0].Message.Content, nil
}
```

### 使用示例

```go
func main() {
    client, _ := SetupAIAgent("your-openai-api-key")

    // 用户自然语言查询
    queries := []string{
        "帮我找出所有活跃的用户",
        "查询年龄在18到30岁之间的用户",
        "找出邮箱是 john@example.com 的用户",
        "搜索用户名包含 'admin' 的账户",
    }

    for _, query := range queries {
        fmt.Printf("\n用户查询: %s\n", query)
        response, err := RunAIAgentQuery(client, query)
        if err != nil {
            fmt.Printf("错误: %v\n", err)
            continue
        }
        fmt.Printf("AI 回答: %s\n", response)
    }
}
```

## 🎯 向量检索工具

### RAG 查询工具 Schema

```go
func GenerateRAGSearchSchema() map[string]interface{} {
    return map[string]interface{}{
        "name": "search_knowledge_base",
        "description": "在知识库中搜索与查询相关的文档片段，使用向量相似度匹配",
        "parameters": map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "query": map[string]interface{}{
                    "type":        "string",
                    "description": "用户的查询问题",
                },
                "doc_type": map[string]interface{}{
                    "type":        "string",
                    "description": "文档类型过滤（可选）",
                    "enum":        []string{"tutorial", "api", "blog", "faq"},
                },
                "language": map[string]interface{}{
                    "type":        "string",
                    "description": "语言过滤（可选）",
                    "enum":        []string{"zh", "en"},
                },
                "top_k": map[string]interface{}{
                    "type":        "integer",
                    "description": "返回最相关的 K 个结果",
                    "default":     5,
                    "minimum":     1,
                    "maximum":     20,
                },
                "score_threshold": map[string]interface{}{
                    "type":        "number",
                    "description": "最低相关性分数（0-1）",
                    "default":     0.7,
                    "minimum":     0.0,
                    "maximum":     1.0,
                },
            },
            "required": []string{"query"},
        },
    }
}
```

### RAG 查询执行

```go
type DocumentChunk struct {
    ID        int64     `json:"id"`
    Content   string    `json:"content"`
    Embedding []float32 `json:"embedding"`
    DocType   string    `json:"doc_type"`
    Language  string    `json:"language"`
    Metadata  string    `json:"metadata"`
}

func ExecuteRAGSearch(params map[string]interface{}, embeddingFunc func(string) ([]float32, error)) (map[string]interface{}, error) {
    // 获取查询文本
    query, ok := params["query"].(string)
    if !ok || query == "" {
        return nil, fmt.Errorf("query is required")
    }

    // 生成查询向量
    queryVector, err := embeddingFunc(query)
    if err != nil {
        return nil, err
    }

    // Top-K 和分数阈值
    topK := 5
    if k, ok := params["top_k"].(float64); ok {
        topK = int(k)
    }

    scoreThreshold := 0.7
    if threshold, ok := params["score_threshold"].(float64); ok {
        scoreThreshold = threshold
    }

    // 构建查询
    builder := xb.Of(&DocumentChunk{}).
        VectorSearch("embedding", queryVector, topK)

    // ⭐ xb 自动过滤 nil/0/空字符串，无需判断
    docType, _ := params["doc_type"].(string)
    lang, _ := params["language"].(string)
    
    builder.Eq("doc_type", docType).
            Eq("language", lang)

    // 构建并生成 Qdrant JSON
    built := builder.
        QdrantX(func(qx *xb.QdrantBuilderX) {
            qx.ScoreThreshold(float32(scoreThreshold))
        }).
        Build()

    qdrantJSON, err := built.JsonOfSelect()
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "qdrant_json": qdrantJSON,
        "top_k":       topK,
    }, nil
}
```

## 🔒 安全控制

### 参数验证

```go
type QueryValidator struct {
    MaxLimit      int
    AllowedTables []string
    AllowedFields map[string][]string
}

func (v *QueryValidator) ValidateSearchParams(params map[string]interface{}, tableName string) error {
    // 检查表名白名单
    if !contains(v.AllowedTables, tableName) {
        return fmt.Errorf("table %s is not allowed", tableName)
    }

    // 检查 limit 范围
    if limit, ok := params["limit"].(float64); ok {
        if int(limit) > v.MaxLimit {
            return fmt.Errorf("limit %d exceeds maximum %d", int(limit), v.MaxLimit)
        }
    }

    // 检查字段白名单
    allowedFields := v.AllowedFields[tableName]
    for key := range params {
        if !contains(allowedFields, key) && key != "limit" {
            return fmt.Errorf("field %s is not allowed for table %s", key, tableName)
        }
    }

    return nil
}
```

### 查询审计

```go
type QueryAudit struct {
    Timestamp time.Time              `json:"timestamp"`
    UserID    string                 `json:"user_id"`
    Query     string                 `json:"query"`
    Params    map[string]interface{} `json:"params"`
    SQL       string                 `json:"sql"`
    Duration  time.Duration          `json:"duration"`
    Error     string                 `json:"error,omitempty"`
}

func AuditQuery(ctx context.Context, params map[string]interface{}, fn func() (string, []interface{}, error)) (string, []interface{}, error) {
    audit := &QueryAudit{
        Timestamp: time.Now(),
        UserID:    getUserIDFromContext(ctx),
        Params:    params,
    }

    start := time.Now()
    sql, args, err := fn()
    audit.Duration = time.Since(start)
    audit.SQL = sql

    if err != nil {
        audit.Error = err.Error()
    }

    // 记录审计日志
    logAudit(audit)

    return sql, args, err
}
```

## 📊 OpenAPI 规范生成

### 自动生成 REST API 规范

```go
func GenerateOpenAPISpec() map[string]interface{} {
    return map[string]interface{}{
        "openapi": "3.0.0",
        "info": map[string]interface{}{
            "title":       "User Search API",
            "description": "AI-powered user search API built with xb",
            "version":     "1.0.0",
        },
        "paths": map[string]interface{}{
            "/api/users/search": map[string]interface{}{
                "post": map[string]interface{}{
                    "summary":     "搜索用户",
                    "description": "根据多个条件搜索用户",
                    "requestBody": map[string]interface{}{
                        "required": true,
                        "content": map[string]interface{}{
                            "application/json": map[string]interface{}{
                                "schema": map[string]interface{}{
                                    "$ref": "#/components/schemas/SearchUsersRequest",
                                },
                            },
                        },
                    },
                    "responses": map[string]interface{}{
                        "200": map[string]interface{}{
                            "description": "搜索成功",
                            "content": map[string]interface{}{
                                "application/json": map[string]interface{}{
                                    "schema": map[string]interface{}{
                                        "$ref": "#/components/schemas/SearchUsersResponse",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "components": map[string]interface{}{
            "schemas": map[string]interface{}{
                "SearchUsersRequest": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "username": map[string]interface{}{"type": "string"},
                        "email":    map[string]interface{}{"type": "string"},
                        "status":   map[string]interface{}{"type": "string", "enum": []string{"active", "inactive", "banned"}},
                        "min_age":  map[string]interface{}{"type": "integer"},
                        "max_age":  map[string]interface{}{"type": "integer"},
                        "limit":    map[string]interface{}{"type": "integer", "default": 10},
                    },
                },
                "SearchUsersResponse": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "users": map[string]interface{}{
                            "type": "array",
                            "items": map[string]interface{}{
                                "$ref": "#/components/schemas/User",
                            },
                        },
                        "total": map[string]interface{}{"type": "integer"},
                    },
                },
                "User": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "id":       map[string]interface{}{"type": "integer"},
                        "username": map[string]interface{}{"type": "string"},
                        "email":    map[string]interface{}{"type": "string"},
                        "status":   map[string]interface{}{"type": "string"},
                        "age":      map[string]interface{}{"type": "integer"},
                    },
                },
            },
        },
    }
}
```

## 🧪 测试示例

```go
func TestAIAgentQuery(t *testing.T) {
    tests := []struct {
        name     string
        params   map[string]interface{}
        wantSQL  string
        wantArgs []interface{}
    }{
        {
            name: "简单查询",
            params: map[string]interface{}{
                "status": "active",
                "limit":  10,
            },
            wantSQL:  "SELECT * FROM users WHERE status = ? LIMIT ?",
            wantArgs: []interface{}{"active", 10},
        },
        {
            name: "复杂过滤",
            params: map[string]interface{}{
                "username": "john",
                "min_age":  18,
                "max_age":  30,
                "status":   "active",
            },
            wantSQL: "SELECT * FROM users WHERE username LIKE ? AND age >= ? AND age <= ? AND status = ?",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            sql, args, err := ExecuteSearchUsers(tt.params)
            assert.NoError(t, err)
            assert.Equal(t, tt.wantSQL, sql)
            assert.Equal(t, tt.wantArgs, args)
        })
    }
}
```

## 🎯 最佳实践

### 1. 参数验证
- 始终验证 AI 生成的参数
- 使用白名单而非黑名单
- 限制查询范围（limit, offset）

### 2. 性能优化
- 为常见查询添加索引
- 使用连接池
- 限制返回字段数量

### 3. 错误处理
- 提供友好的错误消息
- 记录所有查询日志
- 实现重试机制

### 4. 安全控制
- 永远不要执行 DELETE/UPDATE（除非明确需要）
- 使用参数化查询（xb 默认支持）
- 实现访问控制（RBAC）

## 📚 参考资源

- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling)
- [JSON Schema 规范](https://json-schema.org/)
- [OpenAPI 3.0 规范](https://swagger.io/specification/)

---

**提示**: 结合 [RAG_BEST_PRACTICES.md](./RAG_BEST_PRACTICES.md) 了解如何构建完整的 RAG 应用。

