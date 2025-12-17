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

// VectorSearch vector similarity search
// field: vector field name
// queryVector: query vector
// topK: returns Top-K most similar results
//
// Example:
//
//	builder.VectorSearch("embedding", queryVector, 10)
//
// Generates SQL:
//
//	ORDER BY embedding <-> $1 LIMIT 10
func (cb *CondBuilder) VectorSearch(field string, queryVector Vector, topK int) *CondBuilder {

	// Parameter validation (automatically ignores invalid parameters)
	if field == "" || queryVector == nil || len(queryVector) == 0 {
		return cb
	}

	if topK <= 0 {
		topK = 10 // Default value
	}

	// Create vector search Bb
	bb := Bb{
		Op:  VECTOR_SEARCH,
		Key: field,
		Value: VectorSearchParams{
			QueryVector:    queryVector,
			TopK:           topK,
			DistanceMetric: CosineDistance, // Default cosine distance
		},
	}

	cb.bbs = append(cb.bbs, bb)
	return cb
}

// VectorDistance sets vector distance metric
// Must be called after VectorSearch()
//
// Example:
//
//	builder.VectorSearch("embedding", vec, 10).VectorDistance(xb.L2Distance)
func (cb *CondBuilder) VectorDistance(metric VectorDistance) *CondBuilder {

	// Find the last VECTOR_SEARCH
	length := len(cb.bbs)
	if length == 0 {
		return cb
	}

	for i := length - 1; i >= 0; i-- {
		if cb.bbs[i].Op == VECTOR_SEARCH {
			// Modify distance metric
			if params, ok := cb.bbs[i].Value.(VectorSearchParams); ok {
				params.DistanceMetric = metric
				cb.bbs[i].Value = params
			}
			break
		}
	}

	return cb
}

// VectorDistanceFilter vector distance filtering
// Used for: WHERE distance < threshold
//
// Example:
//
//	builder.VectorDistanceFilter("embedding", queryVector, "<", 0.3)
//
// Generates SQL:
//
//	WHERE (embedding <-> $1) < 0.3
func (cb *CondBuilder) VectorDistanceFilter(
	field string,
	queryVector Vector,
	op string, // <, <=, >, >=, =
	threshold float32,
) *CondBuilder {

	// Parameter validation
	if field == "" || queryVector == nil || len(queryVector) == 0 {
		return cb
	}

	if op == "" {
		op = "<" // Default less than
	}

	// Create vector distance filter Bb
	bb := Bb{
		Op:  VECTOR_DISTANCE_FILTER,
		Key: field,
		Value: VectorDistanceFilterParams{
			QueryVector:    queryVector,
			Operator:       op,
			Threshold:      threshold,
			DistanceMetric: CosineDistance, // Default cosine distance
		},
	}

	cb.bbs = append(cb.bbs, bb)
	return cb
}

// VectorSearchParams vector search parameters
type VectorSearchParams struct {
	QueryVector    Vector
	TopK           int
	DistanceMetric VectorDistance
	Diversity      *DiversityParams // ⭐ Added: diversity parameters (optional)
}

// VectorDistanceFilterParams vector distance filter parameters
type VectorDistanceFilterParams struct {
	QueryVector    Vector
	Operator       string // <, <=, >, >=, =
	Threshold      float32
	DistanceMetric VectorDistance
}

// WithDiversity chain sets diversity parameters
// ⭐ Core: if database doesn't support, will be automatically ignored
//
// Example:
//
//	builder.VectorSearch("embedding", vec, 20).
//	    WithDiversity(xb.DiversityByHash, "semantic_hash")
func (cb *CondBuilder) WithDiversity(strategy DiversityStrategy, params ...interface{}) *CondBuilder {
	// Find the last VECTOR_SEARCH
	length := len(cb.bbs)
	if length == 0 {
		return cb
	}

	for i := length - 1; i >= 0; i-- {
		if cb.bbs[i].Op == VECTOR_SEARCH {
			searchParams, ok := cb.bbs[i].Value.(VectorSearchParams)
			if !ok {
				return cb
			}

			// Initialize DiversityParams
			if searchParams.Diversity == nil {
				searchParams.Diversity = &DiversityParams{
					Enabled:         true,
					Strategy:        strategy,
					OverFetchFactor: 5, // Default 5x over-fetch
				}
			}

			searchParams.Diversity.Strategy = strategy

			// Set parameters according to strategy
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
					searchParams.Diversity.Lambda = 0.5 // Default balanced
				}
			}

			// Optional: set over-fetch factor
			if len(params) > 1 {
				if factor, ok := params[1].(int); ok && factor > 0 {
					searchParams.Diversity.OverFetchFactor = factor
				}
			}

			cb.bbs[i].Value = searchParams
			break
		}
	}

	return cb
}

// WithMinDistance convenience method: sets minimum distance diversity
//
// Example:
//
//	builder.VectorSearch("embedding", vec, 20).
//	    WithMinDistance(0.3)
func (cb *CondBuilder) WithMinDistance(minDistance float32) *CondBuilder {
	return cb.WithDiversity(DiversityByDistance, minDistance)
}

// WithHashDiversity convenience method: sets hash deduplication
//
// Example:
//
//	builder.VectorSearch("embedding", vec, 20).
//	    WithHashDiversity("semantic_hash")
func (cb *CondBuilder) WithHashDiversity(hashField string) *CondBuilder {
	return cb.WithDiversity(DiversityByHash, hashField)
}

// WithMMR convenience method: sets MMR algorithm
//
// Example:
//
//	builder.VectorSearch("embedding", vec, 20).
//	    WithMMR(0.5)  // lambda = 0.5, balances relevance and diversity
func (cb *CondBuilder) WithMMR(lambda float32) *CondBuilder {
	return cb.WithDiversity(DiversityByMMR, lambda)
}
