-- PageIndex 应用数据库架构

-- 1. 文档表
CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    total_pages INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE documents IS 'PageIndex 处理的文档';

-- 2. PageIndex 节点表（扁平化存储）
CREATE TABLE IF NOT EXISTS page_index_nodes (
    id BIGSERIAL PRIMARY KEY,
    doc_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    node_id VARCHAR(50) NOT NULL,
    parent_id VARCHAR(50),
    title TEXT NOT NULL,
    start_page INT NOT NULL,
    end_page INT NOT NULL,
    summary TEXT,
    content TEXT,
    level INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE page_index_nodes IS 'PageIndex 生成的文档层级节点（扁平化）';
COMMENT ON COLUMN page_index_nodes.node_id IS 'PageIndex 节点 ID（如 "0006"）';
COMMENT ON COLUMN page_index_nodes.parent_id IS '父节点 ID（根节点为空）';
COMMENT ON COLUMN page_index_nodes.level IS '层级深度（根节点为 0）';

-- 3. 索引
CREATE INDEX IF NOT EXISTS idx_nodes_doc_id ON page_index_nodes (doc_id);
CREATE INDEX IF NOT EXISTS idx_nodes_node_id ON page_index_nodes (doc_id, node_id);
CREATE INDEX IF NOT EXISTS idx_nodes_parent_id ON page_index_nodes (doc_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_nodes_level ON page_index_nodes (doc_id, level);
CREATE INDEX IF NOT EXISTS idx_nodes_page_range ON page_index_nodes (doc_id, start_page, end_page);
CREATE INDEX IF NOT EXISTS idx_nodes_title ON page_index_nodes USING gin (to_tsvector('english', title));

-- 4. 示例查询
-- 查询文档的顶层节点（章节）
-- SELECT * FROM page_index_nodes WHERE doc_id = 1 AND level = 1 ORDER BY start_page;

-- 查询包含第 25 页的所有节点
-- SELECT * FROM page_index_nodes WHERE doc_id = 1 AND start_page <= 25 AND end_page >= 25 ORDER BY level;

-- 查询特定节点的子节点
-- SELECT * FROM page_index_nodes WHERE doc_id = 1 AND parent_id = '0006' ORDER BY start_page;

-- 模糊搜索标题
-- SELECT * FROM page_index_nodes WHERE doc_id = 1 AND title LIKE '%Financial%' ORDER BY level, start_page;

