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
