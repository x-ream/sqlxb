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
package xb

// QdrantBuilderX Qdrant 专属构建器
// 提供 Qdrant 专属的高层 API
type QdrantBuilderX struct {
	// 内部引用 BuilderX
	builder *BuilderX
}

// QdrantX 使用 Qdrant 专属构建器
// ⭐ 只包含 Qdrant 专属的配置方法（HNSW, ScoreThreshold 等）
// ⭐ 通用方法（VectorSearch, WithHashDiversity）在外部调用
//
// 示例:
//   sqlxb.Of(&CodeVector{}).
//       Eq("language", "golang").              // 通用条件
//       VectorSearch("embedding", vec, 20).    // 通用向量检索
//       WithHashDiversity("semantic_hash").    // 通用多样性
//       QdrantX(func(qx *QdrantBuilderX) {
//           qx.HnswEf(256).                    // ⭐ Qdrant 专属
//              ScoreThreshold(0.8).            // ⭐ Qdrant 专属
//              WithVector(false)               // ⭐ Qdrant 专属
//       }).
//       Build()
func (x *BuilderX) QdrantX(f func(qx *QdrantBuilderX)) *BuilderX {
	qx := &QdrantBuilderX{
		builder: x,
	}

	f(qx)

	return x
}

// HnswEf 设置 HNSW 算法的 ef 参数
// ef 越大，查询精度越高，但速度越慢
// 推荐值: 64-256
func (qx *QdrantBuilderX) HnswEf(ef int) *QdrantBuilderX {
	if ef > 0 {
		bb := Bb{
			op:    QDRANT_HNSW_EF,
			key:   "hnsw_ef",
			value: ef,
		}
		qx.builder.bbs = append(qx.builder.bbs, bb)
	}
	return qx
}

// Exact 设置是否使用精确搜索（不使用索引）
// true: 精确搜索（慢但准确）
// false: 近似搜索（快但可能略不准）
func (qx *QdrantBuilderX) Exact(exact bool) *QdrantBuilderX {
	bb := Bb{
		op:    QDRANT_EXACT,
		key:   "exact",
		value: exact,
	}
	qx.builder.bbs = append(qx.builder.bbs, bb)
	return qx
}

// ScoreThreshold 设置最小相似度阈值
// 只返回相似度 >= threshold 的结果
func (qx *QdrantBuilderX) ScoreThreshold(threshold float32) *QdrantBuilderX {
	bb := Bb{
		op:    QDRANT_SCORE_THRESHOLD,
		key:   "score_threshold",
		value: threshold,
	}
	qx.builder.bbs = append(qx.builder.bbs, bb)
	return qx
}

// WithVector 设置是否返回向量数据
// true: 返回向量（占用带宽）
// false: 不返回向量（节省带宽）
func (qx *QdrantBuilderX) WithVector(withVector bool) *QdrantBuilderX {
	bb := Bb{
		op:    QDRANT_WITH_VECTOR,
		key:   "with_vector",
		value: withVector,
	}
	qx.builder.bbs = append(qx.builder.bbs, bb)
	return qx
}

// HighPrecision 高精度模式（牺牲速度）
func (qx *QdrantBuilderX) HighPrecision() *QdrantBuilderX {
	return qx.HnswEf(512).Exact(false)
}

// Balanced 平衡模式（默认）
func (qx *QdrantBuilderX) Balanced() *QdrantBuilderX {
	return qx.HnswEf(128).Exact(false)
}

// HighSpeed 高速模式（牺牲精度）
func (qx *QdrantBuilderX) HighSpeed() *QdrantBuilderX {
	return qx.HnswEf(32).Exact(false)
}

// Recommend 基于正负样本的推荐查询
//
// 示例:
//   qx.Recommend(func(rb *RecommendBuilder) {
//       rb.Positive(123, 456, 789)
//       rb.Negative(111, 222)
//       rb.Limit(20)
//   })
func (qx *QdrantBuilderX) Recommend(fn func(rb *RecommendBuilder)) *QdrantBuilderX {
	rb := &RecommendBuilder{}
	fn(rb)

	if len(rb.positive) > 0 && rb.limit > 0 {
		bb := Bb{
			op:  QDRANT_RECOMMEND,
			key: "recommend",
			value: map[string]interface{}{
				"positive": rb.positive,
				"negative": rb.negative,
				"limit":    rb.limit,
			},
		}
		qx.builder.bbs = append(qx.builder.bbs, bb)
	}
	return qx
}

// RecommendBuilder 推荐查询构建器
type RecommendBuilder struct {
	positive []int64
	negative []int64
	limit    int
}

// Positive 设置正样本（用户喜欢的）
func (rb *RecommendBuilder) Positive(ids ...int64) *RecommendBuilder {
	rb.positive = ids
	return rb
}

// Negative 设置负样本（用户不喜欢的）
func (rb *RecommendBuilder) Negative(ids ...int64) *RecommendBuilder {
	rb.negative = ids
	return rb
}

// Limit 设置返回数量
func (rb *RecommendBuilder) Limit(limit int) *RecommendBuilder {
	rb.limit = limit
	return rb
}

// Discover 基于上下文向量的探索性查询
// 在一组向量的"中间地带"发现新内容
//
// 示例:
//   qx.Discover(func(db *DiscoverBuilder) {
//       db.Context(101, 102, 103)  // 用户浏览历史
//       db.Limit(20)
//   })
func (qx *QdrantBuilderX) Discover(fn func(db *DiscoverBuilder)) *QdrantBuilderX {
	db := &DiscoverBuilder{}
	fn(db)

	if len(db.context) > 0 && db.limit > 0 {
		bb := Bb{
			op:  QDRANT_DISCOVER,
			key: "discover",
			value: map[string]interface{}{
				"context": db.context,
				"limit":   db.limit,
			},
		}
		qx.builder.bbs = append(qx.builder.bbs, bb)
	}
	return qx
}

// DiscoverBuilder 探索查询构建器
type DiscoverBuilder struct {
	context []int64
	limit   int
}

// Context 设置上下文样本（用户浏览/交互历史）
func (db *DiscoverBuilder) Context(ids ...int64) *DiscoverBuilder {
	db.context = ids
	return db
}

// Limit 设置返回数量
func (db *DiscoverBuilder) Limit(limit int) *DiscoverBuilder {
	db.limit = limit
	return db
}

// ScrollID 设置 Scroll 查询 ID（用于大数据集遍历）
func (qx *QdrantBuilderX) ScrollID(scrollID string) *QdrantBuilderX {
	if scrollID != "" {
		bb := Bb{
			op:    QDRANT_SCROLL,
			key:   "scroll_id",
			value: scrollID,
		}
		qx.builder.bbs = append(qx.builder.bbs, bb)
	}
	return qx
}

// X Qdrant 用户自定义专属参数
// 用于设置 Qdrant 未来可能新增的参数，或 sqlxb 未封装的参数
//
// 示例:
//   qx.X("quantization", map[string]interface{}{
//       "rescore": true,
//   })
func (qx *QdrantBuilderX) X(k string, v interface{}) *QdrantBuilderX {
	bb := Bb{
		op:    QDRANT_XX, // ⭐ Qdrant 专属的 X
		key:   k,
		value: v,
	}
	qx.builder.bbs = append(qx.builder.bbs, bb)
	return qx
}
