# sqlxb 使用场景决策指南

**根据数据特征选择合适的技术方案**

---

## 🎯 场景 1️⃣：语义搜索、个性化推荐

### 使用向量数据库（pgvector / Qdrant）

**适用业务**：
- ✅ 商品推荐（"买了 A 的用户还喜欢..."）
- ✅ 代码搜索（"查找相似的函数实现"）
- ✅ 客户服务（"查找类似的历史工单"）
- ✅ 内容推荐（"相似的文章、视频"）
- ✅ 图片搜索（"找相似的图片"）

**数据特征**：
- 数据碎片化（每条独立）
- 需要相似度匹配
- 无明确结构

**sqlxb 示例**：
```go
sqlxb.Of(&Product{}).
    VectorSearch("embedding", userVector, 20).
    Eq("category", "electronics")
```

**完整示例**：
- [PostgreSQL + pgvector 应用](../examples/pgvector-app/)
- [Qdrant 集成应用](../examples/qdrant-app/)

---

## 🎯 场景 2️⃣：结构化长文档分析

### 使用 PageIndex

**适用业务**：
- ✅ 金融年报分析（"2024 年财务稳定性如何？"）
- ✅ 法律合同检索（"第 3 章违约责任条款"）
- ✅ 技术手册查询（"安装步骤在第几页？"）
- ✅ 学术论文阅读（"方法论部分的内容"）
- ✅ 政策文件分析（"第 2.3 节的具体规定"）

**数据特征**：
- 长文档（50+ 页）
- 有明确的章节结构
- 需要保留上下文

**sqlxb 示例**：
```go
sqlxb.Of(&PageIndexNode{}).
    Eq("doc_id", docID).
    Like("title", "财务稳定").
    Eq("level", 1)
```

**完整示例**：
- [PageIndex 应用](../examples/pageindex-app/)

---

## 🎯 场景 3️⃣：混合检索（结构 + 语义）

### 使用 PageIndex + 向量数据库

**适用业务**：
- ✅ 研报智能问答（"科技板块的投资建议"）
- ✅ 知识库检索（既要结构，又要语义）
- ✅ 医学文献分析（"治疗方案相关章节"）
- ✅ 专利检索（"技术方案相似的专利"）

**数据特征**：
- 既有结构，又需语义
- 长文档 + 精确匹配需求

**sqlxb 示例**：
```go
// 第一步：PageIndex 定位章节
sqlxb.Of(&PageIndexNode{}).
    Like("title", "投资建议").
    Eq("level", 2)

// 第二步：在章节内向量检索
sqlxb.Of(&DocumentChunk{}).
    VectorSearch("embedding", queryVector, 10).
    Gte("page", chapterStartPage).
    Lte("page", chapterEndPage)
```

**完整示例**：
- [RAG 应用](../examples/rag-app/)（混合检索架构）

---

## 🎯 场景 4️⃣：传统业务数据

### 使用标准 SQL（无需向量/PageIndex）

**适用业务**：
- ✅ 用户管理（"查找 18 岁以上用户"）
- ✅ 订单查询（"2024 年 1 月的订单"）
- ✅ 库存管理（"库存不足的商品"）
- ✅ 统计报表（"按地区统计销售额"）

**数据特征**：
- 结构化数据
- 精确条件匹配
- 无需语义理解

**sqlxb 示例**：
```go
sqlxb.Of(&User{}).
    Gte("age", 18).
    Eq("status", "active").
    Paged(...)
```

---

## 🤔 快速决策树

```
你的数据是...

├─ 碎片化（商品、用户、代码片段）
│  └─ 需要"相似"匹配？
│     ├─ 是 → 向量数据库 ✅
│     └─ 否 → 标准 SQL ✅
│
└─ 长文档（报告、手册、合同）
   └─ 有明确章节结构？
      ├─ 是 → PageIndex ✅
      │  └─ 还需要语义匹配？
      │     └─ 是 → PageIndex + 向量 ✅
      └─ 否 → 传统 RAG（分块 + 向量）✅
```

---

## 💡 核心原则

**不要纠结技术选型，看数据特征：**

### 1️⃣ 数据碎片化 + 需要相似度
→ **向量数据库**

### 2️⃣ 长文档 + 有结构 + 需要章节定位
→ **PageIndex**

### 3️⃣ 长文档 + 无结构 + 需要语义
→ **传统 RAG**（分块 + 向量）

### 4️⃣ 结构化数据 + 精确匹配
→ **标准 SQL**

### 5️⃣ 复杂场景
→ **混合使用**

---

## 📊 对比表

| 维度 | 向量数据库 | PageIndex | 传统 RAG | 标准 SQL |
|------|-----------|-----------|---------|---------|
| **数据类型** | 碎片化 | 长文档 | 长文档 | 结构化 |
| **需要语义** | ✅ | ⚠️ 可选 | ✅ | ❌ |
| **需要结构** | ❌ | ✅ | ❌ | ✅ |
| **典型文档** | 商品、代码 | 报告、合同 | 文章、博客 | 用户、订单 |
| **查询方式** | 相似度 | 章节定位 | 相似度 | 精确匹配 |
| **响应时间** | 快（<50ms） | 中（100-500ms） | 快（<100ms） | 极快（<10ms） |
| **准确率** | 80-90% | 95%+ | 70-85% | 100% |

---

## 🚀 实战案例

### 案例 1：电商系统

```
需求：
  - 商品推荐 → 向量数据库
  - 订单查询 → 标准 SQL
  - 用户管理 → 标准 SQL

方案：
  sqlxb.Of(&Product{}).VectorSearch(...)  // 推荐
  sqlxb.Of(&Order{}).Eq("user_id", ...)   // 查询
```

---

### 案例 2：金融机构

```
需求：
  - 年报分析（200+ 页）→ PageIndex
  - 相似企业查找 → 向量数据库
  - 客户信息 → 标准 SQL

方案：
  sqlxb.Of(&PageIndexNode{}).Like("title", ...)  // 年报
  sqlxb.Of(&Company{}).VectorSearch(...)         // 相似企业
  sqlxb.Of(&Customer{}).Eq("region", ...)        // 客户
```

---

### 案例 3：技术文档平台

```
需求：
  - API 手册（500+ 页）→ PageIndex
  - 代码搜索 → 向量数据库
  - 博客文章 → 传统 RAG

方案：
  sqlxb.Of(&PageIndexNode{}).Eq("level", 2)     // 手册章节
  sqlxb.Of(&CodeSnippet{}).VectorSearch(...)    // 代码
  sqlxb.Of(&ArticleChunk{}).VectorSearch(...)   // 博客
```

---

## ✅ 如何选择？

### 问自己 3 个问题

```
1. 数据是碎片化还是长文档？
   - 碎片化 → 向量 or SQL
   - 长文档 → PageIndex or RAG

2. 需要"相似"匹配还是"精确"匹配？
   - 相似 → 向量 or RAG
   - 精确 → SQL or PageIndex

3. 文档有没有明确的章节结构？
   - 有 → PageIndex
   - 无 → RAG
```

---

## 📚 相关文档

- [English Version (英文版)](../README.md#-use-case-decision-guide)
- [Complete Examples](../examples/README.md)
- [PageIndex Integration Guide](./PAGEINDEX_INTEGRATION.md)
- [RAG Best Practices](./ai_application/RAG_BEST_PRACTICES.md)

---

**sqlxb 支持所有场景 —— 一套 API，全部搞定！** ✅

---

**最后更新**: 2025-02-27  
**版本**: v0.10.4

