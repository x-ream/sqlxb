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

// SqlOfVectorSearch 生成向量检索 SQL
// 返回: sql, args
//
// 示例输出:
//
//	SELECT *, embedding <-> ? AS distance
//	FROM code_vectors
//	WHERE language = ?
//	ORDER BY distance
//	LIMIT 10
func (built *Built) SqlOfVectorSearch() (string, []interface{}) {

	var sb strings.Builder
	var args []interface{}

	// 1. SELECT 子句
	sb.WriteString("SELECT ")

	// 添加字段
	if len(built.ResultKeys) > 0 {
		sb.WriteString(strings.Join(built.ResultKeys, ", "))
	} else {
		sb.WriteString("*")
	}

	// 查找向量检索参数
	vectorBb := findVectorSearchBb(built.Conds)
	if vectorBb != nil {
		params := vectorBb.value.(VectorSearchParams)

		// 添加距离字段
		sb.WriteString(fmt.Sprintf(
			", %s %s ? AS distance",
			vectorBb.key,
			params.DistanceMetric,
		))
		args = append(args, params.QueryVector)
	}

	// 2. FROM 子句
	sb.WriteString(" FROM ")
	sb.WriteString(built.OrFromSql)

	// 3. WHERE 子句（标量条件 + 向量距离过滤）
	scalarConds := filterScalarConds(built.Conds)
	vectorDistConds := filterVectorDistanceConds(built.Conds)

	allConds := append(scalarConds, vectorDistConds...)

	if len(allConds) > 0 {
		sb.WriteString(" WHERE ")

		// ⭐ 使用 toCondSql 而不是 buildConditionSql，以正确处理 OR/AND 子查询
		if len(scalarConds) > 0 {
			built.toCondSql(scalarConds, &sb, &args, nil)
		}

		// 构建向量距离过滤条件
		if len(vectorDistConds) > 0 {
			if len(scalarConds) > 0 {
				sb.WriteString(" AND ")
			}
			distSql, distArgs := buildVectorDistanceCondSql(vectorDistConds)
			sb.WriteString(distSql)
			args = append(args, distArgs...)
		}
	}

	// 4. ORDER BY 距离
	if vectorBb != nil {
		sb.WriteString(" ORDER BY distance")
		params := vectorBb.value.(VectorSearchParams)

		// 5. LIMIT Top-K
		sb.WriteString(fmt.Sprintf(" LIMIT %d", params.TopK))
	}

	return sb.String(), args
}

// 辅助函数：查找向量检索 Bb
func findVectorSearchBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].op == VECTOR_SEARCH {
			return &bbs[i]
		}
	}
	return nil
}

// 辅助函数：过滤标量条件
func filterScalarConds(bbs []Bb) []Bb {
	result := []Bb{}
	for _, bb := range bbs {
		// 跳过向量操作符
		if bb.op == VECTOR_SEARCH || bb.op == VECTOR_DISTANCE_FILTER {
			continue
		}
		// ⭐ 跳过 Qdrant 专属操作符（PostgreSQL 不支持）
		if isQdrantSpecificOp(bb.op) {
			continue
		}
		result = append(result, bb)
	}
	return result
}

// isQdrantSpecificOp 判断是否为 Qdrant 专属操作符
func isQdrantSpecificOp(op string) bool {
	return op == QDRANT_HNSW_EF ||
		op == QDRANT_EXACT ||
		op == QDRANT_SCORE_THRESHOLD ||
		op == QDRANT_WITH_VECTOR ||
		op == QDRANT_XX
}

// 辅助函数：过滤向量距离条件
func filterVectorDistanceConds(bbs []Bb) []Bb {
	result := []Bb{}
	for _, bb := range bbs {
		if bb.op == VECTOR_DISTANCE_FILTER {
			result = append(result, bb)
		}
	}
	return result
}

// 辅助函数：构建条件 SQL（复用现有逻辑）
func buildConditionSql(bbs []Bb) (string, []interface{}) {
	// TODO: 这里需要复用 to_sql.go 中的 buildCondSql 逻辑
	// 暂时简化实现
	if len(bbs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	var args []interface{}

	for i, bb := range bbs {
		if i > 0 {
			sb.WriteString(" AND ")
		}

		// 简化处理：只处理基本运算符
		sb.WriteString(bb.key)
		sb.WriteString(" ")
		sb.WriteString(bb.op)
		sb.WriteString(" ?")

		args = append(args, bb.value)
	}

	return sb.String(), args
}

// 辅助函数：构建向量距离过滤 SQL
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

		params := bb.value.(VectorDistanceFilterParams)

		// (field <-> ?) op threshold
		sb.WriteString("(")
		sb.WriteString(bb.key)
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
