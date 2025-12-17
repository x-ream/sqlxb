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
// Vector database common interface (cross-database abstraction)
// ============================================================================

// VectorDBRequest Vector database request common interface
// Suitable for all vector databases (Qdrant, Milvus, Weaviate, Pinecone, etc.)
//
// Design principles:
//  1. Only include common fields supported by all vector databases
//  2. Each database inherits this interface and adds its own fields
//  3. Use **Type mode to support nil initialization and modification
//
// Example:
//
//	// Qdrant inherits common interface
//	type QdrantRequest interface {
//	    VectorDBRequest           // Inherits common interface
//	    GetParams() **QdrantSearchParams  // Own fields
//	}
//
//	// Milvus inherits common interface
//	type MilvusRequest interface {
//	    VectorDBRequest           // Inherits common interface
//	    GetSearchParams() **MilvusSearchParams  // Own fields
//	}
type VectorDBRequest interface {
	// GetScoreThreshold get similarity threshold
	// All vector databases support setting the minimum similarity threshold
	// Return value: **float32 supports nil value judgment and modification
	GetScoreThreshold() **float32

	// GetWithVector whether to return vector data
	// Control whether the query result contains the original vector (saving bandwidth)
	// Return value: *bool supports direct modification
	GetWithVector() *bool

	// GetFilter get filter
	// The filter structure is different for different databases:
	//  - Qdrant: *QdrantFilter
	//  - Milvus: *string (Expr)
	//  - Weaviate: *WeaviateFilter
	// Return value: interface{} allows any type, the caller needs to type assert
	GetFilter() interface{}
}

// ============================================================================
// Common parameter application function (cross-database reuse)
// ============================================================================

// ApplyCommonVectorParams apply all vector database common parameters
// This function can be reused by all databases (Qdrant, Milvus, Weaviate, etc.)
//
// Parameters:
//   - bbs: xb's condition array
//   - req: any request object that implements the VectorDBRequest interface
//
// Example:
//
//	// Qdrant uses
//	ApplyCommonVectorParams(built.Conds, qdrantReq)
//
//	// Milvus uses
//	ApplyCommonVectorParams(built.Conds, milvusReq)
func ApplyCommonVectorParams(bbs []Bb, req VectorDBRequest) {
	for _, bb := range bbs {
		switch bb.Op {
		// ⭐ Note: Currently using QDRANT_* prefix (historical reasons)
		// TODO(future): Rename to VECTOR_SCORE_THRESHOLD (all vector databases common)
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
// Common helper function (cross-database reuse)
// ============================================================================

// ExtractCustomParams extract user-defined parameters (common version)
// Can be reused by all databases (Qdrant, Milvus, Weaviate, etc.)
//
// Parameters:
//   - bbs: xb's condition array
//   - customOp: custom operator (QDRANT_XX, MILVUS_XX, WEAVIATE_XX, etc.)
//
// Return:
//   - map[string]interface{}: extracted user-defined parameters
//
// Example:
//
//	// Qdrant uses
//	customParams := ExtractCustomParams(bbs, QDRANT_XX)
//
//	// Milvus uses
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
// Future expansion example (comment description)
// ============================================================================

/*
Future expansion example: support Milvus

// 1. Define Milvus own request interface
type MilvusRequest interface {
    VectorDBRequest  // Inherits common interface
    GetSearchParams() **MilvusSearchParams  // Milvus own fields
}

// 2. Implement interface
type MilvusSearchRequest struct {
    Vectors        [][]float32
    TopK           int
	MetricType     string
	ScoreThreshold *float32            // Common fields
	WithVector     bool                // Common fields
	SearchParams   *MilvusSearchParams // Milvus own fields
}

func (r *MilvusSearchRequest) GetScoreThreshold() **float32 {
    return &r.ScoreThreshold
}

func (r *MilvusSearchRequest) GetWithVector() *bool {
    return &r.WithVector
}

func (r *MilvusSearchRequest) GetFilter() interface{} {
    return nil // Milvus uses Expr, not Filter
}

func (r *MilvusSearchRequest) GetSearchParams() **MilvusSearchParams {
    return &r.SearchParams
}

// 3. Apply parameters
func applyMilvusParams(bbs []Bb, req MilvusRequest) {
    // Reuse common parameter application
    ApplyCommonVectorParams(bbs, req)

    // Apply Milvus own parameters
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
