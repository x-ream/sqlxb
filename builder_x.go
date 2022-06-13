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

// To build sql, like: SELECT DISTINCT f.id FROM foo f INNER JOIN (SELECT foo_id FROM bar) b ON b.foo_id = f.id
// Sql for MySQL, Clickhouse....
//
// @author Sim
//
func Sub() *BuilderX {
	return NewBuilderX(nil, "")
}

type BuilderX struct {
	Builder
	resultKeys            []string
	sbs                   []*SourceBuilder
	svs                   []interface{}
	havings               []Bb
	groupBys              []string
	isWithoutOptimization bool
}

func NewBuilderX(po Po, alia string) *BuilderX {
	x := new(BuilderX)
	x.bbs = []Bb{}
	x.sbs = []*SourceBuilder{}
	if po != nil {
		var sb = SourceBuilder{
			po:   po,
			alia: alia,
		}
		x.sbs = append(x.sbs, &sb)
	} else if alia != "" {
		panic("No po, alia: " + alia)
	}
	return x
}

func (x *BuilderX) SourceBuilder() *SourceBuilder {
	var sb = SourceBuilder{}
	x.sbs = append(x.sbs, &sb)
	return &sb
}

func (x *BuilderX) ResultKey(resultKey string) *BuilderX {
	if resultKey != "" {
		x.resultKeys = append(x.resultKeys, resultKey)
	}
	return x
}

func (x *BuilderX) ResultKeys(resultKeys ...string) *BuilderX {
	for _, resultKey := range resultKeys {
		if resultKey != "" {
			x.resultKeys = append(x.resultKeys, resultKey)
		}
	}
	return x
}

func (x *BuilderX) Source(po Po) *BuilderX {
	if po != nil {
		sb := SourceBuilder{
			po: po,
		}
		x.sbs = append(x.sbs, &sb)
	}
	return x
}

func (x *BuilderX) Having(op Op, k string, v interface{}) *BuilderX {
	if op == nil || k == "" {
		return x
	}
	bb := Bb{
		op:    op(),
		key:   k,
		value: v,
	}
	x.havings = append(x.havings, bb)
	return x
}

func (x *BuilderX) GroupBy(groupBy string) *BuilderX {
	if groupBy == "" {
		return x
	}
	x.groupBys = append(x.groupBys, groupBy)
	return x
}

func (builder *BuilderX) Build() *Built {
	if builder == nil {
		panic("sqlxb.Builder is nil")
	}
	builder.optimizeSourceBuilder()
	built := Built{
		ResultKeys: builder.resultKeys,
		ConditionX: builder.bbs,
		Sorts:      builder.sorts,
		Havings:    builder.havings,
		GroupBys:   builder.groupBys,
		Sbs:        builder.sbs,
		Svs:        builder.svs,

		Po: builder.po,
	}

	if builder.pageBuilder != nil {
		built.PageCondition = &builder.pageBuilder.condition
	}

	return &built
}
