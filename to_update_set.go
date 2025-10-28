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
	"strings"

	. "github.com/x-ream/xb/internal"
)

func (built *Built) toUpdateSql(bp *strings.Builder, vs *[]interface{}) {
	if built.Updates == nil {
		return
	}

	bp.WriteString(SET)
	length := len(*built.Updates)

	for i := 0; i < length; i++ {
		u := (*built.Updates)[i]
		bp.WriteString(u.key)
		if !strings.Contains(u.key, EQ) {
			bp.WriteString(SPACE)
			bp.WriteString(EQ)
		}
		if u.value != nil {
			bp.WriteString(PLACE_HOLDER)
			*vs = append(*vs, u.value)
		}
		if i < length-1 {
			bp.WriteString(COMMA)
		} else {
			bp.WriteString(SPACE)
		}
	}
}
