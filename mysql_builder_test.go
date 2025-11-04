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

// TestMySQLBuilder_Upsert 测试 MySQLBuilder 构建 Upsert 配置
func TestMySQLBuilder_Upsert(t *testing.T) {
	// ✅ 使用 MySQLBuilder 构建 Custom
	user := &MySQLUser{Name: "张三", Age: 18}
	built := Of(user).
		Custom(
			NewMySQLBuilder().
				UseUpsert(true).
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "张三").
				Set("age", 18)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("=== MySQLBuilder Upsert ===\nSQL: %s\nArgs: %v", sql, args)

	// 验证 SQL 包含 ON DUPLICATE KEY UPDATE
	if !contains(sql, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Expected ON DUPLICATE KEY UPDATE in SQL, got: %s", sql)
	}

	t.Logf("✅ MySQLBuilder Upsert works correctly")
}

// TestMySQLBuilder_Ignore 测试 MySQLBuilder 构建 Ignore 配置
func TestMySQLBuilder_Ignore(t *testing.T) {
	// ✅ 使用 MySQLBuilder 构建 Custom
	user := &MySQLUser{Name: "李四", Age: 20}
	built := Of(user).
		Custom(
			NewMySQLBuilder().
				UseIgnore(true).
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "李四").
				Set("age", 20)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("=== MySQLBuilder Ignore ===\nSQL: %s\nArgs: %v", sql, args)

	// 验证 SQL 包含 INSERT IGNORE
	if !contains(sql, "INSERT IGNORE") {
		t.Errorf("Expected INSERT IGNORE in SQL, got: %s", sql)
	}

	t.Logf("✅ MySQLBuilder Ignore works correctly")
}

// TestMySQLBuilder_ChainStyle 测试 MySQLBuilder 链式调用
func TestMySQLBuilder_ChainStyle(t *testing.T) {
	// ✅ 演示链式调用的流畅性
	user := &MySQLUser{Name: "王五", Age: 25}
	sql, args := Of(user).
		Custom(
			NewMySQLBuilder().
				UseUpsert(true).
				Placeholder("?").
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "王五").
				Set("age", 25)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== Chain Style ===\nSQL: %s\nArgs: %v", sql, args)

	if !contains(sql, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Expected ON DUPLICATE KEY UPDATE in SQL")
	}

	t.Logf("✅ Chain style works perfectly")
}

// TestMySQLBuilder_ConfigReuse 测试 MySQLBuilder 配置复用
func TestMySQLBuilder_ConfigReuse(t *testing.T) {
	// ✅ 配置可以复用
	upsertConfig := NewMySQLBuilder().
		UseUpsert(true).
		Build()

	// 第一次使用
	user1 := &MySQLUser{Name: "用户1", Age: 30}
	sql1, _ := Of(user1).
		Custom(upsertConfig).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "用户1").Set("age", 30)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== First use ===\n%s", sql1)

	// 第二次使用（复用配置）
	user2 := &MySQLUser{Name: "用户2", Age: 35}
	sql2, _ := Of(user2).
		Custom(upsertConfig).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "用户2").Set("age", 35)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== Second use (reused config) ===\n%s", sql2)

	// 两者都应该有 ON DUPLICATE KEY UPDATE
	if !contains(sql1, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("First SQL should have ON DUPLICATE KEY UPDATE")
	}
	if !contains(sql2, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Second SQL should have ON DUPLICATE KEY UPDATE")
	}

	t.Logf("✅ Config reuse works")
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
