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
package xb

import (
	"strings"

	. "github.com/x-ream/xb/internal"
)

func (built *Built) toFromSqlBySql(bp *strings.Builder) bool {
	if (len(built.Fxs) == 0) && (built.OrFromSql != "") {
		var sql = strings.Trim(built.OrFromSql, SPACE)
		if strings.HasPrefix(sql, "FROM") {
			sql = strings.Replace(sql, "FROM ", "", 1)
		}
		bp.WriteString(sql)
		return true
	}
	return false
}

func (built *Built) toFromSqlByBuilder(vs *[]interface{}, sx *FromX, bp *strings.Builder) {
	if sx.join != nil { //Join
		bp.WriteString(SPACE)
		bp.WriteString(sx.join.join)
		bp.WriteString(SPACE)
	}
	if sx.tableName != "" {
		bp.WriteString(sx.tableName)
	} else if sx.sub != nil {
		dataSql, _ := sx.sub.Build().sqlData(vs, nil)
		bp.WriteString(BEGIN_SUB)
		bp.WriteString(dataSql)
		bp.WriteString(END_SUB)
	}
	if sx.alia != "" {
		bp.WriteString(SPACE)
		bp.WriteString(sx.alia)
	}
	if sx.join != nil && sx.join.on != nil { //ON

		if sx.join.on.orUsingKey != "" {
			bp.WriteString(USING_SCRIPT_LEFT)
			bp.WriteString(sx.join.on.orUsingKey)
			bp.WriteString(END_SUB)
		} else {
			bp.WriteString(ON_SCRIPT)
			built.toCondSql(sx.join.on.bbs, bp, vs, nil)
		}
	}
}
