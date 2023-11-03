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
	"strings"
)

func (built *Built) toSourceSqlBySql(bp *strings.Builder) bool {
	if (len(built.Sbs) == 1) && (built.OrSourceSql != "") {
		var sql = strings.Trim(built.OrSourceSql, SPACE)
		if strings.HasPrefix(sql, "FROM") {
			sql = strings.Replace(sql, "FROM ", "", 1)
		}
		bp.WriteString(sql)
		return true
	}
	return false
}

func (built *Built) toSourceSqlByBuilder(vs *[]interface{}, sb *SourceX, bp *strings.Builder) {
	if sb.join != nil { //JOIN
		bp.WriteString(SPACE)
		bp.WriteString(sb.join.join)
		bp.WriteString(SPACE)
	}
	if sb.po != nil {
		bp.WriteString(sb.po.TableName())
	} else if sb.sub != nil {
		dataSql, _ := sb.sub.Build().sqlData(vs, nil)
		bp.WriteString(BEGIN_SUB)
		bp.WriteString(dataSql)
		bp.WriteString(END_SUB)
	}
	if sb.alia != "" {
		bp.WriteString(SPACE)
		bp.WriteString(sb.alia)
	}
	if sb.join != nil && sb.join.on != nil { //ON

		if sb.join.on.orUsingKey != "" {
			bp.WriteString(USING_SCRIPT_LEFT)
			bp.WriteString(sb.join.on.orUsingKey)
			bp.WriteString(END_SUB)
		} else if sb.s != "" {
			bp.WriteString(SPACE)
			bp.WriteString(sb.s)
		} else {
			bp.WriteString(ON_SCRIPT)
			built.toConditionSql(sb.join.on.bbs, bp, vs, nil)
		}
	}
}
