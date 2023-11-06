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
	CondBuilder
	pageBuilder *PageBuilder

	sorts                 []Sort
	resultKeys            []string
	orFromSql             string
	sxs                   []*FromX
	svs                   []interface{}
	havings               []Bb
	groupBys              []string
	aggs                  []Bb
	isWithoutOptimization bool

	po   Po
	alia string
}

func Of(po Po) *BuilderX {
	x := new(BuilderX)
	x.bbs = []Bb{}
	x.sxs = []*FromX{}
	x.po = po
	return x
}

func (x *BuilderX) OfX(fromx func(sb *FromBuilder)) *BuilderX {

	if len(x.sxs) == 0 {
		sb := FromX{
			po:   x.po,
			alia: x.alia,
		}
		x.sxs = append(x.sxs, &sb)
	}

	x.po = nil
	x.alia = ""

	var b = FromBuilder{}
	b.xs = &x.sxs
	b.x = x.sxs[0]
	fromx(&b)
	return x
}

func (x *BuilderX) FromScript(sqlScript string) *BuilderX {
	x.orFromSql = sqlScript
	return x
}

func (x *BuilderX) Alia(alia string) *BuilderX {
	x.alia = alia
	return x
}

func (x *BuilderX) Select(resultKeys ...string) *BuilderX {
	for _, resultKey := range resultKeys {
		if resultKey != "" {
			x.resultKeys = append(x.resultKeys, resultKey)
		}
	}
	return x
}

func (x *BuilderX) Of(po Po) *BuilderX {
	x.po = po
	return x
}

func (x *BuilderX) Having(cond func(cb *CondBuilder)) *BuilderX {
	var cb = new(CondBuilder)
	cond(cb)
	x.havings = cb.bbs
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

func (x *BuilderX) Sub(s string, sub func(sub *BuilderX)) *BuilderX {

	b := new(BuilderX)
	sub(b)
	bb := Bb{
		op:    SUB,
		key:   s,
		value: b,
	}
	x.bbs = append(x.bbs, bb)
	return x
}

func (x *BuilderX) And(sub func(sub *CondBuilder)) *BuilderX {
	x.CondBuilder.And(sub)
	return x
}

func (x *BuilderX) Or(sub func(sub *CondBuilder)) *BuilderX {
	x.CondBuilder.Or(sub)
	return x
}

func (x *BuilderX) OR() *BuilderX {
	x.CondBuilder.OR()
	return x
}

func (x *BuilderX) Bool(preCond BoolFunc, then func(cb *CondBuilder)) *BuilderX {
	x.CondBuilder.Bool(preCond, then)
	return x
}

func (x *BuilderX) Eq(k string, v interface{}) *BuilderX {
	x.doGLE(EQ, k, v)
	return x
}
func (x *BuilderX) Ne(k string, v interface{}) *BuilderX {
	x.doGLE(NE, k, v)
	return x
}
func (x *BuilderX) Gt(k string, v interface{}) *BuilderX {
	x.doGLE(GT, k, v)
	return x
}
func (x *BuilderX) Lt(k string, v interface{}) *BuilderX {
	x.doGLE(LT, k, v)
	return x
}
func (x *BuilderX) Gte(k string, v interface{}) *BuilderX {
	x.doGLE(GTE, k, v)
	return x
}
func (x *BuilderX) Lte(k string, v interface{}) *BuilderX {
	x.doGLE(LTE, k, v)
	return x
}
func (x *BuilderX) Like(k string, v string) *BuilderX {
	if v == "" {
		return x
	}
	x.doLike(LIKE, k, "%"+v+"%")
	return x
}
func (x *BuilderX) NotLike(k string, v string) *BuilderX {
	if v == "" {
		return x
	}
	x.doLike(NOT_LIKE, k, "%"+v+"%")
	return x
}
func (x *BuilderX) LikeRight(k string, v string) *BuilderX {
	if v == "" {
		return x
	}
	x.doLike(LIKE, k, v+"%")
	return x
}
func (x *BuilderX) In(k string, vs ...interface{}) *BuilderX {
	x.doIn(IN, k, vs...)
	return x
}
func (x *BuilderX) Nin(k string, vs ...interface{}) *BuilderX {
	x.doIn(NIN, k, vs...)
	return x
}
func (x *BuilderX) IsNull(key string) *BuilderX {
	x.null(IS_NULL, key)
	return x
}
func (x *BuilderX) NonNull(key string) *BuilderX {
	x.null(NON_NULL, key)
	return x
}

func (x *BuilderX) X(k string, vs ...interface{}) *BuilderX {
	x.CondBuilder.X(k, vs...)
	return x
}

func (x *BuilderX) Sort(orderBy string, direction Direction) *BuilderX {
	if orderBy == "" || direction == nil {
		return x
	}
	sort := Sort{orderBy: orderBy, direction: direction()}
	x.sorts = append(x.sorts, sort)
	return x
}

func (x *BuilderX) Paged(page func(pb *PageBuilder)) *BuilderX {
	pageBuilder := new(PageBuilder)
	x.pageBuilder = pageBuilder
	page(pageBuilder)
	return x
}

func (x *BuilderX) Build() *Built {
	if x == nil {
		panic("sqlxb.Builder is nil")
	}
	x.optimizeFromBuilder()
	built := Built{
		ResultKeys: x.resultKeys,
		ConditionX: x.bbs,
		Sorts:      x.sorts,
		Aggs:       x.aggs,
		Havings:    x.havings,
		GroupBys:   x.groupBys,
		orFromSql:  x.orFromSql,
		Sbs:        x.sxs,
		Svs:        x.svs,

		Po: x.po,
	}

	if x.pageBuilder != nil {
		built.PageCondition = &x.pageBuilder.condition
	}

	return &built
}
