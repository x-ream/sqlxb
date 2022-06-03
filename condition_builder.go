/*
 * Copyright 2020 io.xream.sqlxb
 *
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package sqlxb

type ConditionBuilder struct {
	bbs *[]*Bb
}

type BoolFunc func() bool

func SubCondition() *ConditionBuilder {
	c := new(ConditionBuilder)
	c.bbs = &[]*Bb{}
	return c
}

func (builder *ConditionBuilder) X(k string, vs ...interface{}) *ConditionBuilder {
	bb := Bb{
		op:    X,
		key:   k,
		value: vs,
	}
	*builder.bbs = append(*builder.bbs, &bb)
	return builder
}

func (builder *ConditionBuilder) doIn(p string, k string, arr *[]interface{}) *ConditionBuilder {
	if arr == nil || len(*arr) == 0 {
		return builder
	}

	ss := []string{}
	length := len(*arr)
	for i := 0; i < length; i++ {
		v := (*arr)[i]
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
			if isOK == false {
				continue
			}
			ss = append(ss, s)
		case []interface{}:
			panic("Builder.doIn(ke, ([]arr)), ([]arr) ?")
		}
	}

	bb := Bb{
		op:    p,
		key:   k,
		value: &ss,
	}
	*builder.bbs = append(*builder.bbs, &bb)

	return builder
}

func (builder *ConditionBuilder) doLike(p string, k string, v string) *ConditionBuilder {

	bb := Bb{
		op:    p,
		key:   k,
		value: v,
	}
	*builder.bbs = append(*builder.bbs, &bb)

	return builder
}

func (builder *ConditionBuilder) doGLE(p string, k string, v interface{}) *ConditionBuilder {
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
		if IsNil(v) {
			return builder
		}
	case []interface{}:
		panic("Builder.doGLE(ke, []arr), [] ?")
	default:
		if v == nil {
			return builder
		}
	}

	bb := Bb{
		op:    p,
		key:   k,
		value: v,
	}
	*builder.bbs = append(*builder.bbs, &bb)

	return builder
}

func (builder *ConditionBuilder) null(op string, k string) *ConditionBuilder {
	bb := Bb{
		op:  op,
		key: k,
	}
	*builder.bbs = append(*builder.bbs, &bb)
	return builder
}

func (builder *ConditionBuilder) orAndSub(orAnd string, sub *ConditionBuilder) *ConditionBuilder {
	if sub.bbs == nil || len(*sub.bbs) == 0 {
		return builder
	}

	bb := Bb{
		op:   orAnd,
		key:  orAnd,
		subs: sub.bbs,
	}
	*builder.bbs = append(*builder.bbs, &bb)
	return builder
}

func (builder *ConditionBuilder) orAnd(orAnd string) *ConditionBuilder {
	length := len(*builder.bbs)
	if length == 0 {
		return builder
	}
	pre := *(*builder.bbs)[length-1]
	if pre.op == OR {
		return builder
	}
	bb := Bb{
		op: orAnd,
	}
	*builder.bbs = append(*builder.bbs, &bb)
	return builder
}

func (builder *ConditionBuilder) And(subCondition *ConditionBuilder) *ConditionBuilder {
	return builder.orAndSub(AND_SUB, subCondition)
}

func (builder *ConditionBuilder) Or(sub *ConditionBuilder) *ConditionBuilder {
	return builder.orAndSub(OR_SUB, sub)
}

func (builder *ConditionBuilder) OR() *ConditionBuilder {
	return builder.orAnd(OR)
}

func (builder *ConditionBuilder) Bool(preCondition BoolFunc, then func(cb *ConditionBuilder)) *ConditionBuilder {
	if preCondition == nil {
		panic("ConditionBuilder.Bool para of BoolFunc can not nil")
	}
	if preCondition() == false {
		return builder
	}
	if then == nil {
		panic("ConditionBuilder.Bool para of func(k string, vs... interface{}) can not nil")
	}
	then(builder)
	return builder
}

func (builder *ConditionBuilder) Eq(k string, v interface{}) *ConditionBuilder {
	return builder.doGLE(EQ, k, v)
}
func (builder *ConditionBuilder) Ne(k string, v interface{}) *ConditionBuilder {
	return builder.doGLE(NE, k, v)
}
func (builder *ConditionBuilder) Gt(k string, v interface{}) *ConditionBuilder {
	return builder.doGLE(GT, k, v)
}
func (builder *ConditionBuilder) Lt(k string, v interface{}) *ConditionBuilder {
	return builder.doGLE(LT, k, v)
}
func (builder *ConditionBuilder) Gte(k string, v interface{}) *ConditionBuilder {
	return builder.doGLE(GTE, k, v)
}
func (builder *ConditionBuilder) Lte(k string, v interface{}) *ConditionBuilder {
	return builder.doGLE(LTE, k, v)
}
func (builder *ConditionBuilder) Like(k string, v string) *ConditionBuilder {
	if v == "" {
		return builder
	}
	return builder.doLike(LIKE, k, "%"+v+"%")
}
func (builder *ConditionBuilder) NotLike(k string, v string) *ConditionBuilder {
	if v == "" {
		return builder
	}
	return builder.doLike(NOT_LIKE, k, "%"+v+"%")
}
func (builder *ConditionBuilder) LikeRight(k string, v string) *ConditionBuilder {
	if v == "" {
		return builder
	}
	return builder.doLike(LIKE, k, v+"%")
}
func (builder *ConditionBuilder) In(k string, arr *[]interface{}) *ConditionBuilder {
	return builder.doIn(IN, k, arr)
}
func (builder *ConditionBuilder) Nin(k string, arr *[]interface{}) *ConditionBuilder {
	return builder.doIn(NIN, k, arr)
}
func (builder *ConditionBuilder) IsNull(key string) *ConditionBuilder {
	return builder.null(IS_NULL, key)
}
func (builder *ConditionBuilder) NonNull(key string) *ConditionBuilder {
	return builder.null(NON_NULL, key)
}
