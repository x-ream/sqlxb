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
		Op:    p,
		Key:   k,
		Value: &ss,
	}
	cb.bbs = append(cb.bbs, bb)

	return cb
}

func (cb *CondBuilder) doLike(p string, k string, v string) *CondBuilder {

	bb := Bb{
		Op:    p,
		Key:   k,
		Value: v,
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
		Op:    op,
		Key:   key,
		Value: v,
	}
	cb.bbs = append(cb.bbs, bb)

	return cb
}

func (cb *CondBuilder) null(op string, k string) *CondBuilder {
	bb := Bb{
		Op:  op,
		Key: k,
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

	// ⭐ Check if there are real conditions (not just pure operators)
	hasRealCondition := false
	for _, b := range c.bbs {
		// Pure operator Bb: op=OR/AND, key="", value=nil, subs=nil/empty
		isPureOperator := (b.Op == OR || b.Op == AND) && b.Key == "" && b.Value == nil && (b.Subs == nil || len(b.Subs) == 0)
		if !isPureOperator {
			hasRealCondition = true
			break
		}
	}

	// If there are no real conditions (only pure operators), don't add this OR/AND subquery
	if !hasRealCondition {
		return cb
	}

	bb := Bb{
		Op:   orAnd,
		Key:  orAnd,
		Subs: c.bbs, // ⭐ Keep all bbs (including pure operators, they are used to connect conditions)
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
	if pre.Op == OR {
		return cb
	}
	bb := Bb{
		Op: orAnd,
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

// InRequired required IN condition (panics on empty values)
// Used for scenarios where filter conditions must be provided to prevent accidentally querying all data
//
// Panic scenarios:
//   - Empty slice: InRequired("id") or InRequired("id", []int{}...)
//   - nil value: InRequired("id", nil)
//   - Zero value: InRequired("id", 0)
//
// Example:
//
//	// ✅ Normal usage
//	ids := []int{1, 2, 3}
//	xb.Of(&User{}).InRequired("id", toInterfaces(ids)...).Build()
//
//	// ❌ Panic: empty slice
//	ids := []int{}
//	xb.Of(&User{}).InRequired("id", toInterfaces(ids)...).Build()
//	// panic: InRequired("id") received empty values, this would match all records
func (cb *CondBuilder) InRequired(k string, vs ...interface{}) *CondBuilder {
	// Check if empty
	if vs == nil || len(vs) == 0 {
		panic("InRequired(\"" + k + "\") received empty values, this would match all records. Use In() if optional filtering is intended.")
	}

	// Check if there's only one nil or 0
	if len(vs) == 1 {
		v := vs[0]
		if v == nil {
			panic("InRequired(\"" + k + "\") received [nil], this would match all records. Use In() if optional filtering is intended.")
		}
		// Check various zero values
		switch v.(type) {
		case int:
			if v.(int) == 0 {
				panic("InRequired(\"" + k + "\") received [0], this would match all records. Use In() if optional filtering is intended.")
			}
		case int64:
			if v.(int64) == 0 {
				panic("InRequired(\"" + k + "\") received [0], this would match all records. Use In() if optional filtering is intended.")
			}
		case int32:
			if v.(int32) == 0 {
				panic("InRequired(\"" + k + "\") received [0], this would match all records. Use In() if optional filtering is intended.")
			}
		case uint:
			if v.(uint) == 0 {
				panic("InRequired(\"" + k + "\") received [0], this would match all records. Use In() if optional filtering is intended.")
			}
		case uint64:
			if v.(uint64) == 0 {
				panic("InRequired(\"" + k + "\") received [0], this would match all records. Use In() if optional filtering is intended.")
			}
		case string:
			if v.(string) == "" {
				panic("InRequired(\"" + k + "\") received [\"\"], this would match all records. Use In() if optional filtering is intended.")
			}
		}
	}

	// Call normal doIn (will filter out nil and zero values)
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

// X custom SQL fragment (universal supplement)
//
// Used for special scenarios: query zero values, false, or complex SQL expressions
//
// Two usage patterns:
//
//  1. No parameters (recommended): write SQL fragment directly
//     .X("age = 0")
//     .X("is_active = false")
//     .X("YEAR(created_at) = 2024")
//
//  2. With parameters: use placeholders (still filters zero values)
//     .X("name = ?", name)  // name="" will be filtered
//     .X("age > ?", age)    // age=0 will be filtered ⚠️
//
// ⚠️ Important: If you want to query zero values or false, use the no-parameter approach!
//
// Example:
//
//	// ✅ Query age = 0
//	xb.Of("users").X("age = 0").Build()
//
//	// ✅ Query is_active = false
//	xb.Of("users").X("is_active = false").Build()
//
//	// ✅ Complex expression
//	xb.Of("orders").X("total_amt > discount_amt").Build()
//
//	// ❌ Wrong: age=0 will be filtered
//	xb.Of("users").X("age = ?", 0).Build()
//
//	// ✅ Correct: write SQL directly
//	xb.Of("users").X("age = 0").Build()
//
// ⚠️ For subqueries, use Sub() method (safer and more flexible):
//
//	// ❌ Not recommended: handwritten subquery
//	.X("user_id IN (SELECT id FROM vip_users)")
//
//	// ✅ Recommended: use Sub()
//	.Sub("user_id IN ?", func(sb *BuilderX) {
//	    sb.Of(&VipUser{}).Select("id")
//	})
func (cb *CondBuilder) X(k string, vs ...interface{}) *CondBuilder {
	bb := Bb{
		Op:    XX,
		Key:   k,
		Value: vs,
	}
	cb.bbs = append(cb.bbs, bb)
	return cb
}
