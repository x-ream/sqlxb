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

// ============================================================================
// 向量数据库通用接口（跨数据库抽象）
// ============================================================================

// VectorDBRequest 向量数据库请求通用接口
// 适用于所有向量数据库（Qdrant, Milvus, Weaviate, Pinecone 等）
//
// 设计原则：
//  1. 只包含所有向量数据库都支持的通用字段
//  2. 每个数据库通过继承此接口，添加专属字段
//  3. 使用 **Type 模式支持 nil 初始化和修改
//
// 示例：
//
//	// Qdrant 继承通用接口
//	type QdrantRequest interface {
//	    VectorDBRequest           // 继承
//	    GetParams() **QdrantSearchParams  // 专属字段
//	}
//
//	// Milvus 继承通用接口
//	type MilvusRequest interface {
//	    VectorDBRequest           // 继承
//	    GetSearchParams() **MilvusSearchParams  // 专属字段
//	}
type VectorDBRequest interface {
	// GetScoreThreshold 获取相似度阈值
	// 所有向量数据库都支持设置最低相似度阈值
	// 返回值: **float32 支持 nil 值判断和修改
	GetScoreThreshold() **float32

	// GetWithVector 是否返回向量数据
	// 控制查询结果是否包含原始向量（节省带宽）
	// 返回值: *bool 支持直接修改
	GetWithVector() *bool

	// GetFilter 获取过滤器
	// 不同数据库的过滤器结构不同：
	//  - Qdrant: *QdrantFilter
	//  - Milvus: *string (Expr)
	//  - Weaviate: *WeaviateFilter
	// 返回值: interface{} 允许任意类型，调用者需要类型断言
	GetFilter() interface{}
}

// ============================================================================
// 通用参数应用函数（跨数据库复用）
// ============================================================================

// ApplyCommonVectorParams 应用所有向量数据库通用的参数
// 这个函数可以被 Qdrant、Milvus、Weaviate 等所有数据库复用
//
// 参数:
//   - bbs: xb 的条件数组
//   - req: 实现了 VectorDBRequest 接口的任意请求对象
//
// 说明:
//   - 目前使用 QDRANT_SCORE_THRESHOLD 和 QDRANT_WITH_VECTOR（定义在 oper.go）
//   - 未来可以重命名为 VECTOR_SCORE_THRESHOLD（向量数据库通用）
//
// 示例:
//
//	// Qdrant 使用
//	ApplyCommonVectorParams(built.Conds, qdrantReq)
//
//	// Milvus 使用
//	ApplyCommonVectorParams(built.Conds, milvusReq)
func ApplyCommonVectorParams(bbs []Bb, req VectorDBRequest) {
	for _, bb := range bbs {
		switch bb.Op {
		// ⭐ 注意：目前使用 QDRANT_* 前缀（历史原因）
		// TODO(future): 重命名为 VECTOR_SCORE_THRESHOLD（所有向量数据库通用）
		case QDRANT_SCORE_THRESHOLD:
			if req.GetScoreThreshold() != nil {
				threshold := bb.Value.(float32)
				*req.GetScoreThreshold() = &threshold
			}

		case QDRANT_WITH_VECTOR:
			if req.GetWithVector() != nil {
				*req.GetWithVector() = bb.Value.(bool)
			}
		}
	}
}

// ============================================================================
// 通用辅助函数（跨数据库复用）
// ============================================================================

// ExtractCustomParams 提取用户自定义参数（通用版本）
// 可被 Qdrant、Milvus、Weaviate 等所有数据库复用
//
// 参数:
//   - bbs: xb 的条件数组
//   - customOp: 自定义操作符（QDRANT_XX, MILVUS_XX, WEAVIATE_XX 等）
//
// 返回:
//   - map[string]interface{}: 提取出的自定义参数
//
// 示例:
//
//	// Qdrant
//	customParams := ExtractCustomParams(bbs, QDRANT_XX)
//
//	// Milvus
//	customParams := ExtractCustomParams(bbs, MILVUS_XX)
func ExtractCustomParams(bbs []Bb, customOp string) map[string]interface{} {
	params := make(map[string]interface{})
	for _, bb := range bbs {
		if bb.Op == customOp {
			params[bb.Key] = bb.Value
		}
	}
	return params
}

// ============================================================================
// 未来扩展示例（注释说明）
// ============================================================================

/*
扩展示例：支持 Milvus

// 1. 定义 Milvus 专属请求接口
type MilvusRequest interface {
    VectorDBRequest  // 继承通用接口
    GetSearchParams() **MilvusSearchParams  // Milvus 专属
}

// 2. 实现接口
type MilvusSearchRequest struct {
    Vectors        [][]float32
    TopK           int
    MetricType     string
    ScoreThreshold *float32            // 通用字段
    WithVector     bool                // 通用字段
    SearchParams   *MilvusSearchParams // Milvus 专属
}

func (r *MilvusSearchRequest) GetScoreThreshold() **float32 {
    return &r.ScoreThreshold
}

func (r *MilvusSearchRequest) GetWithVector() *bool {
    return &r.WithVector
}

func (r *MilvusSearchRequest) GetFilter() interface{} {
    return nil // Milvus 使用 Expr，不是 Filter
}

func (r *MilvusSearchRequest) GetSearchParams() **MilvusSearchParams {
    return &r.SearchParams
}

// 3. 应用参数
func applyMilvusParams(bbs []Bb, req MilvusRequest) {
    // 复用通用参数应用
    ApplyCommonVectorParams(bbs, req)

    // 应用 Milvus 专属参数
    for _, bb := range bbs {
        switch bb.op {
        case MILVUS_NPROBE:
            params := req.GetSearchParams()
            if *params == nil {
                *params = &MilvusSearchParams{}
            }
            (*params).NProbe = bb.value.(int)
        }
    }
}
*/
