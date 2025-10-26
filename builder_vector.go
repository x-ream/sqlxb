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

// VectorSearch 向量相似度检索（BuilderX 扩展）
// 与 CondBuilder.VectorSearch() 功能相同，但返回 *BuilderX 用于链式调用
//
// 示例:
//   sqlxb.Of(&CodeVector{}).
//       Eq("language", "golang").
//       VectorSearch("embedding", queryVector, 10).
//       Build().
//       SqlOfVectorSearch()
func (x *BuilderX) VectorSearch(field string, queryVector Vector, topK int) *BuilderX {
	x.CondBuilder.VectorSearch(field, queryVector, topK)
	return x
}

// VectorDistance 设置向量距离度量（BuilderX 扩展）
//
// 示例:
//   builder.VectorSearch("embedding", vec, 10).
//       VectorDistance(sqlxb.L2Distance)
func (x *BuilderX) VectorDistance(metric VectorDistance) *BuilderX {
	x.CondBuilder.VectorDistance(metric)
	return x
}

// VectorDistanceFilter 向量距离过滤（BuilderX 扩展）
//
// 示例:
//   builder.VectorDistanceFilter("embedding", queryVector, "<", 0.3)
func (x *BuilderX) VectorDistanceFilter(
	field string,
	queryVector Vector,
	op string,
	threshold float32,
) *BuilderX {
	x.CondBuilder.VectorDistanceFilter(field, queryVector, op, threshold)
	return x
}

// WithDiversity 设置多样性参数（BuilderX 扩展）
// ⭐ 核心：如果数据库不支持，会被自动忽略
//
// 示例:
//   sqlxb.Of(&CodeVector{}).
//       VectorSearch("embedding", vec, 20).
//       WithDiversity(sqlxb.DiversityByHash, "semantic_hash").
//       Build()
func (x *BuilderX) WithDiversity(strategy DiversityStrategy, params ...interface{}) *BuilderX {
	x.CondBuilder.WithDiversity(strategy, params...)
	return x
}

// WithMinDistance 设置最小距离多样性（BuilderX 扩展）
//
// 示例:
//   sqlxb.Of(&CodeVector{}).
//       VectorSearch("embedding", vec, 20).
//       WithMinDistance(0.3).
//       Build()
func (x *BuilderX) WithMinDistance(minDistance float32) *BuilderX {
	x.CondBuilder.WithMinDistance(minDistance)
	return x
}

// WithHashDiversity 设置哈希去重（BuilderX 扩展）
//
// 示例:
//   sqlxb.Of(&CodeVector{}).
//       VectorSearch("embedding", vec, 20).
//       WithHashDiversity("semantic_hash").
//       Build()
func (x *BuilderX) WithHashDiversity(hashField string) *BuilderX {
	x.CondBuilder.WithHashDiversity(hashField)
	return x
}

// WithMMR 设置 MMR 算法（BuilderX 扩展）
//
// 示例:
//   sqlxb.Of(&CodeVector{}).
//       VectorSearch("embedding", vec, 20).
//       WithMMR(0.5).
//       Build()
func (x *BuilderX) WithMMR(lambda float32) *BuilderX {
	x.CondBuilder.WithMMR(lambda)
	return x
}
