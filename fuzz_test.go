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

// Fuzz test: test string conditions (Eq, Like, etc.)
func FuzzStringConditions(f *testing.F) {
	// Seed corpus
	f.Add("test", "value")
	f.Add("", "")
	f.Add("user_name", "alice")
	f.Add("email", "test@example.com")
	f.Add("very_long_field_name_that_might_cause_issues", "value")
	f.Add("field", "'; DROP TABLE users; --")

	f.Fuzz(func(t *testing.T, field string, value string) {
		// Test should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v (field=%q, value=%q)", r, field, value)
			}
		}()

		// Test Eq
		builder := Of(&FuzzTestStruct{}).Eq(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// Test Ne
		builder = Of(&FuzzTestStruct{}).Ne(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// Test Like
		builder = Of(&FuzzTestStruct{}).Like(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// Test LikeLeft
		builder = Of(&FuzzTestStruct{}).LikeLeft(field, value)
		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz test: test numeric conditions (Gt, Lt, etc.)
func FuzzNumericConditions(f *testing.F) {
	// Seed corpus
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

		// Test Gt
		builder := Of(&FuzzTestStruct{}).Gt(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// Test Gte
		builder = Of(&FuzzTestStruct{}).Gte(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// Test Lt
		builder = Of(&FuzzTestStruct{}).Lt(field, value)
		_, _, _ = builder.Build().SqlOfCond()

		// Test Lte
		builder = Of(&FuzzTestStruct{}).Lte(field, value)
		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz test: test Limit/Offset
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

// Fuzz test: test X() hardcoded conditions
func FuzzXCondition(f *testing.F) {
	f.Add("id > ?", int64(100))
	f.Add("status = ?", int64(1))
	f.Add("", int64(0))
	f.Add("created_at BETWEEN ? AND ?", int64(123))

	f.Fuzz(func(t *testing.T, condition string, value int64) {
		defer func() {
			if r := recover(); r != nil {
				// X() allows any string, should not panic
				t.Errorf("Unexpected panic: %v (condition=%q, value=%d)", r, condition, value)
			}
		}()

		builder := Of(&FuzzTestStruct{}).X(condition, value)
		_, _, _ = builder.Build().SqlOfCond()
	})
}

// Fuzz test helper structure
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
