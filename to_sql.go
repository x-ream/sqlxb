/*
 * Copyright 2020 io.xream.sqlxb
 *
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package sqlxb

import (
	"strings"
	. "github.com/x-ream/sqlxb/internal"
)

type Built struct {
	ResultKeys *[]string
	ConditionX *[]*Bb
	Sorts      *[]*Sort
	Havings    *[]*Bb
	GroupBys   *[]string

	Page               uint
	Rows               uint
	Last               uint64
	IsTotalRowsIgnored bool

	Po Po
}

func (builder *ConditionBuilder) Build() *Built  {
	if builder == nil {
		panic("sqlxb.Builder is nil")
	}
	built := Built{
		ConditionX: builder.bbs,
	}

	return &built
}

func (builder *Builder) Build() *Built {

	if builder == nil {
		panic("sqlxb.Builder is nil")
	}

	built := Built{
		ResultKeys: nil,
		ConditionX: builder.bbs,
		Sorts:      &builder.sorts,
		Havings:    &builder.havings,
		GroupBys:   &builder.groupBys,

		Po: builder.po,
	}
	if builder.pageBuilder != nil {
		built.Page = builder.pageBuilder.page
		built.Rows = builder.pageBuilder.rows
		built.Last = builder.pageBuilder.last
		built.IsTotalRowsIgnored = builder.pageBuilder.isTotalRowsIgnored
	}

	return &built
}

func (built *Built) toResultKeyScript(bp *strings.Builder) {
	//sb := (*bp)
	bp.WriteString(SELECT)
	if built.ResultKeys == nil {
		bp.WriteString(STAR)
	} else {
		length := len(*built.ResultKeys)
		if length == 0 {
			bp.WriteString(STAR)
		} else {
			for i := 0; i < length; i++ {
				kp := (*built.ResultKeys)[i]
				bp.WriteString(kp)
				if i < length-1 {
					bp.WriteString(COMMA)
				} else {
					bp.WriteString(SPACE)
				}
			}
		}
	}
}

func (built *Built) toResultKeyScriptOfCount(bpCount *strings.Builder) {
	if bpCount == nil {
		return
	}
	bpCount.WriteString("SELECT COUNT(*) ")
}

func (built *Built) toSourceScriptOfCount(bpCount *strings.Builder) {
	if bpCount == nil {
		return
	}
	built.toSourceScript(bpCount)
}

func (built *Built) toConditionScriptOfCount(bbs *[]*Bb, bpCount *strings.Builder) {
	if bpCount == nil {
		return
	}
	built.toConditionScript(bbs, bpCount)
}

func (built *Built) toGroupBySqlOfCount(bys *[]string, bpCount *strings.Builder) {
	if bpCount == nil {
		return
	}
	built.toGroupBySql(bys,bpCount)
}

func (built *Built) toSourceScript(bp *strings.Builder) {
	if built.Po == nil {
		bp.WriteString("?")
	}else {
		bp.WriteString(built.Po.TableName())
	}

	length := len(*built.ConditionX)
	if length == 0 {
		return
	}
}

func (built *Built) toBb(bb *Bb, bp *strings.Builder)  {
	op := bb.op
	switch op {
	case X:
		bp.WriteString(bb.key)
	case IN, NIN:
		bp.WriteString(bb.key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.op)
		bp.WriteString(SPACE)
		bp.WriteString(BEGIN_SUB)
		for _,s := range *(bb.value.(*[]string)) {
			bp.WriteString(s)
		}
		bp.WriteString(END_SUB)
	case IS_NULL, NON_NULL:
		bp.WriteString(bb.key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.op)
	case AND, OR:
		if bb.subs == nil || len(*bb.subs) == 0 {
			return
		}
		bp.WriteString(BEGIN_SUB)
		built.toConditionScript(bb.subs, bp)
		bp.WriteString(END_SUB)
	default:
		bp.WriteString(bb.key)
		bp.WriteString(SPACE)
		bp.WriteString(bb.op)
		bp.WriteString(PLACE_HOLDER)
	}
}

func (built *Built) toConditionScript(bbs *[]*Bb, bp *strings.Builder) {

	if bbs == nil {
		return
	}
	length := len(*bbs)
	if length == 0 {
		return
	}
	for i := 0; i < length; i++ {
		bb := (*bbs)[i]
		built.toBb(bb, bp)
		if i < length-1 {
			next := (*bbs)[i+1]
			if built.isOr(next) {
				if built.isOR(next){
					if i + 1 < length -1 {
						nextNext := (*bbs)[i+2]
						if !built.isOR(nextNext) {
							bp.WriteString(OR_SCRIPT)
						}
						i++
					}
				}else {
					bp.WriteString(OR_SCRIPT)
				}
			} else {
				bp.WriteString(AND_SCRIPT)
			}
		}
	}

}

func (built *Built) toGroupBySql(bys *[]string, bp *strings.Builder) {
	length := len(*bys)
	if bys == nil || length == 0 {
		return
	}
	bp.WriteString(GROUP_BY)
	for i := 0; i<length; i++ {
		bp.WriteString((*bys)[i])
		if i < length - 1 {
			bp.WriteString(COMMA)
		}
	}
}

func (built *Built) toHavingSql(bys *[]*Bb, bp *strings.Builder) {
	if bys == nil || len(*bys) == 0 {
		return
	}
	bp.WriteString(HAVING)
	built.toConditionScript(bys,bp)
}

func (built *Built) toSortSql(bbs *[]*Sort, bp *strings.Builder) {
	length := len(*bbs)
	if length == 0 {
		return
	}
	bp.WriteString(ORDER_BY)
	for i := 0; i < length; i++  {
		sort := (*bbs)[i]
		bp.WriteString(sort.orderBy)
		bp.WriteString(sort.direction)
		if i < length -1 {
			bp.WriteString(COMMA)
		}
	}
}

func (built *Built) isOr(bb *Bb) bool {
	return bb.op == OR
}

func (built *Built) isOR(bb *Bb) bool {
	return bb.op == OR && bb.key == ""
}

func (built *Built) countBuilder() *strings.Builder  {
	var sbCount *strings.Builder
	if built.Rows > 1 && !built.IsTotalRowsIgnored {
		sbCount = &strings.Builder{}
	}
	return sbCount
}

func (built *Built) SqlOfCondition() (*[]interface{}, *string, error) {
	sb := strings.Builder{}
	built.toConditionScript(built.ConditionX, &sb)
	conditionSql := sb.String()
	return nil, &conditionSql, nil
}

func (built *Built) sqlCount(sbCount *strings.Builder) *string{
	if sbCount == nil {
		return nil
	}
	countSql := sbCount.String()
	return &countSql
}

func (built *Built) Sql() (*[]interface{}, *string, *string, error) {

	sb := strings.Builder{}
	var sbCount = built.countBuilder()
	built.toResultKeyScript(&sb)
	built.toResultKeyScriptOfCount(sbCount)
	sb.WriteString(FROM)
	built.toSourceScript(&sb)
	built.toSourceScriptOfCount(sbCount)
	sb.WriteString(WHERE)
	built.toConditionScript(built.ConditionX, &sb)
	built.toConditionScriptOfCount(built.ConditionX,sbCount)
	built.toGroupBySql(built.GroupBys,&sb)
	built.toGroupBySqlOfCount(built.GroupBys,sbCount)
	built.toHavingSql(built.Havings,&sb)
	built.toSortSql(built.Sorts, &sb)
	dataSql := sb.String()
	countSql := built.sqlCount(sbCount)
	return nil, &dataSql,countSql,nil
}


