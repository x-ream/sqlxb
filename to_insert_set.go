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
	"strings"

	. "github.com/fndome/xb/internal"
)

func (built *Built) sqlInsert(vs *[]interface{}) string {

	bp := strings.Builder{}
	bp.Grow(128) // 预分配 128 字节，INSERT 语句通常不太长
	bp.WriteString(INSERT)
	bp.WriteString(built.OrFromSql)
	bp.WriteString(SPACE)
	bp.WriteString(BEGIN_SUB)
	length := len(*built.Inserts)
	for i := 0; i < length; i++ {
		v := (*built.Inserts)[i]
		bp.WriteString(v.Key)
		if i < length-1 {
			bp.WriteString(COMMA)
		}
		*vs = append(*vs, v.Value)
	}

	bp.WriteString(END_SUB)
	bp.WriteString(VALUES)
	bp.WriteString(BEGIN_SUB)
	for i := 0; i < length; i++ {
		bp.WriteString(PLACE_HOLDER)
		if i < length-1 {
			bp.WriteString(COMMA)
		}
	}
	bp.WriteString(END_SUB)

	return bp.String()
}
