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
	"strconv"
	"strings"
)


func adapterResultKeyAlia(km map[string]string, k string, reg string)  {
	arr := strings.Split(k, reg)
	alia := arr[1]
	if strings.Contains(alia,"`") {
		alia = strings.Replace(alia,"`","",2)
	}
	km[alia] = alia
}

func (built *Built) toResultKeyScript(bp *strings.Builder, km map[string]string) {
	bp.WriteString(SELECT)
	if built.ResultKeys == nil {
		bp.WriteString(STAR)
	} else {
		length := len(built.ResultKeys)
		if length == 0 {
			bp.WriteString(STAR)
		} else {
			for i := 0; i < length; i++ {
				kp := (built.ResultKeys)[i]
				if km != nil {
					k := strings.Trim(kp,SPACE)
					if strings.Contains(k,AS) {
						adapterResultKeyAlia(km,k,AS)
					}else if strings.HasSuffix(k,END_SUB) {
						panic(k + ", AS $alia required, multiSource, suggested fmt: AS `t0.c0`")
					}else if strings.Contains(k,SPACE) {
						if strings.HasPrefix(k,DISTINCT) || strings.HasPrefix(k,Distinct) {
							var alia = "c" + strconv.Itoa(len(km))
							km[alia] = strings.Split(k,SPACE)[1]
							kp = kp + AS + alia
						}else {
							adapterResultKeyAlia(km,k,SPACE)
						}
					}else {
						var alia = "c" + strconv.Itoa(len(km))
						km[alia] = k
						kp = kp + AS + alia
					}
				}
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
	if built.ResultKeys != nil && len(built.ResultKeys) > 0{
		bpCount.WriteString(COUNT_KEY_SCRIPT_LEFT)
		bpCount.WriteString(built.ResultKeys[0])
		bpCount.WriteString(END_SUB)
		bpCount.WriteString(SPACE)
	}else {
		bpCount.WriteString(COUNT_BASE_SCRIPT)
	}
}
