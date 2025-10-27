# LlamaIndex 集成指南

## 📋 概述

本文档介绍如何将 sqlxb 与 Python LlamaIndex 框架集成，构建高性能的 RAG 和数据查询应用。

## 🚀 快速开始

### 自定义向量存储

```python
from llama_index.core.vector_stores import (
    VectorStore,
    VectorStoreQuery,
    VectorStoreQueryResult,
)
from llama_index.core.schema import NodeWithScore, TextNode
from typing import List, Optional, Any
import requests

class SqlxbVectorStore(VectorStore):
    """sqlxb 向量存储适配器"""
    
    def __init__(self, backend_url: str, collection_name: str = "default"):
        self.backend_url = backend_url
        self.collection_name = collection_name
    
    def add(self, nodes: List[TextNode], **add_kwargs: Any) -> List[str]:
        """添加节点"""
        documents = []
        for node in nodes:
            doc = {
                "content": node.get_content(),
                "embedding": node.get_embedding(),
                "metadata": node.metadata,
                "node_id": node.node_id,
            }
            documents.append(doc)
        
        response = requests.post(
            f"{self.backend_url}/api/documents",
            json={"documents": documents, "collection": self.collection_name}
        )
        response.raise_for_status()
        
        return [doc["id"] for doc in response.json()["created"]]
    
    def query(self, query: VectorStoreQuery, **kwargs: Any) -> VectorStoreQueryResult:
        """执行查询"""
        response = requests.post(
            f"{self.backend_url}/api/vector-search",
            json={
                "embedding": query.query_embedding,
                "filters": query.filters or {},
                "top_k": query.similarity_top_k,
                "score_threshold": kwargs.get("score_threshold", 0.0)
            }
        )
        response.raise_for_status()
        
        results = response.json()["results"]
        
        # 转换为 LlamaIndex 格式
        nodes = []
        similarities = []
        ids = []
        
        for result in results:
            node = TextNode(
                text=result["content"],
                metadata=result["metadata"],
                node_id=result["id"],
            )
            nodes.append(NodeWithScore(node=node, score=result["score"]))
            similarities.append(result["score"])
            ids.append(str(result["id"]))
        
        return VectorStoreQueryResult(
            nodes=nodes,
            similarities=similarities,
            ids=ids
        )
    
    def delete(self, ref_doc_id: str, **delete_kwargs: Any) -> None:
        """删除文档"""
        requests.delete(
            f"{self.backend_url}/api/documents/{ref_doc_id}"
        )
```

### 基础 RAG 应用

```python
from llama_index.core import VectorStoreIndex, ServiceContext, StorageContext
from llama_index.core import SimpleDirectoryReader
from llama_index.embeddings.openai import OpenAIEmbedding
from llama_index.llms.openai import OpenAI

# 初始化组件
embed_model = OpenAIEmbedding()
llm = OpenAI(model="gpt-4", temperature=0)

# 创建向量存储
vector_store = SqlxbVectorStore(
    backend_url="http://localhost:8080",
    collection_name="my_docs"
)

storage_context = StorageContext.from_defaults(vector_store=vector_store)
service_context = ServiceContext.from_defaults(
    embed_model=embed_model,
    llm=llm
)

# 加载文档
documents = SimpleDirectoryReader("./docs").load_data()

# 构建索引
index = VectorStoreIndex.from_documents(
    documents,
    storage_context=storage_context,
    service_context=service_context,
)

# 查询
query_engine = index.as_query_engine(similarity_top_k=5)
response = query_engine.query("如何使用 sqlxb 进行向量检索？")

print(response)
```

## 🎯 高级功能

### 混合检索

```python
from llama_index.core.retrievers import VectorIndexRetriever
from llama_index.core.query_engine import RetrieverQueryEngine

# 创建检索器，支持元数据过滤
retriever = VectorIndexRetriever(
    index=index,
    similarity_top_k=10,
    filters={
        "doc_type": "tutorial",
        "language": "zh"
    }
)

# 创建查询引擎
query_engine = RetrieverQueryEngine.from_args(
    retriever=retriever,
    service_context=service_context
)

response = query_engine.query("sqlxb 的核心特性")
```

### 子问题查询

```python
from llama_index.core.query_engine import SubQuestionQueryEngine
from llama_index.core.tools import QueryEngineTool, ToolMetadata

# 为不同文档类型创建独立索引
tutorial_index = VectorStoreIndex.from_documents(
    tutorial_docs, storage_context=storage_context
)
api_index = VectorStoreIndex.from_documents(
    api_docs, storage_context=storage_context
)

# 定义查询工具
query_engine_tools = [
    QueryEngineTool(
        query_engine=tutorial_index.as_query_engine(),
        metadata=ToolMetadata(
            name="tutorial_docs",
            description="包含 sqlxb 教程和使用指南"
        ),
    ),
    QueryEngineTool(
        query_engine=api_index.as_query_engine(),
        metadata=ToolMetadata(
            name="api_docs",
            description="包含 sqlxb API 参考文档"
        ),
    ),
]

# 创建子问题查询引擎
sub_question_engine = SubQuestionQueryEngine.from_defaults(
    query_engine_tools=query_engine_tools,
    service_context=service_context
)

response = sub_question_engine.query(
    "sqlxb 如何集成 Qdrant？有哪些 API 可以使用？"
)
```

### 聊天引擎

```python
from llama_index.core.chat_engine import ContextChatEngine
from llama_index.core.memory import ChatMemoryBuffer

memory = ChatMemoryBuffer.from_defaults(token_limit=3000)

chat_engine = ContextChatEngine.from_defaults(
    retriever=retriever,
    memory=memory,
    service_context=service_context,
)

# 多轮对话
response1 = chat_engine.chat("sqlxb 支持哪些数据库？")
print(response1)

response2 = chat_engine.chat("Qdrant 怎么集成？")  # 有上下文记忆
print(response2)
```

## 🤖 Agent 集成

```python
from llama_index.core.agent import ReActAgent
from llama_index.core.tools import FunctionTool

def search_docs(query: str, doc_type: str = "all") -> str:
    """搜索文档库"""
    filters = {} if doc_type == "all" else {"doc_type": doc_type}
    
    retriever = VectorIndexRetriever(
        index=index,
        similarity_top_k=3,
        filters=filters
    )
    
    nodes = retriever.retrieve(query)
    return "\n\n".join([node.get_content() for node in nodes])

# 定义工具
tools = [
    FunctionTool.from_defaults(fn=search_docs),
]

# 创建 Agent
agent = ReActAgent.from_tools(
    tools,
    llm=llm,
    verbose=True
)

# 使用 Agent
response = agent.chat("帮我找一下关于向量检索的教程")
print(response)
```

## 📊 完整应用示例

```python
class DocQASystem:
    def __init__(self, backend_url: str):
        self.vector_store = SqlxbVectorStore(backend_url=backend_url)
        self.embed_model = OpenAIEmbedding()
        self.llm = OpenAI(model="gpt-4")
        
        self.storage_context = StorageContext.from_defaults(
            vector_store=self.vector_store
        )
        self.service_context = ServiceContext.from_defaults(
            embed_model=self.embed_model,
            llm=self.llm
        )
        
        self.index = None
    
    def index_directory(self, directory: str):
        """索引目录"""
        documents = SimpleDirectoryReader(directory).load_data()
        
        self.index = VectorStoreIndex.from_documents(
            documents,
            storage_context=self.storage_context,
            service_context=self.service_context,
        )
        
        return len(documents)
    
    def query(self, question: str, filters: dict = None):
        """查询"""
        retriever = VectorIndexRetriever(
            index=self.index,
            similarity_top_k=5,
            filters=filters or {}
        )
        
        query_engine = RetrieverQueryEngine.from_args(
            retriever=retriever,
            service_context=self.service_context
        )
        
        return query_engine.query(question)

# 使用
qa_system = DocQASystem("http://localhost:8080")
qa_system.index_directory("./docs")
response = qa_system.query("如何使用 sqlxb？")
print(response)
```

## 🔧 性能优化

### 异步批量处理

```python
import asyncio

async def async_index_documents(documents: List):
    """异步索引文档"""
    tasks = []
    for doc in documents:
        task = index.ainsert(doc)
        tasks.append(task)
    
    await asyncio.gather(*tasks)

# 使用
asyncio.run(async_index_documents(documents))
```

### 流式响应

```python
# 流式查询响应
query_engine = index.as_query_engine(streaming=True)
streaming_response = query_engine.query("sqlxb 的特性")

for text in streaming_response.response_gen:
    print(text, end="", flush=True)
```

## 📚 参考资源

- [LlamaIndex 官方文档](https://docs.llamaindex.ai/)
- [sqlxb 示例项目](https://github.com/x-ream/sqlxb/tree/main/examples)

---

**提示**: 结合 [LANGCHAIN_INTEGRATION.md](./LANGCHAIN_INTEGRATION.md) 比较两个框架的差异。

