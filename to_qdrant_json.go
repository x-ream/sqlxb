// Copyright 2020 io.xream.sqlxb
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
package sqlxb

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

	bytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
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
		HnswEf:      128, // 默认值，可优化
		Exact:       false,
		IndexedOnly: false,
	}

	return req, nil
}

// buildQdrantFilter 构建 Qdrant 过滤器
func buildQdrantFilter(bbs []Bb) (*QdrantFilter, error) {
	filter := &QdrantFilter{
		Must: []QdrantCondition{},
	}

	for _, bb := range bbs {
		// 跳过向量检索和向量距离过滤（单独处理）
		if bb.op == VECTOR_SEARCH || bb.op == VECTOR_DISTANCE_FILTER {
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
