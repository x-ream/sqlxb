# PageIndex 文档结构化检索应用

这是一个结合 **Vectify AI PageIndex** 和 **xb** 的完整应用，展示如何存储和查询文档的层级结构。

## 📋 什么是 PageIndex？

**PageIndex** 是 Vectify AI 开发的基于推理的 RAG 框架，它：
- 解析 PDF 文档为层级结构树
- 使用 LLM 提取逻辑结构（章节、小节）
- 模拟人类专家查阅报告的方式
- 比传统分块检索更准确

## 🏗️ 架构设计

```
PageIndex (Python) → 生成 JSON 结构
                     ↓
              PostgreSQL (存储)
                     ↓
              xb (查询)
                     ↓
              应用层 (推理定位)
```

## 📊 数据模型

### 扁平化存储（推荐）

将 PageIndex 的层级 JSON 拆分为关系表，便于查询：

```sql
CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(500),
    total_pages INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE page_index_nodes (
    id BIGSERIAL PRIMARY KEY,
    doc_id BIGINT REFERENCES documents(id),
    node_id VARCHAR(50),
    parent_id VARCHAR(50),
    title TEXT,
    start_page INT,
    end_page INT,
    summary TEXT,
    level INT,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ON page_index_nodes (doc_id, node_id);
CREATE INDEX ON page_index_nodes (doc_id, parent_id);
CREATE INDEX ON page_index_nodes (doc_id, level);
CREATE INDEX ON page_index_nodes (start_page, end_page);
```

## 🚀 快速开始

### 1. 安装 PageIndex（Python 环境）

```bash
git clone https://github.com/VectifyAI/PageIndex.git
cd PageIndex
pip3 install -r requirements.txt

# 配置 OpenAI API Key
echo "CHATGPT_API_KEY=your-api-key" > .env
```

### 2. 处理 PDF 文档

```bash
python3 run_pageindex.py --pdf_path /path/to/report.pdf
# 输出：report_structure.json
```

### 3. 启动 Go 应用

```bash
cd examples/pageindex-app
go mod tidy
go run *.go
```

### 4. 导入 PageIndex 结果

```bash
# 导入生成的 JSON 结构
curl -X POST http://localhost:8080/api/import \
  -H "Content-Type: application/json" \
  -d @report_structure.json
```

### 5. 查询文档

```bash
# 按标题搜索
curl "http://localhost:8080/api/search/title?doc_id=1&keyword=Financial"

# 按页码查询
curl "http://localhost:8080/api/search/page?doc_id=1&page=25"

# 按层级查询
curl "http://localhost:8080/api/search/level?doc_id=1&level=2"

# 查询子节点
curl "http://localhost:8080/api/nodes/0006/children"
```

## 📁 项目结构

```
pageindex-app/
├── README.md
├── main.go              # 主程序
├── model.go             # 数据模型
├── repository.go        # 数据访问层
├── handler.go           # HTTP 处理器
├── importer.go          # PageIndex JSON 导入器
├── repository_test.go   # 测试
└── go.mod
```

## 💡 核心特性

### 1. 层级查询

```go
// 查询文档的第一层节点（章节）
nodes := FindTopLevelNodes(docID)

// 查询特定节点的子节点
children := FindChildNodes(nodeID)

// 递归查询所有后代
descendants := FindDescendants(nodeID)
```

### 2. 页码定位

```go
// 查询包含第 25 页的所有节点
nodes := FindNodesByPage(docID, 25)

// 查询页码范围
nodes := FindNodesByPageRange(docID, 20, 30)
```

### 3. 标题搜索

```go
// 模糊搜索标题
nodes := SearchNodesByTitle(docID, "Financial Stability")
```

## 🎯 与传统 RAG 的区别

| 特性 | 传统 RAG | PageIndex + xb |
|------|---------|------------------|
| 分块策略 | 固定大小 | 逻辑结构 |
| 检索方式 | 向量相似度 | 结构推理 + 查询 |
| 上下文理解 | 弱 | 强（保留层级） |
| 查询工具 | 向量数据库 | xb + PostgreSQL |
| 适用场景 | 通用文档 | 结构化报告、书籍 |

## 📚 相关文档

- [PageIndex GitHub](https://github.com/VectifyAI/PageIndex)
- [xb Builder Best Practices](../../doc/BUILDER_BEST_PRACTICES.md)
- [RAG Best Practices](../../doc/ai_application/RAG_BEST_PRACTICES.md)

---

**版本**: v0.10.4  
**最后更新**: 2025-02-27

