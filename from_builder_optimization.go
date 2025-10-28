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
)

func (x *BuilderX) WithoutOptimization() *BuilderX {
	x.isWithoutOptimization = true
	return x
}

func (x *BuilderX) optimizeFromBuilder() {
	if x.isWithoutOptimization {
		return
	}
	if len(x.resultKeys) == 0 || len(x.sxs) < 2 {
		return
	}

	x.removeFromBuilder(x.sxs, func(useds *[]*FromX, ele *FromX, i int) bool {
		if i == 0 {
			return false
		}
		if ele.sub != nil || (ele.join != nil && strings.Contains(ele.join.join, "LEFT")) {
			return false
		}
		for _, u := range *useds {
			//if used, can not remove
			if (ele.sub == nil && ele.alia == u.alia) || ele.tableName == u.tableName {
				return false
			}
		}
		for _, v := range *x.conds() {
			if ele.tableName != "" && strings.Contains(v, ele.tableName+".") { //has return or condition
				return false
			}
			if strings.Contains(v, ele.alia+".") { ////has return or condition
				return false
			}
		}

		//target
		for j := len(x.sxs) - 1; j > i; j-- {
			var sb = x.sxs[j]

			if sb.join != nil && sb.join.on != nil && sb.join.on.bbs != nil {
				for _, bb := range sb.join.on.bbs {
					v := bb.key
					if ele.tableName != "" && strings.Contains(v, ele.tableName+".") { //has return or condition
						return false
					}
					if strings.Contains(v, ele.alia+".") { ////has return or condition
						return false
					}
				}
			}
		}

		return true
	})
}

func (x *BuilderX) conds() *[]string {
	condArr := []string{}
	for _, v := range x.resultKeys {
		condArr = append(condArr, v)
	}

	bbps := x.CondBuilder.bbs

	if bbps != nil {
		for _, v := range bbps {
			condArr = append(condArr, v.key)
		}
	}

	if len(x.sxs) > 0 {
		for _, sb := range x.sxs {
			if sb.join != nil && sb.join.on != nil && sb.join.on.bbs != nil {
				for i, bb := range sb.join.on.bbs {
					if i > 0 {
						condArr = append(condArr, bb.key)
					}
				}
			}
		}
	}
	return &condArr
}

func (x *BuilderX) removeFromBuilder(sbs []*FromX, canRemove canRemove) {
	useds := []*FromX{}
	j := 0
	leng := len(sbs)

	for i := leng - 1; i > -1; i-- {
		ele := (sbs)[i]
		if !canRemove(&useds, ele, i) {
			useds = append(useds, ele)
			j++
		}
	}

	length := len(useds)
	j = 0
	if length < leng {
		for i := length - 1; i > -1; i-- { //reverse
			(x.sxs)[j] = useds[i]
			j++
		}
		x.sxs = (x.sxs)[:j]
	}
}

type canRemove func(useds *[]*FromX, ele *FromX, i int) bool
