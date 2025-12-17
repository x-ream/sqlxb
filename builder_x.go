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
	"fmt"
	"strings"

	"github.com/fndome/xb/interceptor"
)

// ToId build sql, like: SELECT DISTINCT f.id FROM foo f INNER_JOIN JOIN (SELECT foo_id FROM bar) b ON b.foo_id = f.id
// Sql for MySQL, Clickhouse....
//
// @author Sim
type BuilderX struct {
	CondBuilderX
	pageBuilder *PageBuilder

	inserts               *[]Bb
	updates               *[]Bb
	sorts                 []Sort
	resultKeys            []string
	orFromSql             string
	sxs                   []*FromX
	svs                   []interface{}
	havings               []Bb
	groupBys              []string
	aggs                  []Bb
	last                  string
	isDistinct            bool
	isWithoutOptimization bool

	alia        string
	limitValue  int                   // ⭐ LIMIT value (v0.10.1)
	offsetValue int                   // ⭐ OFFSET value (v0.10.1)
	meta        *interceptor.Metadata // ⭐ Metadata (v0.9.2)
	customImpl  Custom                // ⭐ Database-specific config (v0.11.0) (private field)
	withs       []withClause
	unions      []unionClause
}

type withClause struct {
	name      string
	recursive bool
	builder   *BuilderX
}

type unionClause struct {
	operator string
	builder  *BuilderX
}

// Meta configures metadata (chainable)
// Mainly used to pass through TraceID, TenantID and other context information
func (x *BuilderX) Meta(fn func(meta *interceptor.Metadata)) *BuilderX {
	if fn != nil {
		fn(x.ensureMeta())
	}
	return x
}

func (x *BuilderX) ensureMeta() *interceptor.Metadata {
	if x.meta == nil {
		x.meta = &interceptor.Metadata{}
	}
	return x.meta
}

// Custom sets database-specific configuration
//
// Parameters:
//   - custom: Database-specific config (QdrantCustom/MilvusCustom/WeaviateCustom, etc.)
//
// Returns:
//   - *BuilderX: Chainable
//
// Example:
//
//	// Using QdrantBuilder (recommended)
//	built := xb.Of(&CodeVector{}).
//	    Custom(
//	        xb.NewQdrantBuilder().
//	            HnswEf(512).
//	            ScoreThreshold(0.8).
//	            Build()
//	    ).
//	    Insert(...).
//	    Build()
//
//	json, _ := built.JsonOfInsert()  // ⭐ Automatically uses Qdrant
//
//	// Direct Custom usage (example: future implementation using Builder pattern)
//	// built := xb.Of(&User{}).
//	//     Custom(xb.NewMilvusBuilder().Build()).
//	//     Insert(...).
//	//     Build()
//	//
//	// json, _ := built.JsonOfInsert()  // ⭐ Automatically uses Milvus
func (x *BuilderX) Custom(custom Custom) *BuilderX {
	x.customImpl = custom
	return x
}

// With defines a common table expression (CTE)
func (x *BuilderX) With(name string, fn func(sb *BuilderX)) *BuilderX {
	return x.addWith(name, false, fn)
}

// WithRecursive defines a recursive common table expression (CTE)
func (x *BuilderX) WithRecursive(name string, fn func(sb *BuilderX)) *BuilderX {
	return x.addWith(name, true, fn)
}

func (x *BuilderX) addWith(name string, recursive bool, fn func(sb *BuilderX)) *BuilderX {
	if name == "" || fn == nil {
		return x
	}
	sb := X()
	sb.customImpl = x.customImpl
	fn(sb)
	x.withs = append(x.withs, withClause{
		name:      name,
		recursive: recursive,
		builder:   sb,
	})
	return x
}

// UNION combines queries
func (x *BuilderX) UNION(kind UNION, fn func(sb *BuilderX)) *BuilderX {
	if fn == nil {
		return x
	}
	sb := X()
	sb.customImpl = x.customImpl
	fn(sb)
	operator := unionDistinct
	if kind != nil {
		operator = kind()
		if operator == "" {
			operator = unionDistinct
		}
	}
	x.unions = append(x.unions, unionClause{
		operator: operator,
		builder:  sb,
	})
	return x
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
		Op:    AGG,
		Key:   fn,
		Value: vs,
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

// InRequired required IN condition (panics on empty values)
// See CondBuilder.InRequired() documentation for details
func (x *BuilderX) InRequired(k string, vs ...interface{}) *BuilderX {
	x.CondBuilder.InRequired(k, vs...)
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

// Limit sets the number of records to return (for simple queries, not pagination)
// Supports PostgreSQL and MySQL
func (x *BuilderX) Limit(limit int) *BuilderX {
	if limit > 0 {
		x.limitValue = limit
	}
	return x
}

// Offset sets the number of records to skip (usually used with Limit)
// Supports PostgreSQL and MySQL
func (x *BuilderX) Offset(offset int) *BuilderX {
	if offset > 0 {
		x.offsetValue = offset
	}
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
		panic("xb.Builder is nil")
	}

	// ⭐ Execute BeforeBuild interceptors (only set metadata)
	for _, ic := range interceptor.GetAll() {
		if err := ic.BeforeBuild(x.ensureMeta()); err != nil {
			panic(fmt.Sprintf("Interceptor %s BeforeBuild failed: %v", ic.Name(), err))
		}
	}

	baseFrom := x.normalizeFrom()
	withs := x.buildWithClauses()
	unions := x.buildUnionClauses()

	if x.inserts != nil && len(*(x.inserts)) > 0 {
		built := Built{
			OrFromSql: baseFrom,
			Inserts:   x.inserts,
			Meta:      x.meta,       // ⭐ Pass metadata
			Custom:    x.customImpl, // ⭐ Pass Custom
			Alia:      x.alia,
			Withs:     withs,
			Unions:    unions,
		}

		// ⭐ Execute AfterBuild interceptors
		for _, ic := range interceptor.GetAll() {
			if err := ic.AfterBuild(&built); err != nil {
				panic(fmt.Sprintf("Interceptor %s AfterBuild failed: %v", ic.Name(), err))
			}
		}

		return &built
	}

	x.optimizeFromBuilder()

	built := Built{
		ResultKeys:  x.resultKeys,
		Updates:     x.updates,
		Conds:       x.bbs,
		Sorts:       x.sorts,
		Aggs:        x.aggs,
		Havings:     x.havings,
		GroupBys:    x.groupBys,
		Last:        x.last,
		OrFromSql:   baseFrom,
		Fxs:         x.sxs,
		Svs:         x.svs,
		LimitValue:  x.limitValue,
		OffsetValue: x.offsetValue,
		Meta:        x.meta,
		Custom:      x.customImpl,
		Alia:        x.alia,
		Withs:       withs,
		Unions:      unions,
	}

	if x.pageBuilder != nil {
		built.PageCondition = &x.pageBuilder.condition
	}

	for _, ic := range interceptor.GetAll() {
		if err := ic.AfterBuild(&built); err != nil {
			panic(fmt.Sprintf("Interceptor %s AfterBuild failed: %v", ic.Name(), err))
		}
	}

	return &built
}

func (x *BuilderX) normalizeFrom() string {
	if x.orFromSql == "" {
		return ""
	}
	if x.alia == "" {
		return x.orFromSql
	}
	if len(strings.Fields(x.orFromSql)) == 1 {
		return x.orFromSql + " " + x.alia
	}
	return x.orFromSql
}

func (x *BuilderX) buildWithClauses() []WithClause {
	if len(x.withs) == 0 {
		return nil
	}

	result := make([]WithClause, 0, len(x.withs))
	for _, clause := range x.withs {
		if clause.builder == nil {
			continue
		}
		subBuilt := clause.builder.Build()
		sql, args, _ := subBuilt.SqlOfSelect()
		result = append(result, WithClause{
			Name:      clause.name,
			SQL:       sql,
			Args:      append([]interface{}(nil), args...),
			Recursive: clause.recursive,
		})
	}
	return result
}

func (x *BuilderX) buildUnionClauses() []UnionClause {
	if len(x.unions) == 0 {
		return nil
	}
	result := make([]UnionClause, 0, len(x.unions))
	for _, clause := range x.unions {
		if clause.builder == nil {
			continue
		}
		subBuilt := clause.builder.Build()
		sql, args, _ := subBuilt.SqlOfSelect()
		result = append(result, UnionClause{
			Operator: clause.operator,
			SQL:      sql,
			Args:     append([]interface{}(nil), args...),
		})
	}
	return result
}
