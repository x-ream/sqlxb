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
	"github.com/x-ream/sqlxb/internal"
	"strings"
)


func (builder *BuilderX) WithoutOptimization() *BuilderX {
	builder.isWithoutOptimization = true
	return builder
}

func (builder *BuilderX) OptimizeSourceBuilder() {
	if builder.isWithoutOptimization {
		return
	}

	if len(builder.resultKeys) == 0 {
		return
	}

	conds := builder.conds()

	length := len(*builder.sbs)
	newArr := []interface{}{}

	for i := 1; i<length; i++ {
		newArr = append(newArr,(*builder.sbs)[i])
	}
	sbs := internal.DelEle(&newArr, func(e interface{}) bool {
		for _,v := range *conds {
			sb := e.(*SourceBuilder)
			if sb.sub == nil && strings.HasPrefix(v, sb.po.TableName()+".") {
				return false
			}
			if strings.HasPrefix(v, sb.alia+".") {
				return false
			}
		}
		return true
	})
	if len(*sbs) < length -1 {
		*builder.sbs = (*builder.sbs)[:1]
		for _, v := range *sbs {
			*builder.sbs = append(*builder.sbs,v.(*SourceBuilder))
		}
	}
}

func (builder *BuilderX) conds() *[]string {
	condArr := []string{}
	for _,v := range builder.resultKeys {
		condArr = append(condArr,v)
	}

	bbps := builder.ConditionBuilder.bbs

	if bbps != nil {
		for _, v := range *bbps {
			condArr = append(condArr, v.key)
		}
	}

	sbs := builder.sbs

	if len(*sbs) > 0 {
		for _, v := range *sbs {
			sbbps := v.bbs
			if sbbps != nil {
				for _, v := range *sbbps {
					condArr = append(condArr, v.key)
				}
			}
		}
	}
	return &condArr
}