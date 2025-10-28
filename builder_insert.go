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
package xb

import (
	"encoding/json"
	"time"
)

type InsertBuilder struct {
	bbs []Bb
}

func (b *InsertBuilder) Set(k string, v interface{}) *InsertBuilder {

	buffer, ok := v.([]byte)
	if ok {
		b.bbs = append(b.bbs, Bb{
			key:   k,
			value: buffer,
		})
		return b
	}

	defer func() *InsertBuilder {
		if s := recover(); s != nil {
			bytes, _ := json.Marshal(v)
			b.bbs = append(b.bbs, Bb{
				key:   k,
				value: string(bytes),
			})
		}
		return b
	}()

	switch v.(type) {
	case string:
	case uint64, uint, int64, int, int32, int16, int8, bool, byte, float64, float32:
		if v == 0 {
			return b
		}
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return b
		}
		v = n
	case time.Time:
		ts := v.(time.Time).Format("2006-01-02 15:04:05")
		v = ts
	case Vector:
		// Vector 类型：不做处理，保持原样
		// 让 database/sql 调用 driver.Valuer 接口
		// Vector.Value() 会返回正确的数据库格式
	case interface{}:
		bytes, _ := json.Marshal(v)
		v = string(bytes)
	default:
		if v == nil {
			return b
		}
	}

	b.bbs = append(b.bbs, Bb{
		key:   k,
		value: v,
	})
	return b
}
