# Semantic Kernel 集成指南 (.NET)

## 📋 概述

本文档介绍如何将 sqlxb 与 Microsoft Semantic Kernel (.NET) 集成，构建企业级 AI 应用。

## 🚀 快速开始

### 1. Go Backend API

```go
// 与 LangChain 集成相同，提供 HTTP API
// 参考 LANGCHAIN_INTEGRATION.md 中的 Go Backend 服务
```

### 2. .NET 客户端

```csharp
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.Memory;
using System.Net.Http.Json;

public class SqlxbMemoryStore : IMemoryStore
{
    private readonly HttpClient _httpClient;
    private readonly string _collectionName;

    public SqlxbMemoryStore(string backendUrl, string collectionName = "default")
    {
        _httpClient = new HttpClient { BaseAddress = new Uri(backendUrl) };
        _collectionName = collectionName;
    }

    public async Task<string> UpsertAsync(
        string collection,
        MemoryRecord record,
        CancellationToken cancellationToken = default)
    {
        var document = new
        {
            content = record.Metadata.Text,
            embedding = record.Embedding.ToArray(),
            metadata = new
            {
                id = record.Metadata.Id,
                description = record.Metadata.Description,
                additionalMetadata = record.Metadata.AdditionalMetadata
            }
        };

        var response = await _httpClient.PostAsJsonAsync(
            "/api/documents",
            new { documents = new[] { document }, collection },
            cancellationToken
        );

        response.EnsureSuccessStatusCode();
        var result = await response.Content.ReadFromJsonAsync<InsertResponse>(cancellationToken);
        
        return result.Created[0].Id.ToString();
    }

    public async Task<IAsyncEnumerable<MemoryRecord>> GetNearestMatchesAsync(
        string collection,
        ReadOnlyMemory<float> embedding,
        int limit,
        double minRelevanceScore = 0.0,
        bool withEmbeddings = false,
        CancellationToken cancellationToken = default)
    {
        var request = new
        {
            embedding = embedding.ToArray(),
            filters = new { collection },
            top_k = limit,
            score_threshold = minRelevanceScore
        };

        var response = await _httpClient.PostAsJsonAsync(
            "/api/vector-search",
            request,
            cancellationToken
        );

        response.EnsureSuccessStatusCode();
        var result = await response.Content.ReadFromJsonAsync<SearchResponse>(cancellationToken);

        return ConvertToMemoryRecords(result.Results, withEmbeddings);
    }

    private async IAsyncEnumerable<MemoryRecord> ConvertToMemoryRecords(
        SearchResult[] results,
        bool withEmbeddings)
    {
        foreach (var result in results)
        {
            var metadata = new MemoryRecordMetadata(
                isReference: false,
                id: result.Metadata["id"].ToString(),
                text: result.Content,
                description: result.Metadata.GetValueOrDefault("description")?.ToString() ?? "",
                externalSourceName: result.Metadata.GetValueOrDefault("source")?.ToString() ?? "",
                additionalMetadata: result.Metadata.GetValueOrDefault("additional")?.ToString() ?? ""
            );

            var embedding = withEmbeddings
                ? new ReadOnlyMemory<float>(result.Embedding)
                : ReadOnlyMemory<float>.Empty;

            yield return new MemoryRecord(
                metadata,
                embedding,
                key: result.Id.ToString(),
                timestamp: DateTimeOffset.UtcNow
            );
        }
    }

    // 实现其他 IMemoryStore 接口方法...
}
```

### 3. 基础 RAG 应用

```csharp
using Microsoft.SemanticKernel;
using Microsoft.SemanticKernel.Connectors.OpenAI;
using Microsoft.SemanticKernel.Memory;

class Program
{
    static async Task Main(string[] args)
    {
        // 1. 初始化 Semantic Kernel
        var kernel = Kernel.CreateBuilder()
            .AddOpenAIChatCompletion("gpt-4", "your-api-key")
            .Build();

        // 2. 配置 sqlxb Memory Store
        var memoryStore = new SqlxbMemoryStore(
            backendUrl: "http://localhost:8080",
            collectionName: "my_docs"
        );

        var embeddingGenerator = new OpenAITextEmbeddingGenerationService(
            "text-embedding-ada-002",
            "your-api-key"
        );

        var memory = new SemanticTextMemory(memoryStore, embeddingGenerator);

        // 3. 添加文档到记忆
        await memory.SaveInformationAsync(
            collection: "docs",
            id: "doc1",
            text: "sqlxb 是一个 AI-First 的 ORM 库，支持向量数据库。",
            description: "sqlxb 介绍"
        );

        await memory.SaveInformationAsync(
            collection: "docs",
            id: "doc2",
            text: "sqlxb 支持 PostgreSQL 和 Qdrant 两种后端。",
            description: "支持的数据库"
        );

        // 4. 查询
        var query = "sqlxb 支持哪些数据库？";
        
        var results = memory.SearchAsync(
            collection: "docs",
            query: query,
            limit: 5,
            minRelevanceScore: 0.7
        );

        // 5. 构建上下文
        var context = new StringBuilder();
        await foreach (var result in results)
        {
            context.AppendLine($"[相关性: {result.Relevance:F2}]");
            context.AppendLine(result.Metadata.Text);
            context.AppendLine();
        }

        // 6. 生成回答
        var prompt = $@"
基于以下文档内容回答问题：

{context}

问题：{query}

回答：";

        var response = await kernel.InvokePromptAsync(prompt);
        Console.WriteLine(response);
    }
}
```

## 🎯 高级功能

### 插件集成

```csharp
using Microsoft.SemanticKernel;
using System.ComponentModel;

public class DocumentSearchPlugin
{
    private readonly ISemanticTextMemory _memory;

    public DocumentSearchPlugin(ISemanticTextMemory memory)
    {
        _memory = memory;
    }

    [KernelFunction, Description("搜索技术文档")]
    public async Task<string> SearchDocs(
        [Description("搜索查询")] string query,
        [Description("文档类型：tutorial, api, blog")] string docType = "all")
    {
        var results = _memory.SearchAsync(
            collection: "docs",
            query: query,
            limit: 5,
            minRelevanceScore: 0.7,
            filter: docType != "all" ? new { doc_type = docType } : null
        );

        var documents = new StringBuilder();
        await foreach (var result in results)
        {
            documents.AppendLine($"[{result.Relevance:F2}] {result.Metadata.Text}");
        }

        return documents.ToString();
    }

    [KernelFunction, Description("搜索代码示例")]
    public async Task<string> SearchCode(
        [Description("代码功能描述")] string description)
    {
        var results = _memory.SearchAsync(
            collection: "code",
            query: description,
            limit: 3,
            minRelevanceScore: 0.75
        );

        var examples = new List<string>();
        await foreach (var result in results)
        {
            examples.Add(result.Metadata.Text);
        }

        return string.Join("\n\n---\n\n", examples);
    }
}

// 使用插件
var kernel = Kernel.CreateBuilder()
    .AddOpenAIChatCompletion("gpt-4", apiKey)
    .Build();

var memory = new SemanticTextMemory(memoryStore, embeddingGenerator);
kernel.ImportPluginFromObject(new DocumentSearchPlugin(memory));

// 自动调用插件
var settings = new OpenAIPromptExecutionSettings
{
    ToolCallBehavior = ToolCallBehavior.AutoInvokeKernelFunctions
};

var result = await kernel.InvokePromptAsync(
    "帮我找一些关于向量检索的代码示例",
    new KernelArguments(settings)
);

Console.WriteLine(result);
```

### Planner 集成

```csharp
using Microsoft.SemanticKernel.Planning;

// 创建 Planner
var planner = new HandlebarsPlanner(new HandlebarsPlannerOptions
{
    AllowLoops = true
});

// 定义目标
var goal = "帮我了解 sqlxb 的向量检索功能，并给出代码示例";

// 生成计划
var plan = await planner.CreatePlanAsync(kernel, goal);

Console.WriteLine("执行计划:");
Console.WriteLine(plan);

// 执行计划
var result = await plan.InvokeAsync(kernel);
Console.WriteLine("\n结果:");
Console.WriteLine(result);
```

### 聊天历史管理

```csharp
using Microsoft.SemanticKernel.ChatCompletion;

public class RAGChatService
{
    private readonly Kernel _kernel;
    private readonly ISemanticTextMemory _memory;
    private readonly ChatHistory _chatHistory;

    public RAGChatService(Kernel kernel, ISemanticTextMemory memory)
    {
        _kernel = kernel;
        _memory = memory;
        _chatHistory = new ChatHistory();
        
        _chatHistory.AddSystemMessage(
            "你是一个技术文档助手，基于提供的文档内容回答问题。"
        );
    }

    public async Task<string> ChatAsync(string userMessage)
    {
        // 1. 检索相关文档
        var relevantDocs = new StringBuilder();
        await foreach (var doc in _memory.SearchAsync("docs", userMessage, limit: 5))
        {
            relevantDocs.AppendLine(doc.Metadata.Text);
        }

        // 2. 添加用户消息（带上下文）
        var messageWithContext = $@"
相关文档:
{relevantDocs}

用户问题: {userMessage}
";
        _chatHistory.AddUserMessage(messageWithContext);

        // 3. 获取 AI 回复
        var chatService = _kernel.GetRequiredService<IChatCompletionService>();
        var response = await chatService.GetChatMessageContentAsync(
            _chatHistory,
            new OpenAIPromptExecutionSettings { Temperature = 0.7 }
        );

        // 4. 添加到历史
        _chatHistory.AddAssistantMessage(response.Content);

        return response.Content;
    }
}

// 使用
var chatService = new RAGChatService(kernel, memory);

Console.WriteLine(await chatService.ChatAsync("sqlxb 是什么？"));
Console.WriteLine(await chatService.ChatAsync("它支持哪些数据库？"));  // 有上下文
```

## 🤖 企业应用示例

### 文档问答系统

```csharp
public class EnterpriseDocQA
{
    private readonly Kernel _kernel;
    private readonly SqlxbMemoryStore _memoryStore;
    private readonly ILogger _logger;

    public async Task<QAResponse> AskAsync(
        string question,
        string userId,
        QAOptions options)
    {
        try
        {
            // 1. 权限检查
            var userRoles = await GetUserRoles(userId);
            
            // 2. 检索（带权限过滤）
            var results = await _memoryStore.GetNearestMatchesAsync(
                collection: "enterprise_docs",
                embedding: await EmbedAsync(question),
                limit: options.TopK,
                minRelevanceScore: options.MinScore,
                filter: new { allowed_roles = userRoles }
            );

            // 3. 构建上下文
            var context = await BuildContextAsync(results);

            // 4. 生成回答
            var answer = await GenerateAnswerAsync(question, context);

            // 5. 记录审计日志
            await LogQueryAsync(userId, question, answer);

            return new QAResponse
            {
                Answer = answer,
                Sources = results.Select(r => r.Metadata.Id).ToList(),
                Confidence = CalculateConfidence(results)
            };
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "查询失败");
            throw;
        }
    }
}
```

### 多租户支持

```csharp
public class MultiTenantRAGService
{
    private readonly Dictionary<string, ISemanticTextMemory> _tenantMemories;

    public MultiTenantRAGService(string backendUrl)
    {
        _tenantMemories = new Dictionary<string, ISemanticTextMemory>();
    }

    public async Task<string> QueryForTenant(
        string tenantId,
        string query)
    {
        // 获取或创建租户专用 memory
        if (!_tenantMemories.ContainsKey(tenantId))
        {
            var store = new SqlxbMemoryStore(
                backendUrl: _backendUrl,
                collectionName: $"tenant_{tenantId}"
            );
            
            _tenantMemories[tenantId] = new SemanticTextMemory(
                store,
                _embeddingGenerator
            );
        }

        var memory = _tenantMemories[tenantId];
        
        // 租户隔离的查询
        var results = memory.SearchAsync(
            collection: $"tenant_{tenantId}",
            query: query,
            limit: 5
        );

        // ... 处理结果
    }
}
```

## 🔧 配置与优化

### 依赖注入

```csharp
// Program.cs (ASP.NET Core)
builder.Services.AddSingleton<SqlxbMemoryStore>(sp =>
    new SqlxbMemoryStore(
        backendUrl: builder.Configuration["SqlxbBackendUrl"],
        collectionName: "docs"
    )
);

builder.Services.AddSingleton<ISemanticTextMemory>(sp =>
{
    var store = sp.GetRequiredService<SqlxbMemoryStore>();
    var embedding = new OpenAITextEmbeddingGenerationService(
        "text-embedding-ada-002",
        builder.Configuration["OpenAI:ApiKey"]
    );
    return new SemanticTextMemory(store, embedding);
});

builder.Services.AddTransient<Kernel>(sp =>
{
    return Kernel.CreateBuilder()
        .AddOpenAIChatCompletion(
            "gpt-4",
            builder.Configuration["OpenAI:ApiKey"]
        )
        .Build();
});
```

### 配置文件

```json
{
  "SqlxbBackendUrl": "http://localhost:8080",
  "OpenAI": {
    "ApiKey": "your-api-key",
    "ChatModel": "gpt-4",
    "EmbeddingModel": "text-embedding-ada-002"
  },
  "RAG": {
    "TopK": 5,
    "MinRelevanceScore": 0.7,
    "CacheDuration": "00:15:00"
  }
}
```

## 📚 参考资源

- [Semantic Kernel 官方文档](https://learn.microsoft.com/en-us/semantic-kernel/)
- [sqlxb GitHub](https://github.com/x-ream/xb)

---

**相关文档**: [LANGCHAIN_INTEGRATION.md](./LANGCHAIN_INTEGRATION.md)

