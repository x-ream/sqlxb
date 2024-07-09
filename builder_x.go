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

import "fmt"

// To build sql, like: SELECT DISTINCT f.id FROM foo f INNER_JOIN JOIN (SELECT foo_id FROM bar) b ON b.foo_id = f.id
// Sql for MySQL, Clickhouse....
//
// @author Sim
type BuilderX struct {
	CondBuilderX
	pageBuilder *PageBuilder

	sorts                 []Sort
	resultKeys            []string
	orFromSql             string
	inserts               *[]Bb
	updates               *[]Bb
	sxs                   []*FromX
	svs                   []interface{}
	havings               []Bb
	groupBys              []string
	aggs                  []Bb
	last                  string
	isDistinct            bool
	isWithoutOptimization bool

	alia string
}

func Of(tableNameOrPo interface{}) *BuilderX {
	x := X()
	if tableNameOrPo != nil {
		switch tableNameOrPo.(type) {
		case string:
			x.orFromSql = tableNameOrPo.(string)
		case Po:
			x.orFromSql = tableNameOrPo.(Po).TableName()
		default:
			panic("No  `func (* Po) TableName() string` of interface Po: " + fmt.Sprintf("%s", tableNameOrPo))
		}
	}
	return x
}

func X() *BuilderX {
	x := new(BuilderX)
	x.bbs = []Bb{}
	x.sxs = []*FromX{}
	return x
}

func (x *BuilderX) FromX(f func(fb *FromBuilder)) *BuilderX {

	if len(x.sxs) == 0 {
		sb := FromX{
			alia: x.alia,
		}
		if x.orFromSql != "" {
			sb.tableName = x.orFromSql
		}
		x.sxs = append(x.sxs, &sb)
	}

	x.orFromSql = ""

	b := FromBuilder{
		xs: &x.sxs,
		x:  x.sxs[0],
	}
	f(&b)
	return x
}

func (x *BuilderX) From(orFromSql string) *BuilderX {
	x.orFromSql = orFromSql
	return x
}

func (x *BuilderX) As(alia string) *BuilderX {
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

func (x *BuilderX) Having(f func(cb *CondBuilderX)) *BuilderX {
	var cb = new(CondBuilderX)
	f(cb)
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

func (x *BuilderX) Sub(s string, f func(sb *BuilderX)) *BuilderX {
	x.CondBuilderX.Sub(s, f)
	return x
}

func (x *BuilderX) Any(f func(x *BuilderX)) *BuilderX {
	f(x)
	return x
}

func (x *BuilderX) And(f func(cb *CondBuilder)) *BuilderX {
	x.CondBuilder.And(f)
	return x
}

func (x *BuilderX) Or(f func(cb *CondBuilder)) *BuilderX {
	x.CondBuilder.Or(f)
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
func (x *BuilderX) LikeLeft(k string, v string) *BuilderX {
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
	if orderBy == "" {
		return x
	}
	if direction == nil {
		sort := Sort{orderBy: orderBy, direction: ""}
		x.sorts = append(x.sorts, sort)
	} else {
		sort := Sort{orderBy: orderBy, direction: direction()}
		x.sorts = append(x.sorts, sort)
	}

	return x
}

func (x *BuilderX) Paged(f func(pb *PageBuilder)) *BuilderX {
	pageBuilder := new(PageBuilder)
	x.pageBuilder = pageBuilder
	f(pageBuilder)
	return x
}

func (x *BuilderX) Last(last string) *BuilderX {
	x.last = last
	return x
}

func (x *BuilderX) Update(f func(ub *UpdateBuilder)) *BuilderX {
	builder := new(UpdateBuilder)
	x.updates = &builder.bbs
	f(builder)
	return x
}

func (x *BuilderX) Insert(f func(b *InsertBuilder)) *BuilderX {
	builder := new(InsertBuilder)
	x.inserts = &builder.bbs
	f(builder)
	return x
}

func (x *BuilderX) Build() *Built {
	if x == nil {
		panic("sqlxb.Builder is nil")
	}

	if x.inserts != nil && len(*(x.inserts)) > 0 {
		built := Built{
			OrFromSql: x.orFromSql,
			Inserts:   x.inserts,
		}
		return &built
	}

	x.optimizeFromBuilder()

	built := Built{
		ResultKeys: x.resultKeys,
		Updates:    x.updates,
		Conds:      x.bbs,
		Sorts:      x.sorts,
		Aggs:       x.aggs,
		Havings:    x.havings,
		GroupBys:   x.groupBys,
		Last:       x.last,
		OrFromSql:  x.orFromSql,
		Fxs:        x.sxs,
		Svs:        x.svs,
	}

	if x.pageBuilder != nil {
		built.PageCondition = &x.pageBuilder.condition
	}

	return &built
}
