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
	"strconv"
	"strings"

	"github.com/fndome/xb/interceptor"
	. "github.com/fndome/xb/internal"
)

type Built struct {
	Delete     bool
	Inserts    *[]Bb
	Updates    *[]Bb
	ResultKeys []string
	Conds      []Bb
	Sorts      []Sort
	Havings    []Bb
	GroupBys   []string
	Aggs       []Bb
	Last       string
	OrFromSql  string
	Fxs        []*FromX
	Svs        []interface{}

	PageCondition *PageCondition

	// ⭐ Database-specific configuration (Dialect + Custom)
	// If nil, defaults to SQL dialect
	Custom      Custom
	LimitValue  int                   // ⭐ LIMIT value (v0.10.1)
	OffsetValue int                   // ⭐ OFFSET value (v0.10.1)
	Meta        *interceptor.Metadata // ⭐ Metadata (v0.9.2)
	Alia        string
	Withs       []WithClause
	Unions      []UnionClause
}

// WithClause common table expression (CTE) definition
type WithClause struct {
	Name      string
	SQL       string
	Args      []interface{}
	Recursive bool
}

// UnionClause UNION definition
type UnionClause struct {
	Operator string
	SQL      string
	Args     []interface{}
}

// ============================================================================
// Unified Query Generation Interface (v0.11.0)
// ============================================================================

// JsonOfSelect generates query JSON (unified interface)
// ⭐ Automatically selects database dialect based on Built.Custom (Qdrant/Milvus/Weaviate, etc.)
//
// Returns:
//   - JSON string
//   - error
//
// Example:
//
//	// Qdrant
//	built := xb.Of("code_vectors").
//	    Custom(xb.NewQdrantBuilder().Build()).
//	    VectorSearch(...).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ⭐ Automatically uses Qdrant
//
//	// Milvus (example: future implementation using Builder pattern)
//	// built := xb.Of("users").
//	//     Custom(xb.NewMilvusBuilder().Build()).
//	//     VectorSearch(...).
//	//     Build()
//	//
//	// json, _ := built.JsonOfSelect()  // ⭐ Automatically uses Milvus
func (built *Built) JsonOfSelect() (string, error) {
	if built.Custom == nil {
		return "", fmt.Errorf("Custom is nil, use SqlOfSelect() for SQL databases")
	}

	// ⭐ Call Custom.Generate()
	result, err := built.Custom.Generate(built)
	if err != nil {
		return "", err
	}

	// ⭐ Type assertion: expect string (JSON)
	if jsonStr, ok := result.(string); ok {
		return jsonStr, nil
	}

	// If SQLResult, convert to JSON (optional)
	if sqlResult, ok := result.(*SQLResult); ok {
		return "", fmt.Errorf("got SQL result, use SqlOfSelect() instead. SQL: %s", sqlResult.SQL)
	}

	return "", fmt.Errorf("unexpected result type: %T", result)
}

// JsonOfInsert generates insert JSON (vector databases)
// ⭐ Used for insert operations in vector databases like Qdrant/Milvus
func (built *Built) JsonOfInsert() (string, error) {
	if built.Custom == nil {
		return "", fmt.Errorf("Custom is nil, use SqlOfInsert() for SQL databases")
	}

	result, err := built.Custom.Generate(built)
	if err != nil {
		return "", err
	}

	if jsonStr, ok := result.(string); ok {
		return jsonStr, nil
	}

	return "", fmt.Errorf("unexpected result type: %T", result)
}

// JsonOfUpdate generates update JSON (vector databases)
// ⭐ Used for update operations in vector databases like Qdrant/Milvus
func (built *Built) JsonOfUpdate() (string, error) {
	if built.Custom == nil {
		return "", fmt.Errorf("Custom is nil, use SqlOfUpdate() for SQL databases")
	}

	result, err := built.Custom.Generate(built)
	if err != nil {
		return "", err
	}

	if jsonStr, ok := result.(string); ok {
		return jsonStr, nil
	}

	return "", fmt.Errorf("unexpected result type: %T", result)
}

// JsonOfDelete generates delete JSON (vector databases)
// ⭐ Used for delete operations in vector databases like Qdrant/Milvus
func (built *Built) JsonOfDelete() (string, error) {
	if built.Custom == nil {
		return "", fmt.Errorf("Custom is nil, use SqlOfDelete() for SQL databases")
	}

	// ⭐ Automatically set Delete flag
	built.Delete = true

	result, err := built.Custom.Generate(built)
	if err != nil {
		return "", err
	}

	if jsonStr, ok := result.(string); ok {
		return jsonStr, nil
	}

	return "", fmt.Errorf("unexpected result type: %T", result)
}

func (built *Built) toFromSqlOfCount(bpCount *strings.Builder) {
	built.toFromSql(nil, bpCount)
}

func (built *Built) toCondSqlOfCount(bbs []Bb, bpCount *strings.Builder) {
	built.toCondSql(bbs, bpCount, nil, nil)
}

func (built *Built) toGroupBySqlOfCount(bpCount *strings.Builder) {
	built.toGroupBySql(bpCount)
}

func (built *Built) toFromSql(vs *[]interface{}, bp *strings.Builder) {
	if built.OrFromSql == "" {

		if built.toFromSqlBySql(bp) {
			return
		}

		for _, sb := range built.Fxs {
			built.toFromSqlByBuilder(vs, sb, bp)
		}
	} else {
		bp.WriteString(built.OrFromSql)
	}
}

func (built *Built) toBb(bb Bb, bp *strings.Builder, vs *[]interface{}) {
	op := bb.Op
	switch op {
	case XX:
		bp.WriteString(bb.Key)
		if vs != nil && bb.Value != nil {
			arr := bb.Value.([]interface{})
			for _, v := range arr {
				*vs = append(*vs, v)
			}
		}
	case IN, NIN:
		bp.WriteString(bb.Key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.Op)
		bp.WriteString(SPACE)
		bp.WriteString(BEGIN_SUB)
		arr := *(bb.Value.(*[]string))
		inl := len(arr)
		for i := 0; i < inl; i++ {
			bp.WriteString(arr[i])
			if i < inl-1 {
				bp.WriteString(COMMA)
			}
		}
		bp.WriteString(END_SUB)
	case IS_NULL, NON_NULL:
		bp.WriteString(bb.Key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.Op)
	case AND, OR:
		if bb.Subs == nil || len(bb.Subs) == 0 {
			return
		}
		bp.WriteString(BEGIN_SUB)
		built.toCondSql(bb.Subs, bp, vs, nil)
		bp.WriteString(END_SUB)
	case SUB:
		var bx = *bb.Value.(*BuilderX)
		ss, _ := bx.Build().SqlData(vs, nil)
		ss = BEGIN_SUB + ss + END_SUB
		ss = SPACE + ss
		if bb.Key != "" {
			if strings.Contains(bb.Key, PLACE_HOLDER) {
				bp.WriteString(strings.ReplaceAll(bb.Key, PLACE_HOLDER, ss))
			} else {
				bp.WriteString(bb.Key)
				bp.WriteString(SPACE)
				bp.WriteString(ss)
			}
		} else {
			bp.WriteString(ss)
		}
	default:
		bp.WriteString(bb.Key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.Op)
		bp.WriteString(PLACE_HOLDER)
		if vs != nil {
			*vs = append(*vs, bb.Value)
		}
	}
}

func (built *Built) toCondSql(bbs []Bb, bp *strings.Builder, vs *[]interface{}, filterLast func() *Bb) {

	length := len(bbs)

	if filterLast != nil {
		if bb := filterLast(); bb != nil {
			built.toBb(*bb, bp, vs)
			if length > 0 {
				bp.WriteString(AND_SCRIPT)
			}
		}
	}

	if length == 0 {
		return
	}

	for i := 0; i < length; i++ {
		bb := bbs[i]
		built.toBb(bb, bp, vs)
		if i < length-1 {
			nextIdx := i + 1
			next := bbs[nextIdx]
			if built.isOr(next) {
				if built.isOR(next) {
					// next is a pure OR operator (created by OR() method)
					if i+1 < length-1 {
						nextNext := bbs[nextIdx+1]
						if !built.isOR(nextNext) {
							bp.WriteString(OR_SCRIPT)
						}
						i++
					}
				} else if len(next.Subs) > 0 {
					// next is OR_SUB (has subs), use AND to connect
					bp.WriteString(AND_SCRIPT)
				} else {
					// Other OR cases (theoretically should not happen)
					bp.WriteString(OR_SCRIPT)
				}
			} else {
				bp.WriteString(AND_SCRIPT)
			}
		}
	}

}

func (built *Built) toGroupBySql(bp *strings.Builder) {
	if built.GroupBys == nil {
		return
	}
	length := len(built.GroupBys)
	if length == 0 {
		return
	}
	bp.WriteString(GROUP_BY)
	for i := 0; i < length; i++ {
		bp.WriteString(built.GroupBys[i])
		if i < length-1 {
			bp.WriteString(COMMA)
		}
	}
}

func (built *Built) toHavingSql(vs *[]interface{}, bp *strings.Builder) {
	if built.Havings == nil || len(built.Havings) == 0 {
		return
	}
	bp.WriteString(HAVING)
	built.toCondSql(built.Havings, bp, vs, nil)
}

func (built *Built) toHavingSqlOfCount(bp *strings.Builder) {
	built.toHavingSql(nil, bp)
}

func (built *Built) toSortSql(bp *strings.Builder) {
	if built.Sorts == nil {
		return
	}
	length := len(built.Sorts)
	if length == 0 {
		return
	}
	bp.WriteString(ORDER_BY)
	for i := 0; i < length; i++ {
		sort := built.Sorts[i]
		bp.WriteString(sort.orderBy)
		if sort.direction != "" {
			bp.WriteString(SPACE)
			bp.WriteString(sort.direction)
		}
		if i < length-1 {
			bp.WriteString(COMMA)
		}
	}
}

func (built *Built) toPageSql(bp *strings.Builder) {
	// ⭐ Prefer Paged() (web pagination, supports COUNT + Last optimization)
	// If PageCondition exists, ignore Limit/Offset
	if built.PageCondition != nil {
		if built.PageCondition.Rows >= 1 {
			bp.WriteString(LIMIT)
			bp.WriteString(strconv.Itoa(int(built.PageCondition.Rows)))
			if built.PageCondition.Last < 1 {
				if built.PageCondition.Page < 1 {
					built.PageCondition.Page = 1
				}
				bp.WriteString(OFFSET)
				bp.WriteString(strconv.Itoa(int((built.PageCondition.Page - 1) * built.PageCondition.Rows)))
			}
		}
		return // ⭐ Return directly, ignore Limit/Offset
	}

	// ⭐ Only use Limit/Offset when Paged() is not used (simple queries)
	if built.LimitValue > 0 {
		bp.WriteString(LIMIT)
		bp.WriteString(strconv.Itoa(built.LimitValue))
	}

	if built.OffsetValue > 0 {
		bp.WriteString(OFFSET)
		bp.WriteString(strconv.Itoa(built.OffsetValue))
	}
}

func (built *Built) toLastSql(bp *strings.Builder) {
	if built.Last != "" {
		bp.WriteString(SPACE)
		bp.WriteString(built.Last)
	}
}

func (built *Built) isOr(bb Bb) bool {
	return bb.Op == OR
}

func (built *Built) isOR(bb Bb) bool {
	return bb.Op == OR && bb.Key == ""
}

func (built *Built) countBuilder() *strings.Builder {
	var sbCount *strings.Builder
	pageCondition := built.PageCondition
	if pageCondition != nil && pageCondition.Rows > 1 && !pageCondition.IsTotalRowsIgnored {
		sb := strings.Builder{}
		sb.Grow(128) // Pre-allocate 128 bytes
		sbCount = &sb
	}
	return sbCount
}

func (built *Built) SqlOfPage() (string, string, []interface{}, map[string]string) {
	// ⭐ If Custom is set, try to get from Custom
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				// ⭐ Prefer CountSQL provided by Custom
				countSQL := sqlResult.CountSQL
				if countSQL == "" {
					// ⭐ If Custom didn't provide, use default generation
					countSQL = built.SqlCount()
				}

				meta := sqlResult.Meta
				if meta == nil {
					meta = make(map[string]string)
				}
				return countSQL, sqlResult.SQL, sqlResult.Args, meta
			}
		}
	}

	// ⭐ Default implementation
	vs := []interface{}{}
	km := make(map[string]string)
	dataSql, kmp := built.SqlData(&vs, km)
	countSQL := built.SqlCount()

	return countSQL, dataSql, vs, kmp
}

func (built *Built) SqlOfSelect() (string, []interface{}, map[string]string) {
	// ⭐ If Custom is set, try to get from Custom
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			// ⭐ Type assertion: expect *SQLResult
			if sqlResult, ok := result.(*SQLResult); ok {
				meta := sqlResult.Meta
				if meta == nil {
					meta = make(map[string]string)
				}
				return sqlResult.SQL, sqlResult.Args, meta
			}
		}
		// If Custom didn't return SQLResult, continue with default implementation
	}

	// ⭐ Default implementation (original logic)
	vs := []interface{}{}
	km := make(map[string]string)
	dataSql, kmp := built.SqlData(&vs, km)
	return dataSql, vs, kmp
}

func (built *Built) SqlOfInsert() (string, []interface{}) {
	// ⭐ If Custom is set, try to get from Custom
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				return sqlResult.SQL, sqlResult.Args
			}
		}
	}

	// ⭐ Default implementation
	vs := []interface{}{}
	sql := built.SqlInsert(&vs)
	return sql, vs
}

func (built *Built) SqlOfUpdate() (string, []interface{}) {
	// ⭐ If Custom is set, try to get from Custom
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				return sqlResult.SQL, sqlResult.Args
			}
		}
	}

	// ⭐ Default implementation
	vs := []interface{}{}
	km := make(map[string]string)
	dataSql, _ := built.SqlData(&vs, km)
	return dataSql, vs
}

func (built *Built) SqlOfDelete() (string, []interface{}) {
	// ⭐ If Custom is set, try to get from Custom
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				return sqlResult.SQL, sqlResult.Args
			}
		}
	}

	// ⭐ Default implementation
	vs := []interface{}{}
	sql := built.sqlDelete(&vs)
	return sql, vs
}

func (built *Built) SqlOfCond() (string, string, []interface{}) {
	vs := []interface{}{}

	joinB := strings.Builder{}
	joinB.Grow(128) // Pre-allocate 128 bytes for JOIN statements
	if built.Fxs != nil {
		for i, fx := range built.Fxs {
			if i > 0 {
				built.toFromSqlByBuilder(&vs, fx, &joinB)
			}
		}
	}

	condB := strings.Builder{}
	condB.Grow(256) // Pre-allocate 256 bytes for WHERE conditions
	built.toCondSql(built.Conds, &condB, &vs, built.filterLast)

	return joinB.String(), condB.String(), vs
}

func (built *Built) toSqlCount(sbCount *strings.Builder) string {
	if sbCount == nil {
		return ""
	}
	countSql := sbCount.String()
	return countSql
}

func (built *Built) countSqlFrom(sbCount *strings.Builder) {
	sbCount.WriteString(FROM)
}

func (built *Built) countSqlWhere(sbCount *strings.Builder) {
	built.sqlWhere(sbCount)
}

func (built *Built) sqlFrom(bp *strings.Builder) {
	if built.Updates == nil {
		bp.WriteString(FROM)
	}
}

func (built *Built) sqlWhere(bp *strings.Builder) {
	if len(built.Conds) == 0 {
		return
	}
	bp.WriteString(WHERE)
}

func (built *Built) toDelete(bp *strings.Builder) {
	bp.WriteString(DELETE)
}

func (built *Built) sqlDelete(vs *[]interface{}) string {
	sb := strings.Builder{}
	sb.Grow(128) // Pre-allocate 128 bytes to reduce memory reallocation
	built.toDelete(&sb)
	built.sqlFrom(&sb)
	built.toFromSql(vs, &sb)
	built.toUpdateSql(&sb, vs)
	built.sqlWhere(&sb)
	built.toCondSql(built.Conds, &sb, vs, built.filterLast)
	built.toAggSql(vs, &sb)
	built.toGroupBySql(&sb)
	built.toHavingSql(vs, &sb)
	built.toSortSql(&sb)
	built.toPageSql(&sb)
	built.toLastSql(&sb)
	deleteSql := sb.String()
	return deleteSql
}

// SqlData generates data query SQL (SELECT or UPDATE)
//
// Notes:
//   - Generates complete SELECT or UPDATE SQL (does not include pagination COUNT SQL)
//   - Custom implementations can call this method to generate base SQL
//
// Parameters:
//   - vs: parameter list (pointer)
//   - km: metadata map
//
// Returns:
//   - string: SQL statement
//   - map[string]string: metadata
//
// Example:
//
//	vs := []interface{}{}
//	km := make(map[string]string)
//	sql, meta := built.SqlData(&vs, km)
//	// SELECT * FROM users WHERE age > ?
func (built *Built) SqlData(vs *[]interface{}, km map[string]string) (string, map[string]string) {
	sb := strings.Builder{}
	sb.Grow(256) // Pre-allocate 256 bytes, SELECT statements are usually longer
	built.appendWithClauses(&sb, vs)
	built.writeSelectCore(&sb, vs, km)
	built.appendUnionClauses(&sb, vs)
	built.toSortSql(&sb)
	built.toPageSql(&sb)
	built.toLastSql(&sb)
	dataSql := sb.String()
	return dataSql, km
}

// SqlCount generates COUNT SQL (for pagination)
//
// Notes:
//   - Used to generate COUNT(*) SQL, usually used with SqlData for pagination
//   - Custom implementations can call this method to generate count SQL
//
// Returns:
//   - string: COUNT SQL
//
// Example:
//
//	countSQL := built.SqlCount()
//	// SELECT COUNT(*) FROM users WHERE age > ?
func (built *Built) SqlCount() string {
	sbCount := built.countBuilder()
	if sbCount == nil {
		return ""
	}
	sbCount.Grow(128) // Pre-allocate 128 bytes, COUNT statements are relatively short
	built.appendWithClauses(sbCount, nil)
	built.toResultKeySqlOfCount(sbCount)
	built.countSqlFrom(sbCount)
	built.toFromSqlOfCount(sbCount)
	built.countSqlWhere(sbCount)
	built.toCondSqlOfCount(built.Conds, sbCount)
	built.toAggSqlOfCount(sbCount)
	built.toGroupBySqlOfCount(sbCount)
	built.toHavingSqlOfCount(sbCount)
	countSql := built.toSqlCount(sbCount)
	return countSql
}

func (built *Built) appendWithClauses(sb *strings.Builder, vs *[]interface{}) {
	if len(built.Withs) == 0 {
		return
	}

	hasRecursive := false
	for _, clause := range built.Withs {
		if clause.Recursive {
			hasRecursive = true
			break
		}
	}

	sb.WriteString("WITH ")
	if hasRecursive {
		sb.WriteString("RECURSIVE ")
	}

	for i, clause := range built.Withs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(clause.Name)
		sb.WriteString(" AS (")
		sb.WriteString(clause.SQL)
		sb.WriteString(")")
		if vs != nil && len(clause.Args) > 0 {
			*vs = append(*vs, clause.Args...)
		}
	}
	sb.WriteString(" ")
}

func (built *Built) appendUnionClauses(sb *strings.Builder, vs *[]interface{}) {
	if len(built.Unions) == 0 {
		return
	}
	for _, clause := range built.Unions {
		if clause.Operator == "" || clause.SQL == "" {
			continue
		}
		sb.WriteString(" ")
		sb.WriteString(clause.Operator)
		sb.WriteString(" (")
		sb.WriteString(clause.SQL)
		sb.WriteString(")")
		if vs != nil && len(clause.Args) > 0 {
			*vs = append(*vs, clause.Args...)
		}
	}
}

func (built *Built) writeSelectCore(sb *strings.Builder, vs *[]interface{}, km map[string]string) {
	built.toResultKeySql(sb, km)
	built.sqlFrom(sb)
	built.toFromSql(vs, sb)
	built.toUpdateSql(sb, vs)
	built.sqlWhere(sb)
	built.toCondSql(built.Conds, sb, vs, built.filterLast)
	built.toAggSql(vs, sb)
	built.toGroupBySql(sb)
	built.toHavingSql(vs, sb)
}
