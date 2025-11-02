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
	limitValue  int                   // ⭐ 新增：LIMIT 值（v0.10.1）
	offsetValue int                   // ⭐ 新增：OFFSET 值（v0.10.1）
	meta        *interceptor.Metadata // ⭐ 新增：元数据（v0.9.2）
	custom      Custom                // ⭐ 新增：数据库专属配置（v0.11.0）
}

// Meta 获取元数据
func (x *BuilderX) Meta() *interceptor.Metadata {
	if x.meta == nil {
		x.meta = &interceptor.Metadata{}
	}
	return x.meta
}

// Custom 设置数据库专属配置
//
// 参数:
//   - custom: 数据库专属配置（QdrantCustom/MilvusCustom/WeaviateCustom 等）
//
// 返回:
//   - *BuilderX: 链式调用
//
// 示例:
//
//	// Qdrant 高精度模式
//	built := xb.Of("code_vectors").
//	    Custom(xb.QdrantHighPrecision()).
//	    VectorSearch(...).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Qdrant
//
//	// Milvus 默认模式
//	built := xb.Of("users").
//	    Custom(xb.NewMilvusCustom()).
//	    VectorSearch(...).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ⭐ 自动使用 Milvus
func (x *BuilderX) Custom(custom Custom) *BuilderX {
	x.custom = custom
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

// Limit 设置返回记录数量（适用于简单查询，非分页场景）
// 支持 PostgreSQL 和 MySQL
func (x *BuilderX) Limit(limit int) *BuilderX {
	if limit > 0 {
		x.limitValue = limit
	}
	return x
}

// Offset 设置跳过记录数量（通常与 Limit 配合使用）
// 支持 PostgreSQL 和 MySQL
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

	// ⭐ 执行 BeforeBuild 拦截器（只设置元数据）
	for _, ic := range interceptor.GetAll() {
		if err := ic.BeforeBuild(x.Meta()); err != nil {
			panic(fmt.Sprintf("Interceptor %s BeforeBuild failed: %v", ic.Name(), err))
		}
	}

	if x.inserts != nil && len(*(x.inserts)) > 0 {
		built := Built{
			OrFromSql: x.orFromSql,
			Inserts:   x.inserts,
			Meta:      x.meta,   // ⭐ 传递元数据
			Custom:    x.custom, // ⭐ 传递 Custom
		}

		// ⭐ 执行 AfterBuild 拦截器
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
		OrFromSql:   x.orFromSql,
		Fxs:         x.sxs,
		Svs:         x.svs,
		LimitValue:  x.limitValue,  // ⭐ 传递 Limit
		OffsetValue: x.offsetValue, // ⭐ 传递 Offset
		Meta:        x.meta,        // ⭐ 传递元数据
		Custom:      x.custom,      // ⭐ 传递 Custom
	}

	if x.pageBuilder != nil {
		built.PageCondition = &x.pageBuilder.condition
	}

	// ⭐ 执行 AfterBuild 拦截器
	for _, ic := range interceptor.GetAll() {
		if err := ic.AfterBuild(&built); err != nil {
			panic(fmt.Sprintf("Interceptor %s AfterBuild failed: %v", ic.Name(), err))
		}
	}

	return &built
}
