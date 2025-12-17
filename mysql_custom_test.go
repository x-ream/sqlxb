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

// Test user model (using MySQLUser to avoid conflicts with other tests)
type MySQLUser struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (MySQLUser) TableName() string {
	return "users"
}

// ============================================================================
// Basic test
// ============================================================================

func TestMySQLCustom_ImplementsCustomInterface(t *testing.T) {
	var custom Custom = DefaultMySQLCustom()
	if custom == nil {
		t.Error("MySQLCustom should implement Custom interface")
	}
	t.Log("✅ MySQLCustom implements Custom interface")
}

func TestMySQLCustom_DefaultBehavior(t *testing.T) {
	custom := DefaultMySQLCustom()

	built := Of(&MySQLUser{}).
		Custom(custom).
		Eq("name", "张三").
		Gt("age", 18).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	if !strings.Contains(sql, "WHERE name = ? AND age > ?") {
		t.Errorf("Expected MySQL placeholder (?), got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ Default MySQL Custom works")
}

// ============================================================================
// UPSERT test
// ============================================================================

func TestMySQLCustom_Upsert(t *testing.T) {
	custom := NewMySQLBuilder().UseUpsert(true).Build()

	user := &MySQLUser{Name: "张三", Age: 18}
	built := Of(user).
		Custom(custom).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", user.Name).Set("age", user.Age)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify contains ON DUPLICATE KEY UPDATE
	if !strings.Contains(sql, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Expected ON DUPLICATE KEY UPDATE, got: %s", sql)
	}

	// Verify contains VALUES(name), VALUES(age)
	if !strings.Contains(sql, "VALUES(name)") || !strings.Contains(sql, "VALUES(age)") {
		t.Errorf("Expected VALUES() clause, got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ MySQL UPSERT works")
}

// TestBuilt_SqlOfUpsert tests SqlOfUpsert() convenience method (no Custom needed)
func TestBuilt_SqlOfUpsert(t *testing.T) {
	user := &MySQLUser{Name: "LiSi", Age: 25}
	built := Of(user).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 100).
				Set("name", user.Name).
				Set("age", user.Age)
		}).
		Build()

	sql, args := built.SqlOfUpsert()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify UPSERT syntax
	if !strings.Contains(sql, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Expected ON DUPLICATE KEY UPDATE in SQL")
	}

	if !strings.Contains(sql, "name = VALUES(name)") {
		t.Errorf("Expected name = VALUES(name) in SQL")
	}

	// id should be skipped (not in UPDATE clause)
	if strings.Contains(sql, "id = VALUES(id)") {
		t.Errorf("Should not update id field in UPSERT")
	}

	if len(args) != 3 { // id, name, age
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	t.Log("✅ SqlOfUpsert() convenience method works")
}

// ============================================================================
// INSERT IGNORE test
// ============================================================================

func TestMySQLCustom_InsertIgnore(t *testing.T) {
	custom := NewMySQLBuilder().UseIgnore(true).Build()

	user := &MySQLUser{Name: "张三", Age: 18}
	built := Of(user).
		Custom(custom).
		Insert(func(ib *InsertBuilder) {
			ib.Set("name", user.Name).Set("age", user.Age)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify contains INSERT IGNORE
	if !strings.Contains(sql, "INSERT IGNORE") {
		t.Errorf("Expected INSERT IGNORE, got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ MySQL INSERT IGNORE works")
}

// ============================================================================
// Singleton test
// ============================================================================

func TestDefaultMySQLCustom_Singleton(t *testing.T) {
	custom1 := DefaultMySQLCustom()
	custom2 := DefaultMySQLCustom()

	if custom1 != custom2 {
		t.Error("DefaultMySQLCustom should return singleton")
	}

	t.Log("✅ DefaultMySQLCustom is singleton")
}

// ============================================================================
// Comparison test with default behavior
// ============================================================================

func TestMySQLCustom_VsDefaultBehavior(t *testing.T) {
	// Default behavior (no Custom set)
	built1 := Of(&MySQLUser{}).
		Eq("name", "张三").
		Build()

	sql1, args1, _ := built1.SqlOfSelect()

	// Using MySQLCustom
	built2 := Of(&MySQLUser{}).
		Custom(DefaultMySQLCustom()).
		Eq("name", "张三").
		Build()

	sql2, args2, _ := built2.SqlOfSelect()

	t.Logf("Default SQL: %s", sql1)
	t.Logf("MySQL Custom SQL: %s", sql2)

	// Both should be consistent (because MySQLCustom's default behavior is standard SQL)
	if sql1 != sql2 {
		t.Errorf("SQL should be the same, got:\nDefault: %s\nMySQL: %s", sql1, sql2)
	}

	if len(args1) != len(args2) {
		t.Errorf("Args count should be the same")
	}

	t.Log("✅ MySQLCustom default behavior matches default SQL generation")
}

// ============================================================================
// Complex scenario test
// ============================================================================

func TestMySQLCustom_ComplexQuery(t *testing.T) {
	custom := NewMySQLBuilder().Build()

	built := Of(&MySQLUser{}).
		Custom(custom).
		Eq("name", "张三").
		Gt("age", 18).
		In("status", "active", "pending").
		Paged(func(pb *PageBuilder) {
			pb.Page(2).Rows(10)
		}).
		Build()

	countSQL, dataSQL, args, _ := built.SqlOfPage()

	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)

	// Verify pagination
	if !strings.Contains(dataSQL, "LIMIT 10 OFFSET 10") {
		t.Errorf("Expected LIMIT/OFFSET, got: %s", dataSQL)
	}

	t.Log("✅ MySQL Custom handles complex query")
}

// ============================================================================
// Update/Delete test
// ============================================================================

func TestMySQLCustom_Update(t *testing.T) {
	custom := NewMySQLBuilder().Build()

	user := &MySQLUser{Name: "李四", Age: 20}
	built := Of(user).
		Custom(custom).
		Update(func(ub *UpdateBuilder) {
			ub.Set("name", user.Name).Set("age", user.Age)
		}).
		Eq("id", 123).
		Build()

	sql, args := built.SqlOfUpdate()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	if !strings.Contains(sql, "UPDATE users SET") {
		t.Errorf("Expected UPDATE, got: %s", sql)
	}

	// args: [李四, 20, 123]
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}

	t.Log("✅ MySQL Custom handles UPDATE")
}

func TestMySQLCustom_Delete(t *testing.T) {
	// ⭐ MySQL DELETE syntax is consistent with default, no Custom needed
	// 测试不设置 Custom 的 DELETE
	built := Of(&MySQLUser{}).
		Eq("id", 123).
		Build()

	sql, args := built.SqlOfDelete()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	if !strings.Contains(sql, "DELETE FROM users") {
		t.Errorf("Expected DELETE, got: %s", sql)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}

	t.Log("✅ MySQL DELETE works (no Custom needed)")
}

// ============================================================================
// Preset mode comparison test
// ============================================================================

func TestMySQLCustom_PresetModes(t *testing.T) {
	tests := []struct {
		name   string
		custom *MySQLCustom
	}{
		{"Default", NewMySQLBuilder().Build()},
		{"WithUpsert", NewMySQLBuilder().UseUpsert(true).Build()},
		{"WithIgnore", NewMySQLBuilder().UseIgnore(true).Build()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &MySQLUser{Name: "test"}
			built := Of(user).
				Custom(tt.custom).
				Insert(func(ib *InsertBuilder) {
					ib.Set("name", user.Name)
				}).
				Build()

			sql, _ := built.SqlOfInsert()
			t.Logf("%s SQL: %s", tt.name, sql)

			// 验证生成的 SQL 是有效的
			if !strings.Contains(sql, "INSERT") {
				t.Errorf("Invalid SQL: %s", sql)
			}
		})
	}

	t.Log("✅ All preset modes work")
}

// ============================================================================
// Meta Map test (field alias mapping) - business scenario
// ============================================================================

func TestMySQLCustom_MetaMap_WithTablePrefix(t *testing.T) {
	custom := NewMySQLBuilder().Build()

	// ⭐ Business scenario: field queries with table prefix
	// Meta map is the field mapping reference when ORM scanning results
	built := Of("users").
		As("u").
		Custom(custom).
		Select("u.id", "u.name", "u.age", "COUNT(*) AS total").
		Eq("u.status", "active").
		Paged(func(pb *PageBuilder) {
			pb.Page(2).Rows(10)
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v (len=%d)", meta, len(meta))

	// ⭐ Verify Meta contains field mappings
	// u.id → auto-generated alias c0
	// u.name → auto-generated alias c1
	// u.age → auto-generated alias c2
	// COUNT(*) AS total → total
	if len(meta) == 0 {
		t.Error("❌ Meta should not be empty for table.field queries")
	} else {
		t.Logf("✅ Meta contains %d mappings", len(meta))
		for k, v := range meta {
			t.Logf("  Meta[%s] = %s", k, v)
		}
	}

	if _, ok := meta["c0"]; !ok {
		t.Error("Expected auto-generated alias 'c0' for u.id")
	}

	if _, ok := meta["total"]; !ok {
		t.Error("Expected alias 'total' for COUNT(*)")
	}

	t.Log("✅ MySQL Custom with Meta map (table prefixes) works")
}

func TestMySQLCustom_MetaMap_WithJoin(t *testing.T) {
	custom := NewMySQLBuilder().Build()

	// ⭐ Business scenario: multi-table JOIN query
	// Meta map records field-to-table mappings, used for ORM scanning
	built := Of("users").
		As("u").
		Custom(custom).
		Select("u.id", "u.name", "o.order_id", "o.amount AS order_amount").
		FromX(func(fb *FromBuilder) {
			fb.JOIN(INNER).Of("orders").As("o").
				On("o.user_id = u.id")
		}).
		Eq("u.status", "active").
		Paged(func(pb *PageBuilder) {
			pb.Page(1).Rows(20)
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("=== JOIN 查询 Meta Map 测试 ===")
	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v (len=%d)", meta, len(meta))

	// ⭐ Verify Meta mappings
	// u.id → c0
	// u.name → c1
	// o.order_id → c2
	// o.amount AS order_amount → order_amount
	if len(meta) < 3 {
		t.Errorf("Expected at least 3 meta mappings, got %d", len(meta))
	}

	t.Log("✅ Meta mappings:")
	for k, v := range meta {
		t.Logf("  Meta[%s] = %s", k, v)
	}

	if meta["c0"] != "u.id" {
		t.Errorf("Expected meta[c0] = u.id, got %s", meta["c0"])
	}

	if meta["order_amount"] != "order_amount" {
		t.Errorf("Expected meta[order_amount] = order_amount, got %s", meta["order_amount"])
	}

	if !strings.Contains(dataSQL, "LIMIT 20") {
		t.Errorf("Expected LIMIT in MySQL query, got: %s", dataSQL)
	}

	t.Log("✅ MySQL Custom with JOIN and Meta map works perfectly")
}

func TestMySQLCustom_MetaMap_WithSelect(t *testing.T) {
	custom := NewMySQLBuilder().Build()

	built := Of("users").
		Custom(custom).
		Select("id", "name", "email AS user_email").
		Eq("status", "active").
		Build()

	sql, args, meta := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v (len=%d)", meta, len(meta))

	// Verify Meta
	if meta["user_email"] != "user_email" {
		t.Errorf("Expected meta[user_email] = user_email, got %s", meta["user_email"])
	}

	t.Log("✅ Meta mappings:")
	for k, v := range meta {
		t.Logf("  Meta[%s] = %s", k, v)
	}

	t.Log("✅ MySQL Custom with Meta map (SELECT) works")
}
