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

import "time"

type CondBuilder struct {
	bbs []Bb
}

type BoolFunc func() bool

func SubCond() *CondBuilder {
	c := new(CondBuilder)
	c.bbs = []Bb{}
	return c
}

func (builder *CondBuilder) X(k string, vs ...interface{}) *CondBuilder {
	bb := Bb{
		op:    X,
		key:   k,
		value: vs,
	}
	builder.bbs = append(builder.bbs, bb)
	return builder
}

func (builder *CondBuilder) doIn(p string, k string, vs ...interface{}) *CondBuilder {
	if vs == nil || len(vs) == 0 {
		return builder
	}
	if len(vs) == 1 && (vs[0] == nil || vs[0] == "") {
		return builder
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
		case []interface{}:
			panic("Builder.doIn(ke, ([]arr)), ([]arr) ?")
		default:
			panic("Builder.doIn(ke, (*[]arr)), (*[]arr) ?")
		}
	}

	bb := Bb{
		op:    p,
		key:   k,
		value: &ss,
	}
	builder.bbs = append(builder.bbs, bb)

	return builder
}

func (builder *CondBuilder) doLike(p string, k string, v string) *CondBuilder {

	bb := Bb{
		op:    p,
		key:   k,
		value: v,
	}
	builder.bbs = append(builder.bbs, bb)

	return builder
}

func (builder *CondBuilder) doGLE(p string, k string, v interface{}) *CondBuilder {

	switch v.(type) {
	case string:
		if v.(string) == "" {
			return builder
		}
	case uint64, uint, int64, int, int32, int16, int8, bool, byte, float64, float32:
		if v == 0 {
			return builder
		}
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return builder
		}
		return builder.addBb(p, k, n)
	case time.Time:
		ts := v.(time.Time).Format("2006-01-02 15:04:05")
		return builder.addBb(p, k, ts)
	case []interface{}:
		panic("Builder.doGLE(ke, []arr), [] ?")
	default:
		if v == nil {
			return builder
		}
	}
	return builder.addBb(p, k, v)
}

func (builder *CondBuilder) addBb(op string, key string, v interface{}) *CondBuilder {
	bb := Bb{
		op:    op,
		key:   key,
		value: v,
	}
	builder.bbs = append(builder.bbs, bb)

	return builder
}

func (builder *CondBuilder) null(op string, k string) *CondBuilder {
	bb := Bb{
		op:  op,
		key: k,
	}
	builder.bbs = append(builder.bbs, bb)
	return builder
}

func (builder *CondBuilder) orAndSub(orAnd string, sub *CondBuilder) *CondBuilder {
	if sub.bbs == nil || len(sub.bbs) == 0 {
		return builder
	}

	bb := Bb{
		op:   orAnd,
		key:  orAnd,
		subs: sub.bbs,
	}
	builder.bbs = append(builder.bbs, bb)
	return builder
}

func (builder *CondBuilder) orAnd(orAnd string) *CondBuilder {
	length := len(builder.bbs)
	if length == 0 {
		return builder
	}
	pre := builder.bbs[length-1]
	if pre.op == OR {
		return builder
	}
	bb := Bb{
		op: orAnd,
	}
	builder.bbs = append(builder.bbs, bb)
	return builder
}

func (builder *CondBuilder) And(subCondition *CondBuilder) *CondBuilder {
	return builder.orAndSub(AND_SUB, subCondition)
}

func (builder *CondBuilder) Or(sub *CondBuilder) *CondBuilder {
	return builder.orAndSub(OR_SUB, sub)
}

func (builder *CondBuilder) OR() *CondBuilder {
	return builder.orAnd(OR)
}

func (builder *CondBuilder) Bool(preCondition BoolFunc, then func(cb *CondBuilder)) *CondBuilder {
	if preCondition == nil {
		panic("CondBuilder.Bool para of BoolFunc can not nil")
	}
	if !preCondition() {
		return builder
	}
	if then == nil {
		panic("CondBuilder.Bool para of func(k string, vs... interface{}) can not nil")
	}
	then(builder)
	return builder
}

func (builder *CondBuilder) Eq(k string, v interface{}) *CondBuilder {
	return builder.doGLE(EQ, k, v)
}
func (builder *CondBuilder) Ne(k string, v interface{}) *CondBuilder {
	return builder.doGLE(NE, k, v)
}
func (builder *CondBuilder) Gt(k string, v interface{}) *CondBuilder {
	return builder.doGLE(GT, k, v)
}
func (builder *CondBuilder) Lt(k string, v interface{}) *CondBuilder {
	return builder.doGLE(LT, k, v)
}
func (builder *CondBuilder) Gte(k string, v interface{}) *CondBuilder {
	return builder.doGLE(GTE, k, v)
}
func (builder *CondBuilder) Lte(k string, v interface{}) *CondBuilder {
	return builder.doGLE(LTE, k, v)
}
func (builder *CondBuilder) Like(k string, v string) *CondBuilder {
	if v == "" {
		return builder
	}
	return builder.doLike(LIKE, k, "%"+v+"%")
}
func (builder *CondBuilder) NotLike(k string, v string) *CondBuilder {
	if v == "" {
		return builder
	}
	return builder.doLike(NOT_LIKE, k, "%"+v+"%")
}
func (builder *CondBuilder) LikeRight(k string, v string) *CondBuilder {
	if v == "" {
		return builder
	}
	return builder.doLike(LIKE, k, v+"%")
}
func (builder *CondBuilder) In(k string, vs ...interface{}) *CondBuilder {
	return builder.doIn(IN, k, vs...)
}
func (builder *CondBuilder) Nin(k string, vs ...interface{}) *CondBuilder {
	return builder.doIn(NIN, k, vs...)
}
func (builder *CondBuilder) IsNull(key string) *CondBuilder {
	return builder.null(IS_NULL, key)
}
func (builder *CondBuilder) NonNull(key string) *CondBuilder {
	return builder.null(NON_NULL, key)
}
