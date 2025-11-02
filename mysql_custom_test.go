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

// 测试用户模型（使用 MySQLUser 避免与其他测试冲突）
type MySQLUser struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (MySQLUser) TableName() string {
	return "users"
}

// ============================================================================
// 基础测试
// ============================================================================

func TestMySQLCustom_ImplementsCustomInterface(t *testing.T) {
	var custom Custom = NewMySQLCustom()
	if custom == nil {
		t.Error("MySQLCustom should implement Custom interface")
	}
	t.Log("✅ MySQLCustom implements Custom interface")
}

func TestMySQLCustom_DefaultBehavior(t *testing.T) {
	custom := NewMySQLCustom()

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
// UPSERT 测试
// ============================================================================

func TestMySQLCustom_Upsert(t *testing.T) {
	custom := MySQLWithUpsert()

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

	// 验证包含 ON DUPLICATE KEY UPDATE
	if !strings.Contains(sql, "ON DUPLICATE KEY UPDATE") {
		t.Errorf("Expected ON DUPLICATE KEY UPDATE, got: %s", sql)
	}

	// 验证包含 VALUES(name), VALUES(age)
	if !strings.Contains(sql, "VALUES(name)") || !strings.Contains(sql, "VALUES(age)") {
		t.Errorf("Expected VALUES() clause, got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ MySQL UPSERT works")
}

// ============================================================================
// INSERT IGNORE 测试
// ============================================================================

func TestMySQLCustom_InsertIgnore(t *testing.T) {
	custom := MySQLWithIgnore()

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

	// 验证包含 INSERT IGNORE
	if !strings.Contains(sql, "INSERT IGNORE") {
		t.Errorf("Expected INSERT IGNORE, got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ MySQL INSERT IGNORE works")
}

// ============================================================================
// 单例测试
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
// 与默认行为对比测试
// ============================================================================

func TestMySQLCustom_VsDefaultBehavior(t *testing.T) {
	// 默认行为（不设置 Custom）
	built1 := Of(&MySQLUser{}).
		Eq("name", "张三").
		Build()

	sql1, args1, _ := built1.SqlOfSelect()

	// 使用 MySQLCustom
	built2 := Of(&MySQLUser{}).
		Custom(NewMySQLCustom()).
		Eq("name", "张三").
		Build()

	sql2, args2, _ := built2.SqlOfSelect()

	t.Logf("Default SQL: %s", sql1)
	t.Logf("MySQL Custom SQL: %s", sql2)

	// 两者应该一致（因为 MySQLCustom 的默认行为就是标准 SQL）
	if sql1 != sql2 {
		t.Errorf("SQL should be the same, got:\nDefault: %s\nMySQL: %s", sql1, sql2)
	}

	if len(args1) != len(args2) {
		t.Errorf("Args count should be the same")
	}

	t.Log("✅ MySQLCustom default behavior matches default SQL generation")
}

// ============================================================================
// 复杂场景测试
// ============================================================================

func TestMySQLCustom_ComplexQuery(t *testing.T) {
	custom := NewMySQLCustom()

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

	// 验证分页
	if !strings.Contains(dataSQL, "LIMIT 10 OFFSET 10") {
		t.Errorf("Expected LIMIT/OFFSET, got: %s", dataSQL)
	}

	t.Log("✅ MySQL Custom handles complex query")
}

// ============================================================================
// Update/Delete 测试
// ============================================================================

func TestMySQLCustom_Update(t *testing.T) {
	custom := NewMySQLCustom()

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
	// ⭐ MySQL DELETE 语法与默认一致，无需 Custom
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
// 预设模式对比测试
// ============================================================================

func TestMySQLCustom_PresetModes(t *testing.T) {
	tests := []struct {
		name   string
		custom *MySQLCustom
	}{
		{"Default", NewMySQLCustom()},
		{"WithUpsert", MySQLWithUpsert()},
		{"WithIgnore", MySQLWithIgnore()},
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

