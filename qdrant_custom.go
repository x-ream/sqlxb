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
}

// NewQdrantCustom 创建 Qdrant Custom（默认配置）
func NewQdrantCustom() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:         128,
		DefaultScoreThreshold: 0.0,
		DefaultWithVector:     true,
	}
}

// Generate 实现 Custom 接口
// ⭐ 根据操作类型返回不同的 JSON
func (c *QdrantCustom) Generate(built *Built) (interface{}, error) {
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
	json, err := built.toQdrantJSON()
	return json, err
}

// ============================================================================
// 便捷构造函数
// ============================================================================

// QdrantHighPrecision 高精度模式（慢，但准确）
func QdrantHighPrecision() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:         512,
		DefaultScoreThreshold: 0.85,
		DefaultWithVector:     true,
	}
}

// QdrantHighSpeed 高速模式（快，但可能不太准）
func QdrantHighSpeed() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:         32,
		DefaultScoreThreshold: 0.5,
		DefaultWithVector:     false,
	}
}

// QdrantBalanced 平衡模式（默认）
func QdrantBalanced() *QdrantCustom {
	return NewQdrantCustom()
}

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

	for _, bb := range inserts {
		// bb.Value 应该是一个包含 id, vector, payload 的 map 或 struct
		point, err := c.extractPoint(bb.Value)
		if err != nil {
			return "", err
		}
		points = append(points, point)
	}

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

// extractPoint 从 insert 数据中提取 Qdrant Point
func (c *QdrantCustom) extractPoint(value interface{}) (QdrantPoint, error) {
	// 尝试类型断言为 map
	if m, ok := value.(map[string]interface{}); ok {
		point := QdrantPoint{}

		if id, exists := m["id"]; exists {
			point.ID = id
		} else {
			return QdrantPoint{}, fmt.Errorf("point must have 'id' field")
		}

		if vector, exists := m["vector"]; exists {
			point.Vector = vector
		} else {
			return QdrantPoint{}, fmt.Errorf("point must have 'vector' field")
		}

		// payload 是可选的
		if payload, exists := m["payload"]; exists {
			if p, ok := payload.(map[string]interface{}); ok {
				point.Payload = p
			}
		}

		return point, nil
	}

	// 如果是 struct，尝试通过反射提取
	// （简化实现：直接返回错误，要求使用 map）
	return QdrantPoint{}, fmt.Errorf("insert value must be map[string]interface{} with 'id', 'vector', and optional 'payload' fields")
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
