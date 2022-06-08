// Copyright 2020 io.xream.sqlxb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

func (built *Built) toSourceScriptByBuilder(sb *SourceBuilder, bp *strings.Builder, vsp *[]interface{}) {
	if sb.join != nil { //JOIN
		bp.WriteString(SPACE)
		bp.WriteString(sb.join.join)
		bp.WriteString(SPACE)
	}
	if sb.po != nil {
		bp.WriteString(sb.po.TableName())
	} else if sb.sub != nil {
		vs, dataSql := sb.sub.Build().sqlData()
		if vsp != nil {
			for _, v := range *vs {
				*vsp = append(*vsp, v)
			}
		}
		bp.WriteString(BEGIN_SUB)
		bp.WriteString(*dataSql)
		bp.WriteString(END_SUB)
	}
	if sb.alia != "" {
		bp.WriteString(SPACE)
		bp.WriteString(sb.alia)
	}
	if sb.join != nil && sb.join.on != nil { //ON
		bp.WriteString(ON_SCRIPT)
		if sb.alia != "" {
			bp.WriteString(sb.alia)
		} else {
			bp.WriteString(sb.po.TableName())
		}
		bp.WriteString(DOT)
		bp.WriteString(sb.join.on.key)
		bp.WriteString(EQ_SCRIPT)
		bp.WriteString(sb.join.on.targetAlia)
		bp.WriteString(DOT)
		bp.WriteString(sb.join.on.targetKey)
	}
}
