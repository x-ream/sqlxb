# PageIndex + xb 集成指南

## 📋 概述

本文档介绍如何将 **Vectify AI PageIndex** 与 **sqlxb** 结合，构建结构化文档检索系统。

---

## 🎯 PageIndex 是什么？

### 技术定位

```
PageIndex ≠ 数据库产品
PageIndex = 文档结构化处理框架（Python）

开发者: Vectify AI
GitHub: https://github.com/VectifyAI/PageIndex
```

### 核心功能

```
1. PDF 文档解析
   PDF → OCR/文本提取
   
2. 层级结构提取（使用 LLM）
   文本 → 章节、小节、段落
   
3. JSON 结构输出
   层级树 → 结构化 JSON
   
4. 模拟专家查阅
   保留文档逻辑结构
   准确定位相关内容
```

---

## 🏗️ 架构设计

### 传统 RAG vs PageIndex

```
传统 RAG:
  文档 → 固定分块 → 向量化 → 相似度检索
  问题：丢失文档结构

PageIndex:
  文档 → 层级结构 → 存储 → 结构化查询 + LLM推理
  优势：保留逻辑结构
```

### 完整流程

```
第一步（Python）：
  PDF → PageIndex 处理 → JSON 结构

第二步（Golang + sqlxb）：
  JSON → 扁平化 → PostgreSQL 存储

第三步（查询）：
  用户查询 → xb 查询 → 返回相关节点
  
第四步（应用层）：
  节点 → LLM 推理 → 精确内容定位
```

---

## 💾 数据存储设计

### 方案 1：JSONB 存储（简单）

```sql
CREATE TABLE page_index_docs (
    id BIGSERIAL PRIMARY KEY,
    doc_name VARCHAR(500),
    structure JSONB,  -- 整个 PageIndex JSON
    created_at TIMESTAMP
);
```

**优点**：
- 保持原始层级结构
- 导入简单

**缺点**：
- 查询不便
- 需要应用层遍历

---

### 方案 2：扁平化存储（推荐） ✅

```sql
CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(500),
    total_pages INT,
    created_at TIMESTAMP
);

CREATE TABLE page_index_nodes (
    id BIGSERIAL PRIMARY KEY,
    doc_id BIGINT REFERENCES documents(id),
    node_id VARCHAR(50),     -- "0006"
    parent_id VARCHAR(50),   -- "0005"
    title TEXT,              -- "Financial Stability"
    start_page INT,          -- 21
    end_page INT,            -- 28
    summary TEXT,
    level INT,               -- 层级深度
    created_at TIMESTAMP
);

CREATE INDEX ON page_index_nodes (doc_id, node_id);
CREATE INDEX ON page_index_nodes (doc_id, parent_id);
CREATE INDEX ON page_index_nodes (start_page, end_page);
```

**优点**：
- ✅ 使用 xb 高效查询
- ✅ 支持复杂过滤条件
- ✅ 性能更好

**缺点**：
- 需要递归导入

---

## 🔧 使用 xb 查询

### 1. 按标题搜索

```go
// 模糊搜索标题
func SearchByTitle(docID int64, keyword string) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Like("title", keyword).  // ✅ 自动添加 %
        Sort("level", xb.ASC).
        Sort("start_page", xb.ASC).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

// SQL: SELECT * FROM page_index_nodes 
//      WHERE doc_id = ? AND title LIKE ? 
//      ORDER BY level ASC, start_page ASC
```

---

### 2. 按页码定位

```go
// 查询包含第 25 页的所有节点
func FindByPage(docID int64, page int) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Lte("start_page", page).  // ✅ start_page <= 25
        Gte("end_page", page).    // ✅ end_page >= 25
        Sort("level", xb.ASC).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

// SQL: SELECT * FROM page_index_nodes 
//      WHERE doc_id = ? AND start_page <= ? AND end_page >= ?
//      ORDER BY level ASC
```

**为什么用 `Lte` 和 `Gte`？**
```go
// ✅ 更好：使用 xb API
Lte("start_page", page).Gte("end_page", page)

// ❌ 不好：手写 SQL
X("start_page <= ? AND end_page >= ?", page, page)

优势：
  - 类型安全
  - 自动过滤（page 为 0 时忽略）
  - 更清晰
```

---

### 3. 按层级查询

```go
// 查询第一层节点（章节）
func FindTopLevel(docID int64) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Eq("level", 1).  // 第一层
        Sort("start_page", xb.ASC).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

// 查询特定层级范围
func FindByLevelRange(docID int64, minLevel, maxLevel int) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Gte("level", minLevel).
        Lte("level", maxLevel).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}
```

---

### 4. 层级遍历

```go
// 查询子节点
func FindChildren(docID int64, parentNodeID string) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Eq("parent_id", parentNodeID).
        Sort("start_page", xb.ASC).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

// 递归查询所有后代
func FindDescendants(docID int64, nodeID string) ([]*PageIndexNode, error) {
    // 使用 PostgreSQL 递归 CTE
    sql := `
        WITH RECURSIVE descendants AS (
            SELECT * FROM page_index_nodes WHERE doc_id = $1 AND node_id = $2
            UNION ALL
            SELECT n.* FROM page_index_nodes n
            INNER JOIN descendants d ON n.parent_id = d.node_id AND n.doc_id = d.doc_id
        )
        SELECT * FROM descendants ORDER BY level, start_page
    `
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, docID, nodeID)
    return nodes, err
}
```

---

## 💡 高级查询

### 组合查询

```go
// 在特定章节中搜索关键词
func SearchInChapter(docID int64, chapterNodeID, keyword string) ([]*PageIndexNode, error) {
    // 1. 先找到章节节点
    chapter, _ := FindNodeByID(docID, chapterNodeID)
    
    // 2. 在该章节的页码范围内搜索
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Like("title", keyword).
        Gte("start_page", chapter.StartPage).
        Lte("end_page", chapter.EndPage).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

// 查询跨多个层级的节点
func FindCrossLevel(docID int64, keyword string, levels []int) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Like("title", keyword).
        In("level", toInterfaces(levels)...).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

func toInterfaces(ints []int) []interface{} {
    result := make([]interface{}, len(ints))
    for i, v := range ints {
        result[i] = v
    }
    return result
}
```

---

## 🎯 与 LLM 集成

### 推理式查询流程

```go
// 第一步：使用 LLM 分析查询，确定相关层级
func AnalyzeQuery(question string) ([]string, error) {
    prompt := fmt.Sprintf(`
文档结构：
- Level 1: 章节
- Level 2: 小节
- Level 3: 段落

问题：%s

请分析：这个问题最可能在哪个层级找到答案？返回 node_id 列表。
`, question)
    
    // 调用 LLM
    relevantNodeIDs := llm.Call(prompt)
    return relevantNodeIDs, nil
}

// 第二步：使用 xb 查询相关节点
func RetrieveRelevantNodes(docID int64, nodeIDs []string) ([]*PageIndexNode, error) {
    sql, args, _ := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        In("node_id", toInterfaces(nodeIDs)...).
        Build().
        SqlOfSelect()
    
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}

// 第三步：递归展开子节点（如果需要）
func ExpandNodes(docID int64, nodeIDs []string) ([]*PageIndexNode, error) {
    allNodes := []*PageIndexNode{}
    
    for _, nodeID := range nodeIDs {
        // 获取节点本身
        node, _ := FindNodeByID(docID, nodeID)
        allNodes = append(allNodes, node)
        
        // 获取所有后代
        descendants, _ := FindDescendants(docID, nodeID)
        allNodes = append(allNodes, descendants...)
    }
    
    return allNodes, nil
}
```

---

## 📊 性能优化

### 索引策略

```sql
-- 复合索引：文档 + 节点
CREATE INDEX idx_doc_node ON page_index_nodes (doc_id, node_id);

-- 复合索引：文档 + 父节点
CREATE INDEX idx_doc_parent ON page_index_nodes (doc_id, parent_id);

-- 复合索引：文档 + 页码范围
CREATE INDEX idx_page_range ON page_index_nodes (doc_id, start_page, end_page);

-- 全文索引：标题搜索
CREATE INDEX idx_title_fts ON page_index_nodes USING gin (to_tsvector('english', title));
```

### 查询优化

```go
// ✅ 好：使用索引
builder.Eq("doc_id", docID).
        Eq("level", 1)

// ❌ 不好：全表扫描
builder.Like("summary", keyword)  // 如果 summary 很长且没有索引
```

---

## 🔄 数据导入流程

### 完整示例

```go
// 1. PageIndex 处理文档（Python）
// $ python3 run_pageindex.py --pdf_path report.pdf
// 输出：report_structure.json

// 2. 解析 JSON
jsonData, _ := ioutil.ReadFile("report_structure.json")
var pageIndexResult PageIndexJSON
json.Unmarshal(jsonData, &pageIndexResult)

// 3. 创建文档记录
doc := &Document{
    Name:       "Annual Report 2024",
    TotalPages: 100,
}
repo.CreateDocument(doc)

// 4. 递归导入节点
func importNode(docID int64, node PageIndexJSON, parentID string, level int) {
    // 创建当前节点
    dbNode := &PageIndexNode{
        DocID:     docID,
        NodeID:    node.NodeID,
        ParentID:  parentID,
        Title:     node.Title,
        StartPage: node.StartIndex,
        EndPage:   node.EndIndex,
        Summary:   node.Summary,
        Level:     level,
    }
    repo.CreateNode(dbNode)
    
    // 递归处理子节点
    for _, child := range node.Nodes {
        importNode(docID, child, node.NodeID, level+1)
    }
}
```

---

## 📝 最佳实践

### 1. 查询优化

```go
// ✅ 充分利用 xb 的自动过滤
func SearchNodes(docID int64, params SearchParams) ([]*PageIndexNode, error) {
    builder := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Like("title", params.Keyword).       // 空字符串自动忽略
        Gte("level", params.MinLevel).       // 0 自动忽略
        Lte("level", params.MaxLevel).       // 0 自动忽略
        Gte("start_page", params.MinPage).   // 0 自动忽略
        Lte("end_page", params.MaxPage)      // 0 自动忽略
    
    // 不需要手动检查 nil/0！
    
    sql, args, _ := builder.Build().SqlOfSelect()
    var nodes []*PageIndexNode
    err := db.Select(&nodes, sql, args...)
    return nodes, err
}
```

---

### 2. 分页查询

```go
// 分页查询节点
func PagedNodes(docID int64, level, page, rows int) ([]*PageIndexNode, int64, error) {
    builder := xb.Of(&PageIndexNode{}).
        Eq("doc_id", docID).
        Eq("level", level).
        Paged(func(pb *xb.PageBuilder) {
            pb.Page(int64(page)).Rows(int64(rows))
        })
    
    countSql, dataSql, args, _ := builder.Build().SqlOfPage()
    
    // 获取总数
    var total int64
    if countSql != "" {
        db.Get(&total, countSql)
    }
    
    // 获取数据
    var nodes []*PageIndexNode
    err := db.Select(&nodes, dataSql, args...)
    
    return nodes, total, err
}
```

---

## 🎯 应用场景

### 1. 金融报告分析

```
文档：年度财务报告（100+ 页）
PageIndex 识别：
  - Chapter 1: Executive Summary
  - Chapter 2: Financial Stability
    - 2.1 Monitoring
    - 2.2 Cooperation
  - Chapter 3: Risk Management

查询："2024 年财务稳定性如何？"
  → sqlxb: 查找 title 包含 "Financial Stability"
  → 返回 Chapter 2 及其子节点
  → LLM: 基于这些节点内容回答
```

---

### 2. 技术文档检索

```
文档：技术手册（500+ 页）
PageIndex 识别：
  - Part 1: Installation
    - 1.1 Requirements
    - 1.2 Setup
  - Part 2: API Reference
    - 2.1 REST API
    - 2.2 GraphQL

查询："如何安装？"
  → sqlxb: 查找 level=1, title 包含 "Installation"
  → 返回 Part 1 及所有子节点
  → LLM: 提取具体安装步骤
```

---

## 🚀 完整示例

详见 [examples/pageindex-app](../examples/pageindex-app/)

---

## 📚 相关资源

### PageIndex

- [GitHub Repository](https://github.com/VectifyAI/PageIndex)
- [技术博客]（待更新）

### sqlxb

- [Builder Best Practices](./BUILDER_BEST_PRACTICES.md)
- [RAG Best Practices](./ai_application/RAG_BEST_PRACTICES.md)
- [Complete Examples](../examples/README.md)

---

**最后更新**: 2025-02-27  
**版本**: v0.10.4

