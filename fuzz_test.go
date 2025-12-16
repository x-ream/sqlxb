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
	"testing"
)

// Fuzz 测试：测试字符串条件（Eq, Like 等）
func FuzzStringConditions(f *testing.F) {
	// 种子语料
	f.Add("test", "value")
	f.Add("", "")
	f.Add("user_name", "alice")
	f.Add("email", "test@example.com")
	f.Add("very_long_field_name_that_might_cause_issues", "value")
	f.Add("field", "'; DROP TABLE users; --")

	f.Fuzz(func(t *testing.T, field string, value string) {
		// 测试不应该 panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v (field=%q, value=%q)", r, field, value)
			}
		}()

		// 测试 Eq
		builder := Of(&FuzzTestStruct{}).Eq(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// 测试 Ne
		builder = Of(&FuzzTestStruct{}).Ne(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// 测试 Like
		builder = Of(&FuzzTestStruct{}).Like(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// 测试 LikeLeft
		builder = Of(&FuzzTestStruct{}).LikeLeft(field, value)
		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz 测试：测试数值条件（Gt, Lt 等）
func FuzzNumericConditions(f *testing.F) {
	// 种子语料
	f.Add("age", int64(0))
	f.Add("price", int64(100))
	f.Add("id", int64(-1))
	f.Add("count", int64(9223372036854775807)) // int64 max

	f.Fuzz(func(t *testing.T, field string, value int64) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v (field=%q, value=%d)", r, field, value)
			}
		}()

		// 测试 Gt
		builder := Of(&FuzzTestStruct{}).Gt(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// 测试 Gte
		builder = Of(&FuzzTestStruct{}).Gte(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// 测试 Lt
		builder = Of(&FuzzTestStruct{}).Lt(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// 测试 Lte
		builder = Of(&FuzzTestStruct{}).Lte(field, value)
		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz 测试：测试 Limit/Offset
func FuzzPagination(f *testing.F) {
	f.Add(10, 0)
	f.Add(100, 50)
	f.Add(0, 0)
	f.Add(-1, -1)
	f.Add(1000000, 999999)

	f.Fuzz(func(t *testing.T, limit int, offset int) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v (limit=%d, offset=%d)", r, limit, offset)
			}
		}()

		builder := Of(&FuzzTestStruct{}).
			Limit(limit).
			Offset(offset)

		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz 测试：测试 X() 硬编码条件
func FuzzXCondition(f *testing.F) {
	f.Add("id > ?", int64(100))
	f.Add("status = ?", int64(1))
	f.Add("", int64(0))
	f.Add("created_at BETWEEN ? AND ?", int64(123))

	f.Fuzz(func(t *testing.T, condition string, value int64) {
		defer func() {
			if r := recover(); r != nil {
				// X() 允许任何字符串，不应该 panic
				t.Errorf("Unexpected panic: %v (condition=%q, value=%d)", r, condition, value)
			}
		}()

		builder := Of(&FuzzTestStruct{}).X(condition, value)
		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz 测试辅助结构
type FuzzTestStruct struct {
	ID       uint64   `db:"id"`
	Name     string   `db:"name"`
	Age      *int     `db:"age"`
	Price    *float64 `db:"price"`
	IsActive *bool    `db:"is_active"`
}

func (*FuzzTestStruct) TableName() string {
	return "fuzz_test"
}
