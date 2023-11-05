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

// To build sql, like: SELECT DISTINCT f.id FROM foo f INNER_JOIN JOIN (SELECT foo_id FROM bar) b ON b.foo_id = f.id
// Sql for MySQL, Clickhouse....
//
// @author Sim
type BuilderX struct {
	Builder
	resultKeys            []string
	orSourceSql           string
	sxs                   []*SourceX
	svs                   []interface{}
	havings               []Bb
	groupBys              []string
	aggs                  []Bb
	isWithoutOptimization bool
}

func NewBuilderX() *BuilderX {
	x := new(BuilderX)
	x.Bbs = []Bb{}
	x.sxs = []*SourceX{}

	sb := SourceX{}
	x.sxs = append(x.sxs, &sb)

	return x
}

func Sub(po Po) *BuilderX {
	x := NewBuilderX()
	x.sxs[0].po = po
	return x
}

func (x *BuilderX) SourceX(source func(sb *SourceBuilder)) *BuilderX {
	var b = SourceBuilder{}
	b.xs = &x.sxs
	b.x = x.sxs[0]
	source(&b)
	return x
}

func (x *BuilderX) SourceScript(sqlScript string) *BuilderX {
	x.orSourceSql = sqlScript
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
	if len(x.sxs) == 0 {
		x.sxs = append(x.sxs, new(SourceX))
	}
	x.sxs[0].po = po
	return x
}

func (x *BuilderX) Having(cond func(cb *CondBuilder)) *BuilderX {
	var cb = new(CondBuilder)
	cond(cb)
	x.havings = cb.Bbs
	return x
}

func (x *BuilderX) GroupBy(groupBy string) *BuilderX {
	if groupBy == "" {
		return x
	}
	x.groupBys = append(x.groupBys, groupBy)
	return x
}

func (x *BuilderX) Agg(fn string, vs ...interface{}) *BuilderX {
	if fn == "" {
		return x
	}
	bb := Bb{
		op:    AGG,
		key:   fn,
		value: vs,
	}
	x.aggs = append(x.aggs, bb)
	return x
}

func (x *BuilderX) Build() *Built {
	if x == nil {
		panic("sqlxb.Builder is nil")
	}
	x.optimizeSourceBuilder()
	built := Built{
		ResultKeys:  x.resultKeys,
		ConditionX:  x.Bbs,
		Sorts:       x.sorts,
		Aggs:        x.aggs,
		Havings:     x.havings,
		GroupBys:    x.groupBys,
		OrSourceSql: x.orSourceSql,
		Sbs:         x.sxs,
		Svs:         x.svs,

		Po: x.po,
	}

	if x.pageBuilder != nil {
		built.PageCondition = &x.pageBuilder.condition
	}

	return &built
}

func (x *BuilderX) Sub(s string, sub func(sub *BuilderX)) *BuilderX {

	b := new(BuilderX)
	sub(b)
	bb := Bb{
		op:    SUB,
		key:   s,
		value: b,
	}
	x.Bbs = append(x.Bbs, bb)
	return x
}
