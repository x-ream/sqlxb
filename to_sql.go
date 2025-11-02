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

	// ⭐ 数据库专属配置（Dialect + Custom）
	// 如果为 nil，默认为 SQL 方言
	Custom      Custom
	LimitValue  int                   // ⭐ 新增：LIMIT 值（v0.10.1）
	OffsetValue int                   // ⭐ 新增：OFFSET 值（v0.10.1）
	Meta        *interceptor.Metadata // ⭐ 新增：元数据（v0.9.2）
}

// ============================================================================
// 统一的查询生成接口（v0.11.0）
// ============================================================================

// JsonOfSelect 生成查询 JSON（统一接口）
// ⭐ 根据 Built.Custom 自动选择数据库方言（Qdrant/Milvus/Weaviate 等）
//
// 返回:
//   - JSON 字符串
//   - error
//
// 示例:
//
//	// Qdrant
//	built := xb.Of("code_vectors").
//	    Custom(xb.QdrantHighPrecision()).
//	    VectorSearch(...).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Qdrant
//
//	// Milvus
//	built := xb.Of("users").
//	    Custom(xb.NewMilvusCustom()).
//	    VectorSearch(...).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Milvus
func (built *Built) JsonOfSelect() (string, error) {
	if built.Custom == nil {
		return "", fmt.Errorf("Custom is nil, use SqlOfSelect() for SQL databases")
	}

	// ⭐ 调用 Custom.Generate()
	result, err := built.Custom.Generate(built)
	if err != nil {
		return "", err
	}

	// ⭐ 类型断言：期望是 string（JSON）
	if jsonStr, ok := result.(string); ok {
		return jsonStr, nil
	}

	// 如果是 SQLResult，转换为 JSON（可选）
	if sqlResult, ok := result.(*SQLResult); ok {
		return "", fmt.Errorf("got SQL result, use SqlOfSelect() instead. SQL: %s", sqlResult.SQL)
	}

	return "", fmt.Errorf("unexpected result type: %T", result)
}

// JsonOfInsert 生成插入 JSON（向量数据库）
// ⭐ 用于 Qdrant/Milvus 等向量数据库的插入操作
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

// JsonOfUpdate 生成更新 JSON（向量数据库）
// ⭐ 用于 Qdrant/Milvus 等向量数据库的更新操作
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

// JsonOfDelete 生成删除 JSON（向量数据库）
// ⭐ 用于 Qdrant/Milvus 等向量数据库的删除操作
func (built *Built) JsonOfDelete() (string, error) {
	if built.Custom == nil {
		return "", fmt.Errorf("Custom is nil, use SqlOfDelete() for SQL databases")
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
					// next 是纯 OR 操作符（由 OR() 方法创建）
					if i+1 < length-1 {
						nextNext := bbs[nextIdx+1]
						if !built.isOR(nextNext) {
							bp.WriteString(OR_SCRIPT)
						}
						i++
					}
				} else if len(next.Subs) > 0 {
					// next 是 OR_SUB（有 subs），使用 AND 连接
					bp.WriteString(AND_SCRIPT)
				} else {
					// 其他 OR 情况（理论上不应该发生）
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
	// ⭐ 优先使用 Paged()（Web 分页，支持 COUNT + Last 优化）
	// 如果 PageCondition 存在，忽略 Limit/Offset
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
		return // ⭐ 直接返回，忽略 Limit/Offset
	}

	// ⭐ 只有在没有 Paged() 时，才使用 Limit/Offset（简单查询）
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
		sb.Grow(128) // 预分配 128 字节
		sbCount = &sb
	}
	return sbCount
}

func (built *Built) SqlOfPage() (string, string, []interface{}, map[string]string) {
	// ⭐ 如果设置了 Custom，尝试从 Custom 获取
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				// ⭐ 优先使用 Custom 提供的 CountSQL
				countSQL := sqlResult.CountSQL
				if countSQL == "" {
					// ⭐ 如果 Custom 没提供，使用默认生成
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

	// ⭐ 默认实现
	vs := []interface{}{}
	km := make(map[string]string)
	dataSql, kmp := built.SqlData(&vs, km)
	countSQL := built.SqlCount()

	return countSQL, dataSql, vs, kmp
}

func (built *Built) SqlOfSelect() (string, []interface{}, map[string]string) {
	// ⭐ 如果设置了 Custom，尝试从 Custom 获取
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			// ⭐ 类型断言：期望是 *SQLResult
			if sqlResult, ok := result.(*SQLResult); ok {
				meta := sqlResult.Meta
				if meta == nil {
					meta = make(map[string]string)
				}
				return sqlResult.SQL, sqlResult.Args, meta
			}
		}
		// 如果 Custom 返回的不是 SQLResult，继续使用默认实现
	}

	// ⭐ 默认实现（原有逻辑）
	vs := []interface{}{}
	km := make(map[string]string)
	dataSql, kmp := built.SqlData(&vs, km)
	return dataSql, vs, kmp
}

func (built *Built) SqlOfInsert() (string, []interface{}) {
	// ⭐ 如果设置了 Custom，尝试从 Custom 获取
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				return sqlResult.SQL, sqlResult.Args
			}
		}
	}

	// ⭐ 默认实现
	vs := []interface{}{}
	sql := built.SqlInsert(&vs)
	return sql, vs
}

func (built *Built) SqlOfUpdate() (string, []interface{}) {
	// ⭐ 如果设置了 Custom，尝试从 Custom 获取
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				return sqlResult.SQL, sqlResult.Args
			}
		}
	}

	// ⭐ 默认实现
	vs := []interface{}{}
	km := make(map[string]string)
	dataSql, _ := built.SqlData(&vs, km)
	return dataSql, vs
}

func (built *Built) SqlOfDelete() (string, []interface{}) {
	// ⭐ 如果设置了 Custom，尝试从 Custom 获取
	if built.Custom != nil {
		result, err := built.Custom.Generate(built)
		if err == nil {
			if sqlResult, ok := result.(*SQLResult); ok {
				return sqlResult.SQL, sqlResult.Args
			}
		}
	}

	// ⭐ 默认实现
	vs := []interface{}{}
	sql := built.sqlDelete(&vs)
	return sql, vs
}

func (built *Built) SqlOfCond() (string, string, []interface{}) {
	vs := []interface{}{}

	joinB := strings.Builder{}
	joinB.Grow(128) // 预分配 128 字节用于 JOIN 语句
	if built.Fxs != nil {
		for i, fx := range built.Fxs {
			if i > 0 {
				built.toFromSqlByBuilder(&vs, fx, &joinB)
			}
		}
	}

	condB := strings.Builder{}
	condB.Grow(256) // 预分配 256 字节用于 WHERE 条件
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
	sb.Grow(128) // 预分配 128 字节，减少内存重新分配
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

// SqlData 生成数据查询 SQL（SELECT 或 UPDATE）
//
// 说明：
//   - 生成完整的 SELECT 或 UPDATE SQL（不含分页的 COUNT SQL）
//   - Custom 实现可以调用此方法生成基础 SQL
//
// 参数：
//   - vs: 参数列表（指针）
//   - km: 元数据 map
//
// 返回：
//   - string: SQL 语句
//   - map[string]string: 元数据
//
// 示例：
//
//	vs := []interface{}{}
//	km := make(map[string]string)
//	sql, meta := built.SqlData(&vs, km)
//	// SELECT * FROM users WHERE age > ?
func (built *Built) SqlData(vs *[]interface{}, km map[string]string) (string, map[string]string) {
	sb := strings.Builder{}
	sb.Grow(256) // 预分配 256 字节，SELECT 语句通常较长
	built.toResultKeySql(&sb, km)
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
	dataSql := sb.String()
	return dataSql, km
}

// SqlCount 生成 COUNT SQL（用于分页）
//
// 说明：
//   - 用于生成 COUNT(*) SQL，通常与 SqlData 配合用于分页
//   - Custom 实现可以调用此方法生成 count SQL
//
// 返回：
//   - string: COUNT SQL
//
// 示例：
//
//	countSQL := built.SqlCount()
//	// SELECT COUNT(*) FROM users WHERE age > ?
func (built *Built) SqlCount() string {
	sbCount := built.countBuilder()
	if sbCount == nil {
		return ""
	}
	sbCount.Grow(128) // 预分配 128 字节，COUNT 语句相对较短
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
