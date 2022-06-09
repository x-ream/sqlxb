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
	"strings"
)

func (builder *BuilderX) WithoutOptimization() *BuilderX {
	builder.isWithoutOptimization = true
	return builder
}

func (builder *BuilderX) optimizeSourceBuilder() {
	if builder.isWithoutOptimization {
		return
	}
	if len(builder.resultKeys) == 0 || len(*builder.sbs) < 2 {
		return
	}

	builder.removeSourceBuilder(builder.sbs, func(useds *[]*SourceBuilder, ele *SourceBuilder) bool {

		for _, u := range *useds {
			//if used, can not remove
			if (ele.sub == nil && ele.alia == u.alia) || ele.po == u.po {
				return false
			}
		}

		for _, v := range *builder.conds() {
			if ele.sub == nil && strings.HasPrefix(v, ele.po.TableName()+".") { //has return or condition
				return false
			}
			if strings.HasPrefix(v, ele.alia+".") {////has return or condition
				return false
			}
		}
		return true
	})
}

func (builder *BuilderX) conds() *[]string {
	condArr := []string{}
	for _, v := range builder.resultKeys {
		condArr = append(condArr, v)
	}

	bbps := builder.ConditionBuilder.bbs

	if bbps != nil {
		for _, v := range *bbps {
			condArr = append(condArr, v.key)
		}
	}

	if len(*builder.sbs) > 0 {
		for _, sb := range *builder.sbs {
			if sb.bbs != nil {
				for _, bb := range *sb.bbs {
					condArr = append(condArr, bb.key)
				}
			}
		}
	}
	return &condArr
}

func (builder *BuilderX) removeSourceBuilder(sbs *[]*SourceBuilder, canRemove canRemove) {
	useds := []*SourceBuilder{}
	j := 0
	leng := len(*sbs)
	for i := leng - 1; i > -1; i-- {
		ele := (*sbs)[i]
		if !canRemove(&useds, ele) {
			useds = append(useds, ele)
			j++
		}
	}

	length := len(useds)
	j = 0
	if length < leng {
		for i := length - 1; i > -1; i-- { //reverse
			(*builder.sbs)[j] = useds[i]
			j++
		}
		*builder.sbs = (*builder.sbs)[:j]
	}
}

type canRemove func(useds *[]*SourceBuilder, ele *SourceBuilder) bool
