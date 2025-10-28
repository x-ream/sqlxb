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

import "time"

type CondBuilder struct {
	bbs []Bb
}

type BoolFunc func() bool

func subCondBuilder() *CondBuilder {
	c := new(CondBuilder)
	c.bbs = []Bb{}
	return c
}

func (cb *CondBuilder) doIn(p string, k string, vs ...interface{}) *CondBuilder {
	if vs == nil || len(vs) == 0 {
		return cb
	}
	if len(vs) == 1 && (vs[0] == nil || vs[0] == "") {
		return cb
	}

	ss := []string{}
	length := len(vs)
	for i := 0; i < length; i++ {
		v := vs[i]
		if v == nil {
			continue
		}
		switch v.(type) {
		case string:
			s := "'"
			s += v.(string)
			s += "'"
			ss = append(ss, s)
		case uint64, uint, int, int64, int32, int16, int8, byte, float64, float32:
			s := N2s(v)
			if s == "0" {
				continue
			}
			ss = append(ss, s)
		case *uint64, *uint, *int, *int64, *int32, *int16, *int8, *byte, *float64, *float32:
			s, isOK := Np2s(v)
			if !isOK {
				continue
			}
			ss = append(ss, s)
		case interface{}:
			panic("Builder.doIn(ke, (obj), ([]arr) ? ...")
		default:
			panic("Builder.doIn(ke, (*obj)), (*[]arr) ? ...")
		}
	}

	bb := Bb{
		op:    p,
		key:   k,
		value: &ss,
	}
	cb.bbs = append(cb.bbs, bb)

	return cb
}

func (cb *CondBuilder) doLike(p string, k string, v string) *CondBuilder {

	bb := Bb{
		op:    p,
		key:   k,
		value: v,
	}
	cb.bbs = append(cb.bbs, bb)

	return cb
}

func (cb *CondBuilder) doGLE(p string, k string, v interface{}) *CondBuilder {

	switch v.(type) {
	case string:
		if v.(string) == "" {
			return cb
		}
	case float64:
		if v.(float64) == 0.0 {
			return cb
		}
	case float32:
		if v.(float32) == 0.0 {
			return cb
		}
	case uint64:
		if v.(uint64) == 0 {
			return cb
		}
	case uint:
		if v.(uint) == 0 {
			return cb
		}
	case int64:
		if v.(int64) == 0 {
			return cb
		}
	case int:
		if v.(int) == 0 {
			return cb
		}
	case int32:
		if v.(int32) == 0 {
			return cb
		}
	case int16:
		if v.(int16) == 0 {
			return cb
		}
	case int8:
		if v.(int8) == 0 {
			return cb
		}
	case byte:
		if v.(byte) == 0 {
			return cb
		}
	case bool:
		if v.(bool) == false {
			return cb
		}
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return cb
		}
		return cb.addBb(p, k, n)
	case time.Time:
		ts := v.(time.Time).Format("2006-01-02 15:04:05")
		return cb.addBb(p, k, ts)
	case interface{}:
		panic("Builder.doGLE(ke, obj, [] ? ...")
	default:
		if v == nil {
			return cb
		}
	}
	return cb.addBb(p, k, v)
}

func (cb *CondBuilder) addBb(op string, key string, v interface{}) *CondBuilder {
	bb := Bb{
		op:    op,
		key:   key,
		value: v,
	}
	cb.bbs = append(cb.bbs, bb)

	return cb
}

func (cb *CondBuilder) null(op string, k string) *CondBuilder {
	bb := Bb{
		op:  op,
		key: k,
	}
	cb.bbs = append(cb.bbs, bb)
	return cb
}

func (cb *CondBuilder) orAndSub(orAnd string, f func(cb *CondBuilder)) *CondBuilder {
	c := subCondBuilder()
	f(c)
	if c.bbs == nil || len(c.bbs) == 0 {
		return cb
	}

	// ⭐ 检查是否有实际的条件（不仅仅是纯操作符）
	hasRealCondition := false
	for _, b := range c.bbs {
		// 纯操作符 Bb：op=OR/AND, key="", value=nil, subs=nil/empty
		isPureOperator := (b.op == OR || b.op == AND) && b.key == "" && b.value == nil && (b.subs == nil || len(b.subs) == 0)
		if !isPureOperator {
			hasRealCondition = true
			break
		}
	}

	// 如果没有实际条件（只有纯操作符），不添加此 OR/AND 子查询
	if !hasRealCondition {
		return cb
	}

	bb := Bb{
		op:   orAnd,
		key:  orAnd,
		subs: c.bbs, // ⭐ 保留所有 bbs（包括纯操作符，它们用于连接条件）
	}
	cb.bbs = append(cb.bbs, bb)
	return cb
}

func (cb *CondBuilder) orAnd(orAnd string) *CondBuilder {
	length := len(cb.bbs)
	if length == 0 {
		return cb
	}
	pre := cb.bbs[length-1]
	if pre.op == OR {
		return cb
	}
	bb := Bb{
		op: orAnd,
	}
	cb.bbs = append(cb.bbs, bb)
	return cb
}

func (cb *CondBuilder) And(f func(cb *CondBuilder)) *CondBuilder {
	return cb.orAndSub(AND_SUB, f)
}

func (cb *CondBuilder) Or(f func(cb *CondBuilder)) *CondBuilder {
	return cb.orAndSub(OR_SUB, f)
}

func (cb *CondBuilder) OR() *CondBuilder {
	return cb.orAnd(OR)
}

func (cb *CondBuilder) Bool(preCond BoolFunc, f func(cb *CondBuilder)) *CondBuilder {
	if preCond == nil {
		panic("CondBuilder.Bool para of BoolFunc can not nil")
	}
	if !preCond() {
		return cb
	}
	if f == nil {
		panic("CondBuilder.Bool para of func(k string, vs... interface{}) can not nil")
	}
	f(cb)
	return cb
}

func (cb *CondBuilder) Eq(k string, v interface{}) *CondBuilder {
	return cb.doGLE(EQ, k, v)
}
func (cb *CondBuilder) Ne(k string, v interface{}) *CondBuilder {
	return cb.doGLE(NE, k, v)
}
func (cb *CondBuilder) Gt(k string, v interface{}) *CondBuilder {
	return cb.doGLE(GT, k, v)
}
func (cb *CondBuilder) Lt(k string, v interface{}) *CondBuilder {
	return cb.doGLE(LT, k, v)
}
func (cb *CondBuilder) Gte(k string, v interface{}) *CondBuilder {
	return cb.doGLE(GTE, k, v)
}
func (cb *CondBuilder) Lte(k string, v interface{}) *CondBuilder {
	return cb.doGLE(LTE, k, v)
}

// Like sql: LIKE %value%, Like() default has double %
func (cb *CondBuilder) Like(k string, v string) *CondBuilder {
	if v == "" {
		return cb
	}
	return cb.doLike(LIKE, k, "%"+v+"%")
}
func (cb *CondBuilder) NotLike(k string, v string) *CondBuilder {
	if v == "" {
		return cb
	}
	return cb.doLike(NOT_LIKE, k, "%"+v+"%")
}

// LikeLeft sql: LIKE value%, Like() default has double %, then LikeLeft() remove left %
func (cb *CondBuilder) LikeLeft(k string, v string) *CondBuilder {
	if v == "" {
		return cb
	}
	return cb.doLike(LIKE, k, v+"%")
}
func (cb *CondBuilder) In(k string, vs ...interface{}) *CondBuilder {
	return cb.doIn(IN, k, vs...)
}
func (cb *CondBuilder) Nin(k string, vs ...interface{}) *CondBuilder {
	return cb.doIn(NIN, k, vs...)
}
func (cb *CondBuilder) IsNull(key string) *CondBuilder {
	return cb.null(IS_NULL, key)
}
func (cb *CondBuilder) NonNull(key string) *CondBuilder {
	return cb.null(NON_NULL, key)
}

func (cb *CondBuilder) X(k string, vs ...interface{}) *CondBuilder {
	bb := Bb{
		op:    XX,
		key:   k,
		value: vs,
	}
	cb.bbs = append(cb.bbs, bb)
	return cb
}
