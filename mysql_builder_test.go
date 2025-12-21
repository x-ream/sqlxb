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

// TestMySQLBuilder_Upsert test MySQLBuilder build Upsert configuration
func TestMySQLBuilder_Upsert(t *testing.T) {
	// ✅ Use MySQLBuilder to build Custom
	user := &MySQLUser{Name: "John Doe", Age: 18}
	built := Of(user).
		Custom(
			NewMySQLBuilder().
				UseUpsert(true).
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "ZhangSan").
				Set("age", 18)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("=== MySQLBuilder Upsert ===\nSQL: %s\nArgs: %v", sql, args)

	// Verify SQL contains ON DUPLICATE KEY UPDATE
	if !contains(sql, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Expected ON DUPLICATE KEY UPDATE in SQL, got: %s", sql)
	}

	t.Logf("✅ MySQLBuilder Upsert works correctly")
}

// TestMySQLBuilder_Ignore test MySQLBuilder build Ignore configuration
func TestMySQLBuilder_Ignore(t *testing.T) {
	// ✅ Use MySQLBuilder to build Custom
	user := &MySQLUser{Name: "John Doe", Age: 20}
	built := Of(user).
		Custom(
			NewMySQLBuilder().
				UseIgnore(true).
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "John Doe").
				Set("age", 20)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("=== MySQLBuilder Ignore ===\nSQL: %s\nArgs: %v", sql, args)

	// Verify SQL contains INSERT IGNORE
	if !contains(sql, "INSERT IGNORE") {
		t.Errorf("Expected INSERT IGNORE in SQL, got: %s", sql)
	}

	t.Logf("✅ MySQLBuilder Ignore works correctly")
}

// TestMySQLBuilder_ChainStyle test MySQLBuilder chain style
func TestMySQLBuilder_ChainStyle(t *testing.T) {
	// ✅ Demonstrate the smoothness of chain style
	user := &MySQLUser{Name: "John Doe", Age: 25}
	sql, args := Of(user).
		Custom(
			NewMySQLBuilder().
				UseUpsert(true).
				Placeholder("?").
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "John Doe").
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

// TestMySQLBuilder_ConfigReuse test MySQLBuilder config reuse
func TestMySQLBuilder_ConfigReuse(t *testing.T) {
	// ✅ Config can be reused
	upsertConfig := NewMySQLBuilder().
		UseUpsert(true).
		Build()

	// First use
	user1 := &MySQLUser{Name: "John Doe", Age: 30}
	sql1, _ := Of(user1).
		Custom(upsertConfig).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "John Doe").Set("age", 30)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== First use ===\n%s", sql1)

	// Second use (reuse config)
	user2 := &MySQLUser{Name: "John Doe", Age: 35}
	sql2, _ := Of(user2).
		Custom(upsertConfig).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", "John Doe").Set("age", 35)
		}).
		Build().
		SqlOfInsert()

	t.Logf("=== Second use (reused config) ===\n%s", sql2)

	// Both should have ON DUPLICATE KEY UPDATE
	if !contains(sql1, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("First SQL should have ON DUPLICATE KEY UPDATE")
	}
	if !contains(sql2, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Second SQL should have ON DUPLICATE KEY UPDATE")
	}

	t.Logf("✅ Config reuse works")
}

// Helper function
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
