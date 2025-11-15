// Copyright 2025 me.fndo.xb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xb

import (
	"encoding/json"
	"fmt"
)

// ============================================================================
// QdrantBuilder: Builder 模式配置构建器
// ============================================================================

// QdrantBuilder Qdrant 配置构建器
// 使用 Builder 模式构建 QdrantCustom 配置
type QdrantBuilder struct {
	custom *QdrantCustom
}

// NewQdrantBuilder 创建 Qdrant 配置构建器
//
// 示例:
//
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.8).
//	        Build(),
//	).Build()
func NewQdrantBuilder() *QdrantBuilder {
	return &QdrantBuilder{
		custom: NewQdrantCustom(),
	}
}

// HnswEf 设置 HNSW 算法的 ef 参数
// ef 越大，查询精度越高，但速度越慢
// 推荐值: 64-256
func (qb *QdrantBuilder) HnswEf(ef int) *QdrantBuilder {
	if ef < 1 {
		panic(fmt.Sprintf("HnswEf must be >= 1, got: %d", ef))
	}
	qb.custom.DefaultHnswEf = ef
	return qb
}

// ScoreThreshold 设置最小相似度阈值
// 只返回相似度 >= threshold 的结果
func (qb *QdrantBuilder) ScoreThreshold(threshold float32) *QdrantBuilder {
	if threshold < 0 || threshold > 1 {
		panic(fmt.Sprintf("ScoreThreshold must be in [0, 1], got: %f", threshold))
	}
	qb.custom.DefaultScoreThreshold = threshold
	return qb
}

// WithVector 设置是否返回向量数据
// true: 返回向量（占用带宽）
// false: 不返回向量（节省带宽）
func (qb *QdrantBuilder) WithVector(withVector bool) *QdrantBuilder {
	qb.custom.DefaultWithVector = withVector
	return qb
}

// Build 构建并返回 QdrantCustom 配置
func (qb *QdrantBuilder) Build() *QdrantCustom {
	return qb.custom
}

// ============================================================================
// QdrantCustom: Qdrant 数据库专属配置
// ============================================================================

// QdrantCustom Qdrant 专属配置实现
//
// 实现 Custom 接口，提供 Qdrant 的默认配置和预设模式
type QdrantCustom struct {
	// 默认参数（如果用户没有显式指定，使用这些默认值）
	DefaultHnswEf         int     // 默认 HNSW EF 参数
	DefaultScoreThreshold float32 // 默认相似度阈值
	DefaultWithVector     bool    // 默认是否返回向量

	// 高级 API 配置（Recommend / Discover / Scroll）
	recommendConfig *qdrantRecommendConfig
	discoverConfig  *qdrantDiscoverConfig
	scrollID        string
}

// NewQdrantCustom 创建 Qdrant Custom（默认配置）
//
// ⚠️ 设计原则：只提供这一个构造函数！
//
// 不要添加预设函数（如 QdrantHighPrecision/HighSpeed/Balanced）：
//   - 原因：增加概念负担，影响 Go 生态简洁性
//   - 替代：使用 NewQdrantBuilder() 链式构建
//
// 如果你想添加预设函数，请先问：
//  1. 用户不用这个函数能实现吗？（答案：能，使用 Builder 即可）
//  2. 这会增加概念数量吗？（答案：会）
//  3. 那为什么要加？（答案：...不应该加）
//
// 参考：xb v1.1.0 的教训（5 个预设函数 → v1.2.0 全部删除）
//
// 正确的用户配置方式：
//
// 方式 1: 使用 QdrantBuilder（推荐）
//
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.8).
//	        Build(),
//	).Build()
//
// 方式 2: 直接设置字段
//
//	custom := NewQdrantCustom()
//	custom.DefaultHnswEf = 512
//	xb.Of(...).Custom(custom).Build()
func NewQdrantCustom() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:         128,
		DefaultScoreThreshold: 0.0,
		DefaultWithVector:     false, // 向后兼容：默认不返回向量
	}
}

// Recommend 启用 Qdrant Recommend API
//
// 示例:
//
//	custom := xb.NewQdrantCustom().Recommend(func(rb *xb.RecommendBuilder) {
//		rb.Positive(123, 456).Negative(789).Limit(20)
//	})
func (c *QdrantCustom) Recommend(fn func(rb *RecommendBuilder)) *QdrantCustom {
	if fn == nil {
		c.recommendConfig = nil
		return c
	}

	builder := &RecommendBuilder{}
	fn(builder)

	if len(builder.positive) == 0 {
		panic("Recommend() requires at least one Positive() id")
	}
	if builder.limit <= 0 {
		panic("Recommend() requires Limit() > 0")
	}

	c.recommendConfig = &qdrantRecommendConfig{
		positive: append([]int64(nil), builder.positive...),
		negative: append([]int64(nil), builder.negative...),
		limit:    builder.limit,
	}
	return c
}

// Discover 启用 Qdrant Discover API
//
// 示例:
//
//	custom := xb.NewQdrantCustom().Discover(func(db *xb.DiscoverBuilder) {
//		db.Context(101, 102, 103).Limit(20)
//	})
func (c *QdrantCustom) Discover(fn func(db *DiscoverBuilder)) *QdrantCustom {
	if fn == nil {
		c.discoverConfig = nil
		return c
	}

	builder := &DiscoverBuilder{}
	fn(builder)

	if len(builder.context) == 0 {
		panic("Discover() requires Context() with at least one id")
	}
	if builder.limit <= 0 {
		panic("Discover() requires Limit() > 0")
	}

	c.discoverConfig = &qdrantDiscoverConfig{
		context: append([]int64(nil), builder.context...),
		limit:   builder.limit,
	}
	return c
}

// ScrollID 启用 Qdrant Scroll API
//
// 示例:
//
//	custom := xb.NewQdrantCustom().ScrollID("scroll-abc123")
func (c *QdrantCustom) ScrollID(scrollID string) *QdrantCustom {
	if scrollID == "" {
		panic("ScrollID() requires a non-empty id")
	}
	c.scrollID = scrollID
	return c
}

// Generate 实现 Custom 接口
// ⭐ 根据操作类型返回不同的 JSON
func (c *QdrantCustom) Generate(built *Built) (interface{}, error) {
	built = c.applyAdvancedConfig(built)

	// ⭐ INSERT: 生成 Qdrant upsert JSON
	if built.Inserts != nil && len(*built.Inserts) > 0 {
		return c.generateInsertJSON(built)
	}

	// ⭐ UPDATE: 生成 Qdrant update payload JSON
	if built.Updates != nil && len(*built.Updates) > 0 {
		return c.generateUpdateJSON(built)
	}

	// ⭐ DELETE: 生成 Qdrant delete JSON
	if built.Delete {
		return c.generateDeleteJSON(built)
	}

	// ⭐ SELECT: 生成 Qdrant search JSON
	switch {
	case hasBbWithOp(built.Conds, QDRANT_RECOMMEND):
		return built.toQdrantRecommendJSON()
	case hasBbWithOp(built.Conds, QDRANT_DISCOVER):
		return built.toQdrantDiscoverJSON()
	case hasBbWithOp(built.Conds, QDRANT_SCROLL):
		return built.toQdrantScrollJSON()
	default:
		json, err := built.toQdrantJSON()
		return json, err
	}
}

// ============================================================================
// 使用说明
// ============================================================================
//
// 配置方式：
//
// 方式 1: 使用 QdrantBuilder（推荐）
//
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.85).
//	        Build(),
//	).Build()
//
// 方式 2: 手动设置字段
//
//	custom := NewQdrantCustom()
//	custom.DefaultHnswEf = 512
//	custom.DefaultScoreThreshold = 0.85
//	xb.Of(...).Custom(custom).Build()
//
// 方式 3: 配置复用
//
//	highPrecision := xb.NewQdrantBuilder().HnswEf(512).Build()
//	xb.Of(...).Custom(highPrecision).Build()
//	xb.Of(...).Custom(highPrecision).Build()  // 复用配置
//

// ============================================================================
// Insert/Update/Delete JSON 生成
// ============================================================================

// QdrantPoint Qdrant 点结构
type QdrantPoint struct {
	ID      interface{}            `json:"id"`
	Vector  interface{}            `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// generateInsertJSON 生成 Qdrant upsert JSON
// PUT /collections/{collection_name}/points
func (c *QdrantCustom) generateInsertJSON(built *Built) (string, error) {
	inserts := *built.Inserts
	if len(inserts) == 0 {
		return "", fmt.Errorf("no insert data")
	}

	// Qdrant upsert 请求结构
	type QdrantUpsertRequest struct {
		Points []QdrantPoint `json:"points"`
	}

	points := []QdrantPoint{}

	// ⭐ 使用 Insert(func(ib)) 格式
	// 多个 bb（字段-值对）组成一个 point
	point, err := c.extractPointFromBbs(inserts)
	if err != nil {
		return "", err
	}
	points = append(points, point)

	req := QdrantUpsertRequest{Points: points}
	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant upsert request: %w", err)
	}

	return string(bytes), nil
}

// generateUpdateJSON 生成 Qdrant update payload JSON
// POST /collections/{collection_name}/points/payload
func (c *QdrantCustom) generateUpdateJSON(built *Built) (string, error) {
	updates := *built.Updates
	if len(updates) == 0 {
		return "", fmt.Errorf("no update data")
	}

	// Qdrant update payload 请求结构
	type QdrantUpdateRequest struct {
		Points  []interface{}          `json:"points,omitempty"` // 点 ID 列表
		Filter  *QdrantFilter          `json:"filter,omitempty"` // 或使用过滤器
		Payload map[string]interface{} `json:"payload"`          // 要更新的 payload
	}

	// 提取 payload
	payload := make(map[string]interface{})
	for _, bb := range updates {
		payload[bb.Key] = bb.Value
	}

	req := QdrantUpdateRequest{
		Payload: payload,
	}

	// 从条件中提取 point IDs 或构建 filter
	ids, filter := c.extractIdsOrFilter(built.Conds)
	if len(ids) > 0 {
		req.Points = ids
	} else if filter != nil {
		req.Filter = filter
	}

	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant update request: %w", err)
	}

	return string(bytes), nil
}

// generateDeleteJSON 生成 Qdrant delete JSON
// POST /collections/{collection_name}/points/delete
func (c *QdrantCustom) generateDeleteJSON(built *Built) (string, error) {
	// Qdrant delete 请求结构
	type QdrantDeleteRequest struct {
		Points []interface{} `json:"points,omitempty"` // 点 ID 列表
		Filter *QdrantFilter `json:"filter,omitempty"` // 或使用过滤器
	}

	req := QdrantDeleteRequest{}

	// 从条件中提取 point IDs 或构建 filter
	ids, filter := c.extractIdsOrFilter(built.Conds)
	if len(ids) > 0 {
		req.Points = ids
	} else if filter != nil {
		req.Filter = filter
	} else {
		return "", fmt.Errorf("no delete conditions (points or filter)")
	}

	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant delete request: %w", err)
	}

	return string(bytes), nil
}

// ============================================================================
// 辅助方法
// ============================================================================

// extractPointFromBbs 从 InsertBuilder 的 bbs 中提取 Qdrant Point
// 用于 Insert(func(ib *InsertBuilder)) 格式
func (c *QdrantCustom) extractPointFromBbs(bbs []Bb) (QdrantPoint, error) {
	point := QdrantPoint{
		Payload: make(map[string]interface{}),
	}

	for _, bb := range bbs {
		switch bb.Key {
		case "id":
			point.ID = bb.Value
		case "vector":
			point.Vector = bb.Value
		default:
			// 其他字段放入 payload
			point.Payload[bb.Key] = bb.Value
		}
	}

	// 验证必需字段
	if point.ID == nil {
		return QdrantPoint{}, fmt.Errorf("point must have 'id' field")
	}
	if point.Vector == nil {
		return QdrantPoint{}, fmt.Errorf("point must have 'vector' field")
	}

	return point, nil
}

// extractIdsOrFilter 从条件中提取 point IDs 或构建 filter
func (c *QdrantCustom) extractIdsOrFilter(conds []Bb) ([]interface{}, *QdrantFilter) {
	ids := []interface{}{}

	// 查找 id IN (...) 条件
	for _, bb := range conds {
		if bb.Key == "id" {
			if bb.Op == IN {
				// IN 条件：提取 ID 列表
				if arr, ok := bb.Value.(*[]string); ok {
					for _, id := range *arr {
						ids = append(ids, id)
					}
					return ids, nil
				}
			} else if bb.Op == EQ {
				// 单个 ID
				ids = append(ids, bb.Value)
				return ids, nil
			}
		}
	}

	// 如果没有 id 条件，构建 filter
	if len(conds) > 0 {
		filter := &QdrantFilter{}
		filter.Must = []QdrantCondition{}

		for _, bb := range conds {
			cond := QdrantCondition{
				Key: bb.Key,
			}

			switch bb.Op {
			case EQ:
				cond.Match = &QdrantMatchCondition{Value: bb.Value}
			case GT, GTE, LT, LTE:
				cond.Range = &QdrantRangeCondition{}
				switch bb.Op {
				case GT:
					cond.Range.Gt = toFloat64Ptr(bb.Value)
				case GTE:
					cond.Range.Gte = toFloat64Ptr(bb.Value)
				case LT:
					cond.Range.Lt = toFloat64Ptr(bb.Value)
				case LTE:
					cond.Range.Lte = toFloat64Ptr(bb.Value)
				}
			}

			filter.Must = append(filter.Must, cond)
		}

		return nil, filter
	}

	return nil, nil
}

func toFloat64Ptr(v interface{}) *float64 {
	switch val := v.(type) {
	case float64:
		return &val
	case float32:
		f := float64(val)
		return &f
	case int:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	}
	return nil
}

// qdrantRecommendConfig Recommend API 配置
type qdrantRecommendConfig struct {
	positive []int64
	negative []int64
	limit    int
}

// qdrantDiscoverConfig Discover API 配置
type qdrantDiscoverConfig struct {
	context []int64
	limit   int
}

// RecommendBuilder Recommend API 构建器
type RecommendBuilder struct {
	positive []int64
	negative []int64
	limit    int
}

// Positive 设置正样本 ID
func (rb *RecommendBuilder) Positive(ids ...int64) *RecommendBuilder {
	if len(ids) == 0 {
		return rb
	}
	rb.positive = append(rb.positive, ids...)
	return rb
}

// Negative 设置负样本 ID
func (rb *RecommendBuilder) Negative(ids ...int64) *RecommendBuilder {
	if len(ids) == 0 {
		return rb
	}
	rb.negative = append(rb.negative, ids...)
	return rb
}

// Limit 设置返回条数
func (rb *RecommendBuilder) Limit(limit int) *RecommendBuilder {
	rb.limit = limit
	return rb
}

// DiscoverBuilder Discover API 构建器
type DiscoverBuilder struct {
	context []int64
	limit   int
}

// Context 设置上下文 ID
func (db *DiscoverBuilder) Context(ids ...int64) *DiscoverBuilder {
	if len(ids) == 0 {
		return db
	}
	db.context = append(db.context, ids...)
	return db
}

// Limit 设置返回条数
func (db *DiscoverBuilder) Limit(limit int) *DiscoverBuilder {
	db.limit = limit
	return db
}

// ensureAdvancedConds 将高级配置注入到条件列表
func (c *QdrantCustom) ensureAdvancedConds(conds []Bb) []Bb {
	if c == nil {
		return conds
	}

	if c.recommendConfig != nil && !hasBbWithOp(conds, QDRANT_RECOMMEND) {
		value := map[string]interface{}{
			"positive": append([]int64(nil), c.recommendConfig.positive...),
			"limit":    c.recommendConfig.limit,
		}
		if len(c.recommendConfig.negative) > 0 {
			value["negative"] = append([]int64(nil), c.recommendConfig.negative...)
		}
		conds = append(conds, Bb{
			Op:    QDRANT_RECOMMEND,
			Value: value,
		})
	}

	if c.discoverConfig != nil && !hasBbWithOp(conds, QDRANT_DISCOVER) {
		value := map[string]interface{}{
			"context": append([]int64(nil), c.discoverConfig.context...),
			"limit":   c.discoverConfig.limit,
		}
		conds = append(conds, Bb{
			Op:    QDRANT_DISCOVER,
			Value: value,
		})
	}

	if c.scrollID != "" && !hasBbWithOp(conds, QDRANT_SCROLL) {
		conds = append(conds, Bb{
			Op:    QDRANT_SCROLL,
			Value: c.scrollID,
		})
	}

	return conds
}

func hasBbWithOp(bbs []Bb, op string) bool {
	for _, bb := range bbs {
		if bb.Op == op {
			return true
		}
	}
	return false
}

func (c *QdrantCustom) applyAdvancedConfig(built *Built) *Built {
	if c == nil || built == nil {
		return built
	}

	origLen := len(built.Conds)
	condsCopy := cloneBbs(built.Conds)
	newConds := c.ensureAdvancedConds(condsCopy)
	if len(newConds) == origLen {
		return built
	}

	cloned := *built
	cloned.Conds = newConds
	return &cloned
}

func cloneBbs(bbs []Bb) []Bb {
	if len(bbs) == 0 {
		return nil
	}
	cloned := make([]Bb, len(bbs))
	copy(cloned, bbs)
	return cloned
}
