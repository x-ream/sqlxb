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

// QdrantSearchRequest Qdrant 搜索请求结构
// 文档: https://qdrant.tech/documentation/concepts/search/
type QdrantSearchRequest struct {
	Vector         []float32           `json:"vector"`
	Limit          int                 `json:"limit"`
	Filter         *QdrantFilter       `json:"filter,omitempty"`
	WithPayload    interface{}         `json:"with_payload,omitempty"` // true, false, or []string
	WithVector     bool                `json:"with_vector,omitempty"`
	ScoreThreshold *float32            `json:"score_threshold,omitempty"`
	Offset         int                 `json:"offset,omitempty"`
	Params         *QdrantSearchParams `json:"params,omitempty"`
}

// QdrantFilter Qdrant 过滤器
type QdrantFilter struct {
	Must    []QdrantCondition `json:"must,omitempty"`
	Should  []QdrantCondition `json:"should,omitempty"`
	MustNot []QdrantCondition `json:"must_not,omitempty"`
}

// QdrantCondition Qdrant 条件
type QdrantCondition struct {
	Key   string                `json:"key,omitempty"`
	Match *QdrantMatchCondition `json:"match,omitempty"`
	Range *QdrantRangeCondition `json:"range,omitempty"`
}

// QdrantMatchCondition Qdrant 精确匹配条件
type QdrantMatchCondition struct {
	Value interface{}   `json:"value,omitempty"`
	Any   []interface{} `json:"any,omitempty"`
}

// QdrantRangeCondition Qdrant 范围条件
type QdrantRangeCondition struct {
	Gt  *float64 `json:"gt,omitempty"`
	Gte *float64 `json:"gte,omitempty"`
	Lt  *float64 `json:"lt,omitempty"`
	Lte *float64 `json:"lte,omitempty"`
}

// QdrantSearchParams Qdrant 搜索参数
type QdrantSearchParams struct {
	HnswEf      int  `json:"hnsw_ef,omitempty"`
	Exact       bool `json:"exact,omitempty"`
	IndexedOnly bool `json:"indexed_only,omitempty"`
}

// QdrantRecommendRequest Qdrant 推荐请求结构 (v0.10.0)
// 文档: https://qdrant.tech/documentation/concepts/explore/#recommendation-api
type QdrantRecommendRequest struct {
	Positive       []int64             `json:"positive"`           // 正样本 ID 列表
	Negative       []int64             `json:"negative,omitempty"` // 负样本 ID 列表（可选）
	Limit          int                 `json:"limit"`
	Filter         *QdrantFilter       `json:"filter,omitempty"`
	WithPayload    interface{}         `json:"with_payload,omitempty"` // true, false, or []string
	WithVector     bool                `json:"with_vector,omitempty"`
	ScoreThreshold *float32            `json:"score_threshold,omitempty"`
	Offset         int                 `json:"offset,omitempty"`
	Params         *QdrantSearchParams `json:"params,omitempty"`
	Strategy       string              `json:"strategy,omitempty"` // "average_vector" or "best_score"
}

// QdrantScrollRequest Qdrant Scroll 请求结构 (v0.10.0)
// 文档: https://qdrant.tech/documentation/concepts/points/#scroll-points
type QdrantScrollRequest struct {
	ScrollID    string        `json:"scroll_id,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	Filter      *QdrantFilter `json:"filter,omitempty"`
	WithPayload interface{}   `json:"with_payload,omitempty"`
	WithVector  bool          `json:"with_vector,omitempty"`
}

// QdrantDiscoverRequest Qdrant Discover 请求结构 (v0.10.0)
// 文档: https://qdrant.tech/documentation/concepts/explore/#discovery-api
type QdrantDiscoverRequest struct {
	Context        []int64             `json:"context"` // 上下文样本 ID 列表
	Limit          int                 `json:"limit"`
	Filter         *QdrantFilter       `json:"filter,omitempty"`
	WithPayload    interface{}         `json:"with_payload,omitempty"` // true, false, or []string
	WithVector     bool                `json:"with_vector,omitempty"`
	ScoreThreshold *float32            `json:"score_threshold,omitempty"`
	Offset         int                 `json:"offset,omitempty"`
	Params         *QdrantSearchParams `json:"params,omitempty"`
}

// ToQdrantJSON 转换为 Qdrant 搜索 JSON
// 返回: JSON 字符串, error
//
// 示例输出:
//
//	{
//	  "vector": [0.1, 0.2, 0.3],
//	  "limit": 20,
//	  "filter": {
//	    "must": [
//	      {"key": "language", "match": {"value": "golang"}}
//	    ]
//	  },
//	  "with_payload": true,
//	  "params": {"hnsw_ef": 128}
//	}
func (built *Built) ToQdrantJSON() (string, error) {
	req, err := built.ToQdrantRequest()
	if err != nil {
		return "", err
	}

	// ⭐ 检查是否有用户自定义参数（QDRANT_XX）
	customParams := extractQdrantCustomParams(built.Conds)

	if len(customParams) == 0 {
		// 无自定义参数，直接序列化
		bytes, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
		}
		return string(bytes), nil
	}

	// 有自定义参数，先序列化为 map，再添加自定义字段
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(bytes, &reqMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// ⭐ 添加用户自定义参数
	for k, v := range customParams {
		reqMap[k] = v
	}

	// 重新序列化
	finalBytes, err := json.MarshalIndent(reqMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return string(finalBytes), nil
}

// extractQdrantCustomParams 提取用户自定义 Qdrant 参数
func extractQdrantCustomParams(bbs []Bb) map[string]interface{} {
	params := make(map[string]interface{})
	for _, bb := range bbs {
		if bb.op == QDRANT_XX {
			params[bb.key] = bb.value
		}
	}
	return params
}

// ToQdrantRecommendJSON 转换为 Qdrant 推荐 JSON (v0.10.0)
// 返回: JSON 字符串, error
//
// 示例输出:
//
//	{
//	  "positive": [123, 456, 789],
//	  "negative": [111, 222],
//	  "limit": 20,
//	  "filter": {...},
//	  "strategy": "best_score"
//	}
func (built *Built) ToQdrantRecommendJSON() (string, error) {
	// 查找推荐参数
	recommendBb := findRecommendBb(built.Conds)
	if recommendBb == nil {
		return "", fmt.Errorf("no recommend configuration found")
	}

	recommendData := recommendBb.value.(map[string]interface{})

	// 构建推荐请求
	req := &QdrantRecommendRequest{
		Positive:    recommendData["positive"].([]int64),
		Negative:    []int64{},
		Limit:       recommendData["limit"].(int),
		WithPayload: true,
		WithVector:  false,
	}

	// 处理负样本（可选）
	if negative, ok := recommendData["negative"].([]int64); ok && len(negative) > 0 {
		req.Negative = negative
	}

	// 应用 Qdrant 专属参数
	applyQdrantParamsToRecommend(built.Conds, req)

	// 应用过滤器
	filter, err := buildQdrantFilter(built.Conds)
	if err == nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// 序列化
	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant recommend request: %w", err)
	}
	return string(bytes), nil
}

// ToQdrantScrollJSON 转换为 Qdrant Scroll JSON (v0.10.0)
// 返回: JSON 字符串, error
//
// 示例输出:
//
//	{
//	  "scroll_id": "xxxx-yyyy-zzzz",
//	  "limit": 100,
//	  "filter": {...}
//	}
func (built *Built) ToQdrantScrollJSON() (string, error) {
	// 查找 Scroll ID
	scrollBb := findScrollBb(built.Conds)
	if scrollBb == nil {
		return "", fmt.Errorf("no scroll_id found")
	}

	// 构建 Scroll 请求
	req := &QdrantScrollRequest{
		ScrollID:    scrollBb.value.(string),
		Limit:       100, // 默认值
		WithPayload: true,
		WithVector:  false,
	}

	// 应用过滤器
	filter, err := buildQdrantFilter(built.Conds)
	if err == nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// 应用 WithVector 参数
	for _, bb := range built.Conds {
		if bb.op == QDRANT_WITH_VECTOR {
			req.WithVector = bb.value.(bool)
			break
		}
	}

	// 序列化
	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant scroll request: %w", err)
	}
	return string(bytes), nil
}

// findRecommendBb 查找推荐配置
func findRecommendBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].op == QDRANT_RECOMMEND {
			return &bbs[i]
		}
	}
	return nil
}

// findScrollBb 查找 Scroll ID
func findScrollBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].op == QDRANT_SCROLL {
			return &bbs[i]
		}
	}
	return nil
}

// findDiscoverBb 查找 Discover 配置
func findDiscoverBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].op == QDRANT_DISCOVER {
			return &bbs[i]
		}
	}
	return nil
}

// applyQdrantParamsToRecommend 应用 Qdrant 专属参数到推荐请求
func applyQdrantParamsToRecommend(bbs []Bb, req *QdrantRecommendRequest) {
	for _, bb := range bbs {
		switch bb.op {
		case QDRANT_HNSW_EF:
			if req.Params == nil {
				req.Params = &QdrantSearchParams{}
			}
			req.Params.HnswEf = bb.value.(int)
		case QDRANT_EXACT:
			if req.Params == nil {
				req.Params = &QdrantSearchParams{}
			}
			req.Params.Exact = bb.value.(bool)
		case QDRANT_SCORE_THRESHOLD:
			threshold := bb.value.(float32)
			req.ScoreThreshold = &threshold
		case QDRANT_WITH_VECTOR:
			req.WithVector = bb.value.(bool)
		}
	}
}

// applyQdrantParamsToDiscover 应用 Qdrant 专属参数到探索请求
func applyQdrantParamsToDiscover(bbs []Bb, req *QdrantDiscoverRequest) {
	for _, bb := range bbs {
		switch bb.op {
		case QDRANT_HNSW_EF:
			if req.Params == nil {
				req.Params = &QdrantSearchParams{}
			}
			req.Params.HnswEf = bb.value.(int)
		case QDRANT_EXACT:
			if req.Params == nil {
				req.Params = &QdrantSearchParams{}
			}
			req.Params.Exact = bb.value.(bool)
		case QDRANT_SCORE_THRESHOLD:
			threshold := bb.value.(float32)
			req.ScoreThreshold = &threshold
		case QDRANT_WITH_VECTOR:
			req.WithVector = bb.value.(bool)
		}
	}
}

// ToQdrantDiscoverJSON 转换为 Qdrant Discover JSON (v0.10.0)
// 返回: JSON 字符串, error
//
// 示例输出:
//
//	{
//	  "context": [101, 102, 103],
//	  "limit": 20,
//	  "filter": {...}
//	}
func (built *Built) ToQdrantDiscoverJSON() (string, error) {
	// 查找探索配置
	discoverBb := findDiscoverBb(built.Conds)
	if discoverBb == nil {
		return "", fmt.Errorf("no discover configuration found")
	}

	discoverData := discoverBb.value.(map[string]interface{})

	// 构建探索请求
	req := &QdrantDiscoverRequest{
		Context:     discoverData["context"].([]int64),
		Limit:       discoverData["limit"].(int),
		WithPayload: true,
		WithVector:  false,
	}

	// 应用 Qdrant 专属参数
	applyQdrantParamsToDiscover(built.Conds, req)

	// 应用过滤器
	filter, err := buildQdrantFilter(built.Conds)
	if err == nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// 序列化
	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant discover request: %w", err)
	}
	return string(bytes), nil
}

// ToQdrantRequest 转换为 Qdrant 请求结构
func (built *Built) ToQdrantRequest() (*QdrantSearchRequest, error) {
	// 查找向量检索参数
	vectorBb := findVectorSearchBb(built.Conds)
	if vectorBb == nil {
		return nil, fmt.Errorf("no vector search found")
	}

	params := vectorBb.value.(VectorSearchParams)

	// 构建请求
	req := &QdrantSearchRequest{
		Vector:      params.QueryVector,
		Limit:       params.TopK,
		WithPayload: true,
		WithVector:  false,
	}

	// ⭐ 多样性处理：如果启用了多样性，需要过度获取
	if params.Diversity != nil && params.Diversity.Enabled {
		factor := params.Diversity.OverFetchFactor
		if factor <= 0 {
			factor = 5 // 默认 5 倍
		}
		req.Limit = params.TopK * factor

		// 注意：Qdrant 不原生支持多样性，需要在应用层处理
		// 这里只是获取更多结果，后续需要应用层过滤
	}

	// 构建过滤器
	filter, err := buildQdrantFilter(built.Conds)
	if err != nil {
		return nil, err
	}
	if filter != nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// 设置搜索参数
	req.Params = &QdrantSearchParams{
		HnswEf:      128,
		Exact:       false,
		IndexedOnly: false,
	}

	// ⭐ 应用 Qdrant 专属配置（从 Conds 中提取）
	applyQdrantSpecificConfig(built.Conds, req)

	// 应用分页配置（如果有）
	if built.PageCondition != nil {
		req.Limit = int(built.PageCondition.rows)
		if built.PageCondition.page > 1 {
			req.Offset = int((built.PageCondition.page - 1) * built.PageCondition.rows)
		}

		// 如果启用了多样性，需要覆盖 limit
		if params.Diversity != nil && params.Diversity.Enabled {
			factor := params.Diversity.OverFetchFactor
			if factor <= 0 {
				factor = 5
			}
			req.Limit = int(built.PageCondition.rows) * factor
		}
	}

	return req, nil
}

// applyQdrantSpecificConfig 从 Bb 中提取 Qdrant 专属配置
func applyQdrantSpecificConfig(bbs []Bb, req *QdrantSearchRequest) {
	for _, bb := range bbs {
		switch bb.op {
		case QDRANT_HNSW_EF:
			if ef, ok := bb.value.(int); ok {
				req.Params.HnswEf = ef
			}
		case QDRANT_EXACT:
			if exact, ok := bb.value.(bool); ok {
				req.Params.Exact = exact
			}
		case QDRANT_SCORE_THRESHOLD:
			if threshold, ok := bb.value.(float32); ok {
				req.ScoreThreshold = &threshold
			}
		case QDRANT_WITH_VECTOR:
			if withVec, ok := bb.value.(bool); ok {
				req.WithVector = withVec
			}
		case QDRANT_XX:
			// ⭐ 用户自定义参数
			// 注意：这些参数会被添加到 JSON 的顶层
			// 由于 QdrantSearchRequest 是固定结构，
			// QDRANT_XX 参数在 ToQdrantJSON() 中特殊处理
			// 这里只是标记，实际处理在 ToQdrantJSON()
		}
	}
}

// buildQdrantFilter 构建 Qdrant 过滤器
func buildQdrantFilter(bbs []Bb) (*QdrantFilter, error) {
	filter := &QdrantFilter{
		Must: []QdrantCondition{},
	}

	for _, bb := range bbs {
		// ⭐ 跳过向量专属操作符（单独处理）
		if isVectorOp(bb.op) {
			continue
		}

		// ⭐ 跳过 Qdrant 专属操作符（单独处理）
		if isQdrantOp(bb.op) {
			continue
		}

		cond, err := bbToQdrantCondition(bb)
		if err != nil {
			// ⭐ 关键：不支持的操作不报错，忽略即可
			continue
		}

		if cond != nil {
			filter.Must = append(filter.Must, *cond)
		}
	}

	return filter, nil
}

// isVectorOp 判断是否为向量操作符
func isVectorOp(op string) bool {
	return op == VECTOR_SEARCH || op == VECTOR_DISTANCE_FILTER
}

// isQdrantOp 判断是否为 Qdrant 专属操作符
func isQdrantOp(op string) bool {
	return op == QDRANT_HNSW_EF ||
		op == QDRANT_EXACT ||
		op == QDRANT_SCORE_THRESHOLD ||
		op == QDRANT_WITH_VECTOR ||
		op == QDRANT_XX
}

// bbToQdrantCondition 将 Bb 转换为 Qdrant 条件
func bbToQdrantCondition(bb Bb) (*QdrantCondition, error) {
	switch bb.op {
	case EQ:
		return &QdrantCondition{
			Key: bb.key,
			Match: &QdrantMatchCondition{
				Value: bb.value,
			},
		}, nil

	case NE:
		// 注意：Qdrant 的 != 需要用 must_not
		// 这里简化处理，调用者需要在上层处理
		return nil, fmt.Errorf("NE not directly supported, use must_not")

	case IN:
		// IN 转换为 match.any
		// 注意：IN 的 value 是 *[]string
		var anyValues []interface{}

		switch v := bb.value.(type) {
		case *[]string:
			if v == nil {
				return nil, nil
			}
			for _, s := range *v {
				anyValues = append(anyValues, s)
			}
		case []interface{}:
			anyValues = v
		case []string:
			for _, s := range v {
				anyValues = append(anyValues, s)
			}
		default:
			return nil, fmt.Errorf("IN operator expects []string or []interface{}, got %T", bb.value)
		}

		if len(anyValues) == 0 {
			return nil, nil
		}

		return &QdrantCondition{
			Key: bb.key,
			Match: &QdrantMatchCondition{
				Any: anyValues,
			},
		}, nil

	case GT:
		val, err := toFloat64(bb.value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.key,
			Range: &QdrantRangeCondition{
				Gt: &val,
			},
		}, nil

	case GTE:
		val, err := toFloat64(bb.value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.key,
			Range: &QdrantRangeCondition{
				Gte: &val,
			},
		}, nil

	case LT:
		val, err := toFloat64(bb.value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.key,
			Range: &QdrantRangeCondition{
				Lt: &val,
			},
		}, nil

	case LTE:
		val, err := toFloat64(bb.value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.key,
			Range: &QdrantRangeCondition{
				Lte: &val,
			},
		}, nil

	case LIKE:
		// Qdrant 不原生支持 LIKE，忽略
		return nil, fmt.Errorf("LIKE not supported in Qdrant")

	default:
		// 不支持的操作，返回 nil（忽略）
		return nil, nil
	}
}

// toFloat64 辅助函数：转换为 float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// QdrantDistanceMetric 转换距离度量
func QdrantDistanceMetric(metric VectorDistance) string {
	switch metric {
	case CosineDistance:
		return "Cosine"
	case L2Distance:
		return "Euclid"
	case InnerProduct:
		return "Dot"
	default:
		return "Cosine"
	}
}
