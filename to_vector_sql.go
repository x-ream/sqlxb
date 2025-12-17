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
	"fmt"
	"strings"
)

// SqlOfVectorSearch generates vector search SQL
// Returns: sql, args
//
// Example output:
//
//	SELECT *, embedding <-> ? AS distance
//	FROM code_vectors
//	WHERE language = ?
//	ORDER BY distance
//	LIMIT 10
func (built *Built) SqlOfVectorSearch() (string, []interface{}) {

	var sb strings.Builder
	var args []interface{}

	// 1. SELECT clause
	sb.WriteString("SELECT ")

	// Add fields
	if len(built.ResultKeys) > 0 {
		sb.WriteString(strings.Join(built.ResultKeys, ", "))
	} else {
		sb.WriteString("*")
	}

	// Find vector search parameters
	vectorBb := findVectorSearchBb(built.Conds)
	if vectorBb != nil {
		params := vectorBb.Value.(VectorSearchParams)

		// Add distance field
		sb.WriteString(fmt.Sprintf(
			", %s %s ? AS distance",
			vectorBb.Key,
			params.DistanceMetric,
		))
		args = append(args, params.QueryVector)
	}

	// 2. FROM clause
	sb.WriteString(" FROM ")
	sb.WriteString(built.OrFromSql)

	// 3. WHERE clause (scalar conditions + vector distance filtering)
	scalarConds := filterScalarConds(built.Conds)
	vectorDistConds := filterVectorDistanceConds(built.Conds)

	allConds := append(scalarConds, vectorDistConds...)

	if len(allConds) > 0 {
		sb.WriteString(" WHERE ")

		// ⭐ Use toCondSql instead of buildConditionSql to properly handle OR/AND subqueries
		if len(scalarConds) > 0 {
			built.toCondSql(scalarConds, &sb, &args, nil)
		}

		// Build vector distance filter conditions
		if len(vectorDistConds) > 0 {
			if len(scalarConds) > 0 {
				sb.WriteString(" AND ")
			}
			distSql, distArgs := buildVectorDistanceCondSql(vectorDistConds)
			sb.WriteString(distSql)
			args = append(args, distArgs...)
		}
	}

	// 4. ORDER BY distance
	if vectorBb != nil {
		sb.WriteString(" ORDER BY distance")
		params := vectorBb.Value.(VectorSearchParams)

		// 5. LIMIT Top-K
		sb.WriteString(fmt.Sprintf(" LIMIT %d", params.TopK))
	}

	return sb.String(), args
}

// Helper function: find vector search Bb
func findVectorSearchBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].Op == VECTOR_SEARCH {
			return &bbs[i]
		}
	}
	return nil
}

// Helper function: filter scalar conditions
func filterScalarConds(bbs []Bb) []Bb {
	result := []Bb{}
	for _, bb := range bbs {
		// Skip vector operators
		if bb.Op == VECTOR_SEARCH || bb.Op == VECTOR_DISTANCE_FILTER {
			continue
		}
		// ⭐ Skip Qdrant-specific operators (PostgreSQL doesn't support)
		if isQdrantSpecificOp(bb.Op) {
			continue
		}
		result = append(result, bb)
	}
	return result
}

// isQdrantSpecificOp checks if operator is Qdrant-specific
func isQdrantSpecificOp(op string) bool {
	return op == QDRANT_HNSW_EF ||
		op == QDRANT_EXACT ||
		op == QDRANT_SCORE_THRESHOLD ||
		op == QDRANT_WITH_VECTOR ||
		op == QDRANT_XX
}

// Helper function: filter vector distance conditions
func filterVectorDistanceConds(bbs []Bb) []Bb {
	result := []Bb{}
	for _, bb := range bbs {
		if bb.Op == VECTOR_DISTANCE_FILTER {
			result = append(result, bb)
		}
	}
	return result
}

// Helper function: build condition SQL (reuse existing logic)
func buildConditionSql(bbs []Bb) (string, []interface{}) {
	// TODO: Need to reuse buildCondSql logic from to_sql.go
	// Temporarily simplified implementation
	if len(bbs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	var args []interface{}

	for i, bb := range bbs {
		if i > 0 {
			sb.WriteString(" AND ")
		}

		// Simplified handling: only process basic operators
		sb.WriteString(bb.Key)
		sb.WriteString(" ")
		sb.WriteString(bb.Op)
		sb.WriteString(" ?")

		args = append(args, bb.Value)
	}

	return sb.String(), args
}

// Helper function: build vector distance filter SQL
func buildVectorDistanceCondSql(bbs []Bb) (string, []interface{}) {
	if len(bbs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	var args []interface{}

	for i, bb := range bbs {
		if i > 0 {
			sb.WriteString(" AND ")
		}

		params := bb.Value.(VectorDistanceFilterParams)

		// (field <-> ?) op threshold
		sb.WriteString("(")
		sb.WriteString(bb.Key)
		sb.WriteString(" ")
		sb.WriteString(string(params.DistanceMetric))
		sb.WriteString(" ?)")
		sb.WriteString(" ")
		sb.WriteString(params.Operator)
		sb.WriteString(" ?")

		args = append(args, params.QueryVector, params.Threshold)
	}

	return sb.String(), args
}
