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
	"strconv"
	"strings"

	. "github.com/fndome/xb/internal"
)

func adapterResultKeyAlia(km map[string]string, k string, reg string) {
	arr := strings.Split(k, reg)
	alia := arr[1]
	if strings.Contains(alia, "`") {
		alia = strings.Replace(alia, "`", "", 2)
	}
	km[alia] = alia
}

func buildResultKey(key string, km map[string]string) string {
	if km == nil {
		return key
	}

	k := strings.Trim(key, SPACE)
	if strings.Contains(k, AS) {
		adapterResultKeyAlia(km, k, AS)
	} else if strings.HasSuffix(k, END_SUB) {
		panic(k + ", AS $alia required, multiFrom, suggested fmt: AS `t0.c0`")
	} else if strings.Contains(k, SPACE) {
		if strings.HasPrefix(k, DISTINCT_SCRIPT) || strings.HasPrefix(k, Distinct) {
			var alia = "c" + strconv.Itoa(len(km))
			km[alia] = strings.Split(k, SPACE)[1]
			key = key + AS + alia
		} else {
			adapterResultKeyAlia(km, k, SPACE)
		}
	} else if strings.Contains(k, ".") {
		var alia = "c" + strconv.Itoa(len(km))
		km[alia] = k
		key = key + AS + alia
	}
	return key
}

func (built *Built) toResultKeySql(bp *strings.Builder, km map[string]string) {
	if built.Updates != nil {
		bp.WriteString(UPDATE)
		return
	}
	bp.WriteString(SELECT)
	if built.ResultKeys == nil {
		bp.WriteString(STAR)
	} else {
		length := len(built.ResultKeys)
		if length == 0 {
			bp.WriteString(STAR)
		} else {
			for i := 0; i < length; i++ {
				key := (built.ResultKeys)[i]
				key = buildResultKey(key, km)
				bp.WriteString(key)
				if i < length-1 {
					bp.WriteString(COMMA)
				} else {
					bp.WriteString(SPACE)
				}
			}
		}
	}
}

func (built *Built) toResultKeySqlOfCount(bpCount *strings.Builder) {
	if built.ResultKeys != nil && len(built.ResultKeys) > 0 {
		bpCount.WriteString(COUNT_KEY_SCRIPT_LEFT)
		bpCount.WriteString(built.ResultKeys[0])
		bpCount.WriteString(END_SUB)
		bpCount.WriteString(SPACE)
	} else {
		bpCount.WriteString(COUNT_BASE_SCRIPT)
	}
}
