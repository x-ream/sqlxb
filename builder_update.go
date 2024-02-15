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

type UpdateBuilder struct {
	bbs []Bb
}

func (ub *UpdateBuilder) Set(k string, v interface{}) *UpdateBuilder {

	switch v.(type) {
	case string:
		if v.(string) == "" {
			return ub
		}
	case uint64, uint, int64, int, int32, int16, int8, bool, byte, float64, float32:
		if v == 0 {
			return ub
		}
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return ub
		}
		v = n
	case time.Time:
		ts := v.(time.Time).Format("2006-01-02 15:04:05")
		v = ts
	case []interface{}:
		panic("Builder.doGLE(ke, []arr), [] ?")
	default:
		if v == nil {
			return ub
		}
	}

	ub.bbs = append(ub.bbs, Bb{
		op:    "SET",
		key:   k,
		value: v,
	})
	return ub
}

func (ub *UpdateBuilder) X(s string) *UpdateBuilder {
	ub.bbs = append(ub.bbs, Bb{
		op:  "SET",
		key: s,
	})
	return ub
}

func (ub *UpdateBuilder) Any(f func(*UpdateBuilder)) *UpdateBuilder {
	f(ub)
	return ub
}

func (ub *UpdateBuilder) Bool(preCond BoolFunc, f func(cb *UpdateBuilder)) *UpdateBuilder {
	if preCond == nil {
		panic("UpdateBuilder.Bool para of BoolFunc can not nil")
	}
	if !preCond() {
		return ub
	}
	if f == nil {
		panic("UpdateBuilder.Bool para of func(k string, vs... interface{}) can not nil")
	}
	f(ub)
	return ub
}
