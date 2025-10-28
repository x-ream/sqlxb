# LangChain 集成指南

## 📋 概述

本文档介绍如何将 sqlxb 与 Python LangChain 框架集成，构建强大的 RAG 应用。

## 🏗️ 架构设计

### 集成方式

```
┌──────────────────────────────────────────┐
│         Python LangChain 应用              │
│  (Chains, Agents, Memory)                │
└──────────────────────────────────────────┘
                    ↓ HTTP/gRPC
┌──────────────────────────────────────────┐
│         Go Backend (sqlxb)                │
│  • VectorSearch API                       │
│  • Hybrid Search API                      │
│  • Document Management API                │
└──────────────────────────────────────────┘
                    ↓
┌──────────────────────────────────────────┐
│      Qdrant / PostgreSQL+pgvector         │
└──────────────────────────────────────────┘
```

## 🚀 快速开始

### 1. Go Backend 服务

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/x-ream/xb"
)

type SearchRequest struct {
    Query         string                 `json:"query"`
    Embedding     []float32              `json:"embedding"`
    Filters       map[string]interface{} `json:"filters"`
    TopK          int                    `json:"top_k"`
    ScoreThreshold float64               `json:"score_threshold"`
}

type SearchResponse struct {
    Results []SearchResult `json:"results"`
}

type SearchResult struct {
    ID       int64                  `json:"id"`
    Content  string                 `json:"content"`
    Metadata map[string]interface{} `json:"metadata"`
    Score    float64                `json:"score"`
}

func handleVectorSearch(w http.ResponseWriter, r *http.Request) {
    var req SearchRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 构建查询
    builder := sqlxb.Of(&DocumentChunk{}).
        VectorSearch("embedding", req.Embedding)
    
    // 添加过滤条件
    if docType, ok := req.Filters["doc_type"].(string); ok {
        builder.Eq("doc_type", docType)
    }
    if lang, ok := req.Filters["language"].(string); ok {
        builder.Eq("language", lang)
    }
    
    // 生成 Qdrant 查询
    built := builder.
        QdrantX(func(qx *sqlxb.QdrantBuilderX) {
            qx.ScoreThreshold(float32(req.ScoreThreshold))
        }).
        Build()

    qdrantJSON, err := built.ToQdrantJSON()
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 执行查询（假设已有 qdrantClient）
    results, err := qdrantClient.Search(r.Context(), qdrantQuery)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 转换结果
    response := SearchResponse{
        Results: convertResults(results),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    http.HandleFunc("/api/vector-search", handleVectorSearch)
    http.HandleFunc("/api/documents", handleDocumentCRUD)
    http.ListenAndServe(":8080", nil)
}
```

### 2. Python LangChain 客户端

```python
from langchain.vectorstores.base import VectorStore
from langchain.embeddings.base import Embeddings
from typing import List, Tuple, Optional, Dict, Any
import requests

class SqlxbVectorStore(VectorStore):
    """sqlxb 向量存储适配器"""
    
    def __init__(
        self,
        backend_url: str,
        embedding: Embeddings,
        collection_name: str = "default"
    ):
        self.backend_url = backend_url
        self.embedding = embedding
        self.collection_name = collection_name
    
    def add_texts(
        self,
        texts: List[str],
        metadatas: Optional[List[dict]] = None,
        **kwargs: Any
    ) -> List[str]:
        """添加文档"""
        embeddings = self.embedding.embed_documents(texts)
        
        documents = []
        for i, (text, emb) in enumerate(zip(texts, embeddings)):
            doc = {
                "content": text,
                "embedding": emb,
                "metadata": metadatas[i] if metadatas else {}
            }
            documents.append(doc)
        
        response = requests.post(
            f"{self.backend_url}/api/documents",
            json={"documents": documents, "collection": self.collection_name}
        )
        response.raise_for_status()
        
        return [str(doc["id"]) for doc in response.json()["created"]]
    
    def similarity_search_with_score(
        self,
        query: str,
        k: int = 4,
        filter: Optional[Dict[str, Any]] = None,
        **kwargs: Any
    ) -> List[Tuple[Document, float]]:
        """相似度搜索（带分数）"""
        # 生成查询向量
        query_embedding = self.embedding.embed_query(query)
        
        # 调用 sqlxb backend
        response = requests.post(
            f"{self.backend_url}/api/vector-search",
            json={
                "query": query,
                "embedding": query_embedding,
                "filters": filter or {},
                "top_k": k,
                "score_threshold": kwargs.get("score_threshold", 0.0)
            }
        )
        response.raise_for_status()
        
        results = response.json()["results"]
        
        # 转换为 LangChain Document 格式
        docs_and_scores = []
        for result in results:
            doc = Document(
                page_content=result["content"],
                metadata=result["metadata"]
            )
            docs_and_scores.append((doc, result["score"]))
        
        return docs_and_scores
    
    def similarity_search(
        self,
        query: str,
        k: int = 4,
        **kwargs: Any
    ) -> List[Document]:
        """相似度搜索（不带分数）"""
        docs_and_scores = self.similarity_search_with_score(query, k, **kwargs)
        return [doc for doc, _ in docs_and_scores]
    
    @classmethod
    def from_texts(
        cls,
        texts: List[str],
        embedding: Embeddings,
        metadatas: Optional[List[dict]] = None,
        backend_url: str = "http://localhost:8080",
        **kwargs: Any
    ) -> "SqlxbVectorStore":
        """从文本创建向量存储"""
        store = cls(backend_url, embedding)
        store.add_texts(texts, metadatas, **kwargs)
        return store
```

### 3. 基础 RAG 应用

```python
from langchain.embeddings import OpenAIEmbeddings
from langchain.chat_models import ChatOpenAI
from langchain.chains import RetrievalQA
from langchain.document_loaders import TextLoader
from langchain.text_splitter import RecursiveCharacterTextSplitter

# 1. 初始化组件
embeddings = OpenAIEmbeddings()
llm = ChatOpenAI(model="gpt-4", temperature=0)

# 2. 创建向量存储
vector_store = SqlxbVectorStore(
    backend_url="http://localhost:8080",
    embedding=embeddings,
    collection_name="my_docs"
)

# 3. 加载并索引文档
loader = TextLoader("docs/knowledge.txt")
documents = loader.load()

text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=500,
    chunk_overlap=50
)
texts = text_splitter.split_documents(documents)

# 添加元数据
metadatas = [
    {
        "source": doc.metadata.get("source", ""),
        "doc_type": "tutorial",
        "language": "zh"
    }
    for doc in texts
]

vector_store.add_texts(
    texts=[doc.page_content for doc in texts],
    metadatas=metadatas
)

# 4. 创建检索链
qa_chain = RetrievalQA.from_chain_type(
    llm=llm,
    chain_type="stuff",
    retriever=vector_store.as_retriever(
        search_kwargs={
            "k": 5,
            "filter": {"language": "zh", "doc_type": "tutorial"}
        }
    ),
    return_source_documents=True
)

# 5. 查询
result = qa_chain({"query": "如何使用 sqlxb 构建向量查询？"})

print(f"回答: {result['result']}")
print(f"\n来源文档:")
for doc in result['source_documents']:
    print(f"  - {doc.metadata['source']}")
```

## 🎯 高级用法

### 混合检索（Hybrid Search）

```python
class SqlxbHybridRetriever(BaseRetriever):
    """支持标量过滤的混合检索器"""
    
    def __init__(
        self,
        vector_store: SqlxbVectorStore,
        base_filters: Optional[Dict[str, Any]] = None,
        score_threshold: float = 0.7
    ):
        self.vector_store = vector_store
        self.base_filters = base_filters or {}
        self.score_threshold = score_threshold
    
    def get_relevant_documents(self, query: str) -> List[Document]:
        """检索相关文档"""
        # 从查询中提取结构化过滤条件
        filters = self._extract_filters(query)
        filters.update(self.base_filters)
        
        # 执行混合检索
        docs_and_scores = self.vector_store.similarity_search_with_score(
            query=query,
            k=20,  # 粗召回
            filter=filters,
            score_threshold=self.score_threshold
        )
        
        # 过滤低分结果
        filtered_docs = [
            doc for doc, score in docs_and_scores
            if score >= self.score_threshold
        ]
        
        return filtered_docs[:5]  # 返回 top-5
    
    def _extract_filters(self, query: str) -> Dict[str, Any]:
        """从查询中提取过滤条件（简化版）"""
        filters = {}
        
        # 语言检测
        if contains_chinese(query):
            filters["language"] = "zh"
        else:
            filters["language"] = "en"
        
        # 文档类型识别
        if "教程" in query or "tutorial" in query.lower():
            filters["doc_type"] = "tutorial"
        elif "API" in query.upper():
            filters["doc_type"] = "api"
        
        return filters

# 使用示例
hybrid_retriever = SqlxbHybridRetriever(
    vector_store=vector_store,
    base_filters={"status": "published"},
    score_threshold=0.75
)

qa_chain = RetrievalQA.from_chain_type(
    llm=llm,
    retriever=hybrid_retriever
)

result = qa_chain({"query": "最近更新的 API 文档"})
```

### 多查询检索（Multi-Query）

```python
from langchain.retrievers.multi_query import MultiQueryRetriever

# 自动生成多个查询变体
multi_query_retriever = MultiQueryRetriever.from_llm(
    retriever=vector_store.as_retriever(),
    llm=llm
)

# 单次查询会自动生成多个变体并合并结果
docs = multi_query_retriever.get_relevant_documents(
    "sqlxb 如何处理向量查询？"
)
# 内部可能生成:
# - "sqlxb vector search usage"
# - "how to use sqlxb for vector queries"
# - "sqlxb vector query examples"
```

### 上下文压缩（Contextual Compression）

```python
from langchain.retrievers import ContextualCompressionRetriever
from langchain.retrievers.document_compressors import LLMChainExtractor

# 创建压缩器
compressor = LLMChainExtractor.from_llm(llm)

# 包装检索器
compression_retriever = ContextualCompressionRetriever(
    base_compressor=compressor,
    base_retriever=vector_store.as_retriever(search_kwargs={"k": 10})
)

# 检索时自动压缩文档，只保留相关部分
compressed_docs = compression_retriever.get_relevant_documents(
    "sqlxb 的核心特性是什么？"
)
```

### 自查询检索（Self-Querying）

```python
from langchain.chains.query_constructor.base import AttributeInfo
from langchain.retrievers.self_query.base import SelfQueryRetriever

# 定义元数据字段信息
metadata_field_info = [
    AttributeInfo(
        name="doc_type",
        description="文档类型: tutorial, api, blog, faq",
        type="string"
    ),
    AttributeInfo(
        name="language",
        description="文档语言: zh, en",
        type="string"
    ),
    AttributeInfo(
        name="created_at",
        description="创建时间，格式为 YYYY-MM-DD",
        type="string"
    ),
    AttributeInfo(
        name="author",
        description="作者名称",
        type="string"
    ),
]

document_content_description = "sqlxb 库的技术文档和教程"

# 创建自查询检索器
self_query_retriever = SelfQueryRetriever.from_llm(
    llm=llm,
    vectorstore=vector_store,
    document_contents=document_content_description,
    metadata_field_info=metadata_field_info,
    verbose=True
)

# 自然语言查询会自动提取过滤条件
docs = self_query_retriever.get_relevant_documents(
    "查找2024年写的关于 API 的中文教程"
)
# 自动提取过滤条件:
# {
#   "doc_type": "tutorial",
#   "language": "zh",
#   "created_at": {"$gte": "2024-01-01"}
# }
```

## 🤖 Agent 集成

### 将 sqlxb 作为 Agent 工具

```python
from langchain.agents import Tool, AgentType, initialize_agent
from langchain.memory import ConversationBufferMemory

# 定义工具
search_tool = Tool(
    name="KnowledgeBaseSearch",
    func=lambda q: vector_store.similarity_search(q, k=3),
    description="""
    用于搜索 sqlxb 技术文档和教程。
    输入应该是一个清晰的问题或关键词。
    返回最相关的文档片段。
    """
)

# 创建 Agent
memory = ConversationBufferMemory(memory_key="chat_history", return_messages=True)

agent = initialize_agent(
    tools=[search_tool],
    llm=llm,
    agent=AgentType.CONVERSATIONAL_REACT_DESCRIPTION,
    memory=memory,
    verbose=True
)

# 对话式查询
response = agent.run("sqlxb 支持哪些数据库？")
print(response)

response = agent.run("那 Qdrant 的集成怎么用？")  # 基于历史上下文
print(response)
```

### 多工具 Agent

```python
from langchain.tools import StructuredTool

# 定义多个工具
search_docs_tool = StructuredTool.from_function(
    func=lambda query, doc_type: vector_store.similarity_search(
        query,
        k=5,
        filter={"doc_type": doc_type}
    ),
    name="SearchDocs",
    description="搜索特定类型的文档。参数: query (str), doc_type (str: tutorial|api|blog|faq)"
)

search_code_tool = StructuredTool.from_function(
    func=lambda query: vector_store.similarity_search(
        query,
        k=3,
        filter={"doc_type": "code_example"}
    ),
    name="SearchCodeExamples",
    description="搜索代码示例。参数: query (str)"
)

# 创建多工具 Agent
agent = initialize_agent(
    tools=[search_docs_tool, search_code_tool],
    llm=llm,
    agent=AgentType.OPENAI_FUNCTIONS,
    verbose=True
)

result = agent.run("我想看看如何使用 sqlxb 进行向量检索的代码示例")
```

## 📊 完整应用示例

### 文档问答系统

```python
import os
from langchain.chains import ConversationalRetrievalChain
from langchain.memory import ConversationBufferMemory

class DocQASystem:
    def __init__(self, backend_url: str, openai_api_key: str):
        self.embeddings = OpenAIEmbeddings(openai_api_key=openai_api_key)
        self.llm = ChatOpenAI(model="gpt-4", temperature=0, openai_api_key=openai_api_key)
        
        self.vector_store = SqlxbVectorStore(
            backend_url=backend_url,
            embedding=self.embeddings
        )
        
        self.memory = ConversationBufferMemory(
            memory_key="chat_history",
            return_messages=True,
            output_key="answer"
        )
        
        self.qa_chain = ConversationalRetrievalChain.from_llm(
            llm=self.llm,
            retriever=self.vector_store.as_retriever(
                search_kwargs={"k": 5, "score_threshold": 0.7}
            ),
            memory=self.memory,
            return_source_documents=True,
            verbose=True
        )
    
    def index_directory(self, directory: str):
        """索引目录中的所有文档"""
        from langchain.document_loaders import DirectoryLoader, TextLoader
        
        loader = DirectoryLoader(
            directory,
            glob="**/*.md",
            loader_cls=TextLoader
        )
        
        documents = loader.load()
        
        text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=500,
            chunk_overlap=50,
            separators=["\n\n", "\n", " ", ""]
        )
        
        splits = text_splitter.split_documents(documents)
        
        # 添加元数据
        for doc in splits:
            doc.metadata.update({
                "language": "zh" if contains_chinese(doc.page_content) else "en",
                "doc_type": self._infer_doc_type(doc),
                "indexed_at": datetime.now().isoformat()
            })
        
        self.vector_store.add_documents(splits)
        
        return len(splits)
    
    def query(self, question: str, filters: Optional[Dict] = None) -> Dict:
        """查询文档"""
        if filters:
            # 临时更新检索器的过滤条件
            self.qa_chain.retriever.search_kwargs["filter"] = filters
        
        result = self.qa_chain({"question": question})
        
        return {
            "answer": result["answer"],
            "sources": [
                {
                    "content": doc.page_content[:200] + "...",
                    "metadata": doc.metadata
                }
                for doc in result["source_documents"]
            ]
        }
    
    def _infer_doc_type(self, doc) -> str:
        """推断文档类型"""
        filename = doc.metadata.get("source", "").lower()
        
        if "tutorial" in filename or "guide" in filename:
            return "tutorial"
        elif "api" in filename:
            return "api"
        elif "blog" in filename:
            return "blog"
        elif "faq" in filename:
            return "faq"
        else:
            return "general"

# 使用示例
if __name__ == "__main__":
    qa_system = DocQASystem(
        backend_url="http://localhost:8080",
        openai_api_key=os.getenv("OPENAI_API_KEY")
    )
    
    # 索引文档
    print("正在索引文档...")
    num_chunks = qa_system.index_directory("./docs")
    print(f"已索引 {num_chunks} 个文档块")
    
    # 交互式问答
    print("\n文档问答系统已就绪。输入 'quit' 退出。\n")
    
    while True:
        question = input("问题: ")
        if question.lower() in ["quit", "exit"]:
            break
        
        result = qa_system.query(
            question,
            filters={"doc_type": "tutorial"}  # 只搜索教程
        )
        
        print(f"\n回答: {result['answer']}\n")
        print("来源:")
        for i, source in enumerate(result['sources'], 1):
            print(f"  [{i}] {source['metadata']['source']}")
        print()
```

## 🔧 性能优化

### 1. 批量Embedding

```python
# 批量生成 embedding 提高效率
texts = [doc.page_content for doc in documents]

# 每次处理 100 个
batch_size = 100
all_embeddings = []

for i in range(0, len(texts), batch_size):
    batch = texts[i:i+batch_size]
    embeddings = embeddings_model.embed_documents(batch)
    all_embeddings.extend(embeddings)

# 批量插入
vector_store.add_texts_with_embeddings(texts, all_embeddings, metadatas)
```

### 2. 异步处理

```python
import asyncio
from langchain.embeddings import OpenAIEmbeddings

class AsyncSqlxbVectorStore(SqlxbVectorStore):
    async def aadd_texts(
        self,
        texts: List[str],
        metadatas: Optional[List[dict]] = None,
        **kwargs
    ) -> List[str]:
        """异步添加文档"""
        embeddings = await self.embedding.aembed_documents(texts)
        
        # ... 异步 HTTP 请求
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.backend_url}/api/documents",
                json={"documents": documents}
            ) as response:
                result = await response.json()
                return [str(doc["id"]) for doc in result["created"]]

# 使用异步版本
async def index_documents_async(docs):
    tasks = [
        vector_store.aadd_texts([doc.page_content], [doc.metadata])
        for doc in docs
    ]
    await asyncio.gather(*tasks)

asyncio.run(index_documents_async(documents))
```

## 📚 完整项目模板

查看 `examples/langchain-rag-app/` 目录获取完整的项目模板，包括:

- ✅ Go Backend API (使用 sqlxb)
- ✅ Python LangChain 客户端
- ✅ FastAPI REST API
- ✅ Streamlit Web UI
- ✅ Docker Compose 部署配置
- ✅ 完整测试套件

## 🤝 社区资源

- [LangChain 官方文档](https://python.langchain.com/)
- [sqlxb 示例仓库](https://github.com/x-ream/xb-examples)
- [常见问题解答](./FAQ.md)

---

**下一步**: 查看 [LLAMAINDEX_INTEGRATION.md](./LLAMAINDEX_INTEGRATION.md) 了解 LlamaIndex 集成。

