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

	"github.com/x-ream/xb/internal"
)

func (built *Built) toAggSql(vs *[]interface{}, bp *strings.Builder) {

	if len(built.Aggs) == 0 {
		return
	}

	for _, bb := range built.Aggs {
		bp.WriteString(internal.SPACE)
		bp.WriteString(bb.key)
		if bb.value != nil && vs != nil {
			for _, v := range bb.value.([]interface{}) {
				*vs = append(*vs, v)
			}
		}
	}

}

func (built *Built) toAggSqlOfCount(bp *strings.Builder) {
	built.toAggSql(nil, bp)
}
