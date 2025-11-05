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
	"strings"
	"testing"
)

// TestInRequired_EmptyValues 测试空值场景
func TestInRequired_EmptyValues(t *testing.T) {
	tests := []struct {
		name      string
		buildFunc func()
		wantPanic bool
		panicMsg  string
	}{
		{
			name: "Empty variadic args",
			buildFunc: func() {
				Of(&User{}).InRequired("id").Build()
			},
			wantPanic: true,
			panicMsg:  "received empty values",
		},
		{
			name: "Empty slice spread",
			buildFunc: func() {
				ids := []interface{}{}
				Of(&User{}).InRequired("id", ids...).Build()
			},
			wantPanic: true,
			panicMsg:  "received empty values",
		},
		{
			name: "Nil value",
			buildFunc: func() {
				Of(&User{}).InRequired("id", nil).Build()
			},
			wantPanic: true,
			panicMsg:  "received [nil]",
		},
		{
			name: "Zero int",
			buildFunc: func() {
				Of(&User{}).InRequired("id", 0).Build()
			},
			wantPanic: true,
			panicMsg:  "received [0]",
		},
		{
			name: "Zero int64",
			buildFunc: func() {
				Of(&User{}).InRequired("id", int64(0)).Build()
			},
			wantPanic: true,
			panicMsg:  "received [0]",
		},
		{
			name: "Zero uint",
			buildFunc: func() {
				Of(&User{}).InRequired("id", uint(0)).Build()
			},
			wantPanic: true,
			panicMsg:  "received [0]",
		},
		{
			name: "Empty string",
			buildFunc: func() {
				Of(&User{}).InRequired("status", "").Build()
			},
			wantPanic: true,
			panicMsg:  `received [""]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Error("Expected panic but didn't panic")
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
				if r != nil {
					msg := r.(string)
					if !strings.Contains(msg, tt.panicMsg) {
						t.Errorf("Panic message doesn't contain expected text.\nGot: %s\nWant substring: %s", msg, tt.panicMsg)
					}
					if !strings.Contains(msg, "Use In() if optional filtering is intended") {
						t.Error("Panic message should suggest using In() for optional filtering")
					}
					t.Logf("✅ Expected panic: %v", r)
				}
			}()

			tt.buildFunc()
		})
	}
}

// TestInRequired_ValidValues 测试有效值场景
func TestInRequired_ValidValues(t *testing.T) {
	tests := []struct {
		name      string
		buildFunc func() *Built
		wantSQL   string
	}{
		{
			name: "Single valid int",
			buildFunc: func() *Built {
				return Of("users").InRequired("id", 1).Build()
			},
			wantSQL: "WHERE id IN (1)",
		},
		{
			name: "Multiple ints",
			buildFunc: func() *Built {
				return Of("users").InRequired("id", 1, 2, 3).Build()
			},
			wantSQL: "WHERE id IN (1, 2, 3)",
		},
		{
			name: "Multiple strings",
			buildFunc: func() *Built {
				return Of("users").InRequired("status", "active", "pending").Build()
			},
			wantSQL: "WHERE status IN ('active', 'pending')",
		},
		{
			name: "Slice spread",
			buildFunc: func() *Built {
				ids := []interface{}{1, 2, 3}
				return Of("users").InRequired("id", ids...).Build()
			},
			wantSQL: "WHERE id IN (1, 2, 3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			built := tt.buildFunc()
			sql, _, _ := built.SqlOfSelect()

			if !strings.Contains(sql, tt.wantSQL) {
				t.Errorf("SQL doesn't contain expected clause.\nGot: %s\nWant to contain: %s", sql, tt.wantSQL)
			}
			t.Logf("✅ Valid SQL: %s", sql)
		})
	}
}

// TestInRequired_ComparedWithIn 测试 InRequired 与 In 的行为差异
func TestInRequired_ComparedWithIn(t *testing.T) {
	t.Run("In() with empty slice - no panic, no WHERE clause", func(t *testing.T) {
		ids := []interface{}{}
		built := Of("users").In("id", ids...).Build()
		sql, _, _ := built.SqlOfSelect()

		if strings.Contains(sql, "WHERE") {
			t.Errorf("In() with empty slice should not add WHERE clause.\nGot: %s", sql)
		}
		t.Logf("✅ In() behavior: %s", sql)
	})

	t.Run("InRequired() with empty slice - panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("InRequired() with empty slice should panic")
			} else {
				t.Logf("✅ InRequired() panicked as expected: %v", r)
			}
		}()

		ids := []interface{}{}
		Of("users").InRequired("id", ids...).Build()
	})

	t.Run("In() with valid values - same SQL as InRequired()", func(t *testing.T) {
		ids := []interface{}{1, 2, 3}

		builtIn := Of("users").In("id", ids...).Build()
		sqlIn, _, _ := builtIn.SqlOfSelect()

		builtRequired := Of("users").InRequired("id", ids...).Build()
		sqlRequired, _, _ := builtRequired.SqlOfSelect()

		if sqlIn != sqlRequired {
			t.Errorf("In() and InRequired() should generate same SQL for valid values.\nIn():       %s\nInRequired(): %s", sqlIn, sqlRequired)
		}
		t.Logf("✅ Same SQL: %s", sqlIn)
	})
}

// TestInRequired_RealWorldScenario 测试真实场景
func TestInRequired_RealWorldScenario(t *testing.T) {
	t.Run("Scenario: Admin deletes selected orders", func(t *testing.T) {
		// ✅ 用户选择了订单
		selectedOrderIDs := []interface{}{101, 102, 103}
		built := Of("orders").InRequired("id", selectedOrderIDs...).Build()
		sql, _, _ := built.SqlOfSelect()

		if !strings.Contains(sql, "WHERE id IN (101, 102, 103)") {
			t.Errorf("Should generate correct WHERE clause.\nGot: %s", sql)
		}
		t.Logf("✅ Admin deletes selected orders: %s", sql)
	})

	t.Run("Scenario: Admin forgets to select orders - panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("✅ Prevented accidental deletion of all orders: %v", r)
			} else {
				t.Error("Should panic when no orders selected")
			}
		}()

		// ❌ 用户忘记选择订单（前端bug或用户误操作）
		selectedOrderIDs := []interface{}{}
		Of("orders").InRequired("id", selectedOrderIDs...).Build()
		// 如果不报错，会删除所有订单！
	})

	t.Run("Scenario: Query orders by status (optional filter)", func(t *testing.T) {
		// ✅ 使用 In() 实现可选过滤
		status := "" // 用户没有选择状态

		var built *Built
		if status != "" {
			built = Of("orders").In("status", status).Build()
		} else {
			built = Of("orders").Build() // 查询所有状态
		}

		sql, _, _ := built.SqlOfSelect()
		if strings.Contains(sql, "WHERE") && status == "" {
			t.Errorf("Should not have WHERE clause when status is empty.\nGot: %s", sql)
		}
		t.Logf("✅ Optional filter with In(): %s", sql)
	})
}

