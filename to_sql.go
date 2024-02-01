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

import (
	. "github.com/x-ream/sqlxb/internal"
	"strconv"
	"strings"
)

type Built struct {
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
}

type Qr struct {
	Rs []Bb
	Cs []Bb

	Svs []interface{}

	Po Po
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
	op := bb.op
	switch op {
	case XX:
		bp.WriteString(bb.key)
		if vs != nil && bb.value != nil {
			arr := bb.value.([]interface{})
			for _, v := range arr {
				*vs = append(*vs, v)
			}
		}
	case IN, NIN:
		bp.WriteString(bb.key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.op)
		bp.WriteString(SPACE)
		bp.WriteString(BEGIN_SUB)
		arr := *(bb.value.(*[]string))
		inl := len(arr)
		for i := 0; i < inl; i++ {
			bp.WriteString(arr[i])
			if i < inl-1 {
				bp.WriteString(COMMA)
			}
		}
		bp.WriteString(END_SUB)
	case IS_NULL, NON_NULL:
		bp.WriteString(bb.key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.op)
	case AND, OR:
		if bb.subs == nil || len(bb.subs) == 0 {
			return
		}
		bp.WriteString(BEGIN_SUB)
		built.toCondSql(bb.subs, bp, vs, nil)
		bp.WriteString(END_SUB)
	case SUB:
		var bx = *bb.value.(*BuilderX)
		ss, _ := bx.Build().sqlData(vs, nil)
		ss = BEGIN_SUB + ss + END_SUB
		ss = SPACE + ss
		if bb.key != "" {
			if strings.Contains(bb.key, PLACE_HOLDER) {
				bp.WriteString(strings.ReplaceAll(bb.key, PLACE_HOLDER, ss))
			} else {
				bp.WriteString(bb.key)
				bp.WriteString(SPACE)
				bp.WriteString(ss)
			}
		} else {
			bp.WriteString(ss)
		}
	default:
		bp.WriteString(bb.key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.op)
		bp.WriteString(PLACE_HOLDER)
		if vs != nil {
			*vs = append(*vs, bb.value)
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
					if i+1 < length-1 {
						nextNext := bbs[nextIdx+1]
						if !built.isOR(nextNext) {
							bp.WriteString(OR_SCRIPT)
						}
						i++
					}
				} else {
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
	if built.PageCondition == nil || built.PageCondition.rows < 1 {
		return
	}
	bp.WriteString(LIMIT)
	bp.WriteString(strconv.Itoa(int(built.PageCondition.rows)))
	if built.PageCondition.last < 1 {
		if built.PageCondition.page < 1 {
			built.PageCondition.page = 1
		}
		bp.WriteString(OFFSET)
		bp.WriteString(strconv.Itoa(int((built.PageCondition.page - 1) * built.PageCondition.rows)))
	}
}

func (built *Built) toLastSql(bp *strings.Builder) {
	if built.Last != "" {
		bp.WriteString(SPACE)
		bp.WriteString(built.Last)
	}
}

func (built *Built) isOr(bb Bb) bool {
	return bb.op == OR
}

func (built *Built) isOR(bb Bb) bool {
	return bb.op == OR && bb.key == ""
}

func (built *Built) countBuilder() *strings.Builder {
	var sbCount *strings.Builder
	pageCondition := built.PageCondition
	if pageCondition != nil && pageCondition.rows > 1 && !pageCondition.isTotalRowsIgnored {
		sbCount = &strings.Builder{}
	}
	return sbCount
}

func (built *Built) SqlOfCond() (string, string, []interface{}) {
	vs := []interface{}{}

	joinB := strings.Builder{}
	if built.Fxs != nil {
		for i, fx := range built.Fxs {
			if i > 0 {
				built.toFromSqlByBuilder(&vs, fx, &joinB)
			}
		}
	}

	condB := strings.Builder{}
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
	bp.WriteString(FROM)
}

func (built *Built) sqlWhere(bp *strings.Builder) {
	if len(built.Conds) == 0 {
		return
	}
	bp.WriteString(WHERE)
}

func (built *Built) Sql() ([]interface{}, string, string, map[string]string) {
	vs := []interface{}{}
	km := make(map[string]string) //nil for sub From builder,
	dataSql, kmp := built.sqlData(&vs, km)
	countSql := built.sqlCount()

	return vs, dataSql, countSql, kmp
}

func (built *Built) sqlData(vs *[]interface{}, km map[string]string) (string, map[string]string) {
	sb := strings.Builder{}
	built.toResultKeySql(&sb, km)
	built.sqlFrom(&sb)
	built.toFromSql(vs, &sb)
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

func (built *Built) sqlCount() string {
	sbCount := built.countBuilder()
	if sbCount == nil {
		return ""
	}
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
