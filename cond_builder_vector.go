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

// VectorSearch 向量相似度检索
// field: 向量字段名
// queryVector: 查询向量
// topK: 返回 Top-K 个最相似的结果
//
// 示例:
//   builder.VectorSearch("embedding", queryVector, 10)
//
// 生成 SQL:
//   ORDER BY embedding <-> $1 LIMIT 10
func (cb *CondBuilder) VectorSearch(field string, queryVector Vector, topK int) *CondBuilder {

	// 参数验证（自动忽略无效参数）
	if field == "" || queryVector == nil || len(queryVector) == 0 {
		return cb
	}

	if topK <= 0 {
		topK = 10 // 默认值
	}

	// 创建向量检索 Bb
	bb := Bb{
		op:  VECTOR_SEARCH,
		key: field,
		value: VectorSearchParams{
			QueryVector:    queryVector,
			TopK:           topK,
			DistanceMetric: CosineDistance, // 默认余弦距离
		},
	}

	cb.bbs = append(cb.bbs, bb)
	return cb
}

// VectorDistance 设置向量距离度量
// 必须在 VectorSearch() 之后调用
//
// 示例:
//   builder.VectorSearch("embedding", vec, 10).VectorDistance(sqlxb.L2Distance)
func (cb *CondBuilder) VectorDistance(metric VectorDistance) *CondBuilder {

	// 找到最后一个 VECTOR_SEARCH
	length := len(cb.bbs)
	if length == 0 {
		return cb
	}

	for i := length - 1; i >= 0; i-- {
		if cb.bbs[i].op == VECTOR_SEARCH {
			// 修改距离度量
			if params, ok := cb.bbs[i].value.(VectorSearchParams); ok {
				params.DistanceMetric = metric
				cb.bbs[i].value = params
			}
			break
		}
	}

	return cb
}

// VectorDistanceFilter 向量距离过滤
// 用于: WHERE distance < threshold
//
// 示例:
//   builder.VectorDistanceFilter("embedding", queryVector, "<", 0.3)
//
// 生成 SQL:
//   WHERE (embedding <-> $1) < 0.3
func (cb *CondBuilder) VectorDistanceFilter(
	field string,
	queryVector Vector,
	op string, // <, <=, >, >=, =
	threshold float32,
) *CondBuilder {

	// 参数验证
	if field == "" || queryVector == nil || len(queryVector) == 0 {
		return cb
	}

	if op == "" {
		op = "<" // 默认小于
	}

	// 创建向量距离过滤 Bb
	bb := Bb{
		op:  VECTOR_DISTANCE_FILTER,
		key: field,
		value: VectorDistanceFilterParams{
			QueryVector:    queryVector,
			Operator:       op,
			Threshold:      threshold,
			DistanceMetric: CosineDistance, // 默认余弦距离
		},
	}

	cb.bbs = append(cb.bbs, bb)
	return cb
}

// VectorSearchParams 向量检索参数
type VectorSearchParams struct {
	QueryVector    Vector
	TopK           int
	DistanceMetric VectorDistance
	Diversity      *DiversityParams // ⭐ 新增：多样性参数（可选）
}

// VectorDistanceFilterParams 向量距离过滤参数
type VectorDistanceFilterParams struct {
	QueryVector    Vector
	Operator       string // <, <=, >, >=, =
	Threshold      float32
	DistanceMetric VectorDistance
}

// WithDiversity 链式设置多样性参数
// ⭐ 核心：如果数据库不支持，会被自动忽略
//
// 示例:
//   builder.VectorSearch("embedding", vec, 20).
//       WithDiversity(sqlxb.DiversityByHash, "semantic_hash")
func (cb *CondBuilder) WithDiversity(strategy DiversityStrategy, params ...interface{}) *CondBuilder {
	// 找到最后一个 VECTOR_SEARCH
	length := len(cb.bbs)
	if length == 0 {
		return cb
	}

	for i := length - 1; i >= 0; i-- {
		if cb.bbs[i].op == VECTOR_SEARCH {
			searchParams, ok := cb.bbs[i].value.(VectorSearchParams)
			if !ok {
				return cb
			}

			// 初始化 DiversityParams
			if searchParams.Diversity == nil {
				searchParams.Diversity = &DiversityParams{
					Enabled:         true,
					Strategy:        strategy,
					OverFetchFactor: 5, // 默认 5 倍过度获取
				}
			}

			searchParams.Diversity.Strategy = strategy

			// 根据策略设置参数
			switch strategy {
			case DiversityByHash:
				if len(params) > 0 {
					if hashField, ok := params[0].(string); ok {
						searchParams.Diversity.HashField = hashField
					}
				}

			case DiversityByDistance:
				if len(params) > 0 {
					if minDist, ok := params[0].(float32); ok {
						searchParams.Diversity.MinDistance = minDist
					} else if minDist, ok := params[0].(float64); ok {
						searchParams.Diversity.MinDistance = float32(minDist)
					}
				}

			case DiversityByMMR:
				if len(params) > 0 {
					if lambda, ok := params[0].(float32); ok {
						searchParams.Diversity.Lambda = lambda
					} else if lambda, ok := params[0].(float64); ok {
						searchParams.Diversity.Lambda = float32(lambda)
					}
				} else {
					searchParams.Diversity.Lambda = 0.5 // 默认平衡
				}
			}

			// 可选：设置过度获取因子
			if len(params) > 1 {
				if factor, ok := params[1].(int); ok && factor > 0 {
					searchParams.Diversity.OverFetchFactor = factor
				}
			}

			cb.bbs[i].value = searchParams
			break
		}
	}

	return cb
}

// WithMinDistance 快捷方法：设置最小距离多样性
//
// 示例:
//   builder.VectorSearch("embedding", vec, 20).
//       WithMinDistance(0.3)
func (cb *CondBuilder) WithMinDistance(minDistance float32) *CondBuilder {
	return cb.WithDiversity(DiversityByDistance, minDistance)
}

// WithHashDiversity 快捷方法：设置哈希去重
//
// 示例:
//   builder.VectorSearch("embedding", vec, 20).
//       WithHashDiversity("semantic_hash")
func (cb *CondBuilder) WithHashDiversity(hashField string) *CondBuilder {
	return cb.WithDiversity(DiversityByHash, hashField)
}

// WithMMR 快捷方法：设置 MMR 算法
//
// 示例:
//   builder.VectorSearch("embedding", vec, 20).
//       WithMMR(0.5)  // lambda = 0.5，平衡相关性和多样性
func (cb *CondBuilder) WithMMR(lambda float32) *CondBuilder {
	return cb.WithDiversity(DiversityByMMR, lambda)
}
