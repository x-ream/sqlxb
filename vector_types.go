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
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math"
)

// Vector 向量类型（兼容 PostgreSQL pgvector）
type Vector []float32

// Value 实现 driver.Valuer 接口
func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}

	// PostgreSQL pgvector 格式: '[1,2,3]'
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// 转换 [1,2,3] 为 '[1,2,3]' 字符串格式
	return string(bytes), nil
}

// Scan 实现 sql.Scanner 接口
func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}

	switch value := value.(type) {
	case []byte:
		return json.Unmarshal(value, v)
	case string:
		return json.Unmarshal([]byte(value), v)
	default:
		return fmt.Errorf("unsupported vector type: %T", value)
	}
}

// VectorDistance 向量距离度量类型
type VectorDistance string

const (
	// CosineDistance 余弦距离（最常用）
	// PostgreSQL: <->
	CosineDistance VectorDistance = "<->"

	// L2Distance 欧氏距离（L2范数）
	// PostgreSQL: <#>
	L2Distance VectorDistance = "<#>"

	// InnerProduct 内积距离（点积）
	// PostgreSQL: <=>
	InnerProduct VectorDistance = "<=>"
)

// Distance 计算两个向量的距离
func (v Vector) Distance(other Vector, metric VectorDistance) float32 {
	if len(v) != len(other) {
		panic("vectors must have same dimension")
	}

	switch metric {
	case CosineDistance:
		return cosineDistance(v, other)
	case L2Distance:
		return l2Distance(v, other)
	case InnerProduct:
		return innerProduct(v, other)
	default:
		return cosineDistance(v, other)
	}
}

// cosineDistance 计算余弦距离
// distance = 1 - (dot(a,b) / (||a|| * ||b||))
func cosineDistance(a, b Vector) float32 {
	var dotProduct, normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1.0 // 完全不相似
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	return 1.0 - (dotProduct / (normA * normB))
}

// l2Distance 计算欧氏距离（L2范数）
// distance = sqrt(sum((a[i] - b[i])^2))
func l2Distance(a, b Vector) float32 {
	var sum float32

	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}

// innerProduct 计算内积距离
// distance = -sum(a[i] * b[i])
// 注意：负号是因为排序时越大越相似
func innerProduct(a, b Vector) float32 {
	var sum float32

	for i := range a {
		sum += a[i] * b[i]
	}

	return -sum
}

// Normalize 向量归一化（L2范数）
func (v Vector) Normalize() Vector {
	var norm float32
	for _, val := range v {
		norm += val * val
	}

	if norm == 0 {
		return v
	}

	norm = float32(math.Sqrt(float64(norm)))
	normalized := make(Vector, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}

	return normalized
}

// Dim 返回向量维度
func (v Vector) Dim() int {
	return len(v)
}

// DiversityStrategy 多样性策略
type DiversityStrategy string

const (
	// DiversityByHash 基于语义哈希去重
	DiversityByHash DiversityStrategy = "hash"

	// DiversityByDistance 基于向量距离去重
	DiversityByDistance DiversityStrategy = "distance"

	// DiversityByMMR 使用 MMR（Maximal Marginal Relevance）算法
	DiversityByMMR DiversityStrategy = "mmr"
)

// DiversityParams 多样性查询参数
type DiversityParams struct {
	// Enabled 是否启用多样性
	Enabled bool

	// Strategy 多样性策略
	Strategy DiversityStrategy

	// HashField 语义哈希字段名（用于 DiversityByHash）
	// 例如: "semantic_hash", "content_hash"
	HashField string

	// MinDistance 结果之间的最小距离（用于 DiversityByDistance）
	// 例如: 0.3 表示结果之间的距离至少为 0.3
	MinDistance float32

	// Lambda MMR 平衡参数（用于 DiversityByMMR）
	// 范围: 0-1
	// 0 = 完全多样性
	// 1 = 完全相关性
	// 0.5 = 平衡（推荐）
	Lambda float32

	// OverFetchFactor 过度获取因子
	// 先获取 TopK * OverFetchFactor 个结果，再应用多样性过滤
	// 默认: 5（获取 5 倍的结果后过滤）
	OverFetchFactor int
}
