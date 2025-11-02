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

package oracle_custom

import (
	"strings"
	"testing"

	"github.com/fndome/xb"
)

// 测试用户模型（Oracle）
type OracleUser struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (OracleUser) TableName() string {
	return "users"
}

// ============================================================================
// 基础测试
// ============================================================================

func TestOracleCustom_ImplementsCustomInterface(t *testing.T) {
	var custom xb.Custom = New()
	if custom == nil {
		t.Error("OracleCustom should implement xb.Custom interface")
	}
	t.Log("✅ OracleCustom implements xb.Custom interface")
}

func TestOracleCustom_DefaultBehavior(t *testing.T) {
	custom := New()

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Eq("name", "张三").
		Gt("age", 18).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证基础查询不受影响
	if !strings.Contains(sql, "WHERE name = ? AND age > ?") {
		t.Errorf("Expected standard SQL, got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ Oracle Custom default behavior works")
}

// ============================================================================
// ROWNUM 分页测试
// ============================================================================

func TestOracleCustom_PageWithRowNum(t *testing.T) {
	custom := New() // 默认使用 ROWNUM

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Eq("age", 18).
		Gt("score", 80).
		Paged(func(pb *xb.PageBuilder) {
			pb.Page(3).Rows(20) // 第 3 页，每页 20 条
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v", meta)

	// 验证 ROWNUM 语法
	if !strings.Contains(dataSQL, "ROWNUM") {
		t.Errorf("Expected ROWNUM, got: %s", dataSQL)
	}

	// 验证嵌套查询
	if !strings.Contains(dataSQL, "SELECT * FROM (") {
		t.Errorf("Expected nested SELECT, got: %s", dataSQL)
	}

	// 验证分页参数
	// 第 3 页，每页 20 条：offset = 40, limit = 20
	// ROWNUM <= 60 (40+20)
	// rn > 40
	if !strings.Contains(dataSQL, "ROWNUM <= 60") {
		t.Errorf("Expected ROWNUM <= 60, got: %s", dataSQL)
	}

	if !strings.Contains(dataSQL, "rn > 40") {
		t.Errorf("Expected rn > 40, got: %s", dataSQL)
	}

	// 验证 Count SQL
	if !strings.Contains(countSQL, "COUNT(*)") {
		t.Errorf("Expected COUNT(*), got: %s", countSQL)
	}

	t.Log("✅ Oracle ROWNUM pagination works")
}

// ============================================================================
// FETCH FIRST 分页测试（Oracle 12c+）
// ============================================================================

func TestOracleCustom_PageWithFetchFirst(t *testing.T) {
	custom := WithFetchFirst() // Oracle 12c+

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Eq("age", 18).
		Paged(func(pb *xb.PageBuilder) {
			pb.Page(2).Rows(10) // 第 2 页，每页 10 条
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v", meta)

	// 验证 FETCH FIRST 语法
	if !strings.Contains(dataSQL, "OFFSET") {
		t.Errorf("Expected OFFSET, got: %s", dataSQL)
	}

	if !strings.Contains(dataSQL, "FETCH NEXT") {
		t.Errorf("Expected FETCH NEXT, got: %s", dataSQL)
	}

	// 验证分页参数
	// 第 2 页，每页 10 条：offset = 10
	if !strings.Contains(dataSQL, "OFFSET 10 ROWS") {
		t.Errorf("Expected OFFSET 10 ROWS, got: %s", dataSQL)
	}

	if !strings.Contains(dataSQL, "FETCH NEXT 10 ROWS ONLY") {
		t.Errorf("Expected FETCH NEXT 10 ROWS ONLY, got: %s", dataSQL)
	}

	t.Log("✅ Oracle FETCH FIRST pagination works")
}

// ============================================================================
// 预设模式对比测试
// ============================================================================

func TestOracleCustom_PresetModes(t *testing.T) {
	tests := []struct {
		name   string
		custom *OracleCustom
	}{
		{"Default (ROWNUM)", New()},
		{"FetchFirst", WithFetchFirst()},
		{"RowNum (Explicit)", WithRowNum()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			built := xb.Of(&OracleUser{}).
				Custom(tt.custom).
				Eq("age", 18).
				Paged(func(pb *xb.PageBuilder) {
					pb.Page(2).Rows(10)
				}).
				Build()

			countSQL, dataSQL, args, meta := built.SqlOfPage()

			t.Logf("%s Count SQL: %s", tt.name, countSQL)
			t.Logf("%s Data SQL: %s", tt.name, dataSQL)
			t.Logf("%s Args: %v", tt.name, args)
			t.Logf("%s Meta: %v", tt.name, meta)

			// 验证生成的 SQL 是有效的
			if !strings.Contains(dataSQL, "SELECT") {
				t.Errorf("Invalid SQL: %s", dataSQL)
			}
		})
	}

	t.Log("✅ All Oracle preset modes work")
}

// ============================================================================
// 单例测试
// ============================================================================

func TestDefaultOracleCustom_Singleton(t *testing.T) {
	custom1 := Default()
	custom2 := Default()

	if custom1 != custom2 {
		t.Error("Default() should return singleton")
	}

	t.Log("✅ Default() is singleton")
}

// ============================================================================
// 复杂查询测试
// ============================================================================

func TestOracleCustom_ComplexQuery(t *testing.T) {
	custom := New()

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Eq("name", "张三").
		Gt("age", 18).
		In("status", "active", "pending").
		Paged(func(pb *xb.PageBuilder) {
			pb.Page(3).Rows(15)
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v (len=%d, nil=%v)", meta, len(meta), meta == nil)

	// 详细打印 Meta
	if len(meta) > 0 {
		for k, v := range meta {
			t.Logf("  Meta[%s] = %s", k, v)
		}
	} else {
		t.Log("  ⚠️  Meta is empty - using SELECT * without aliases")
	}

	// 验证复杂条件
	if len(args) < 2 {
		t.Errorf("Expected at least 2 args, got %d", len(args))
	}

	// 验证 ROWNUM 分页
	// 第 3 页，每页 15 条：offset = 30, limit = 15
	// ROWNUM <= 45
	if !strings.Contains(dataSQL, "ROWNUM <= 45") {
		t.Errorf("Expected ROWNUM <= 45, got: %s", dataSQL)
	}

	t.Log("✅ Oracle Custom handles complex query with pagination")
}

// ============================================================================
// 非分页查询测试
// ============================================================================

func TestOracleCustom_NoPage(t *testing.T) {
	custom := New()

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Eq("age", 18).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 非分页查询应该使用默认实现（不含 ROWNUM）
	if strings.Contains(sql, "ROWNUM") {
		t.Errorf("Non-paged query should not contain ROWNUM, got: %s", sql)
	}

	if !strings.Contains(sql, "WHERE age = ?") {
		t.Errorf("Expected WHERE clause, got: %s", sql)
	}

	t.Log("✅ Oracle Custom handles non-paged query correctly")
}

// ============================================================================
// Insert/Update 测试
// ============================================================================

func TestOracleCustom_Insert(t *testing.T) {
	custom := New()

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("name", "张三").Set("age", 18)
		}).
		Build()

	sql, args := built.SqlOfInsert()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Oracle 的基础 INSERT 与标准 SQL 一致
	if !strings.Contains(sql, "INSERT INTO users") {
		t.Errorf("Expected INSERT, got: %s", sql)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	t.Log("✅ Oracle Custom handles INSERT (standard SQL)")
}

func TestOracleCustom_Update(t *testing.T) {
	custom := New()

	built := xb.Of(&OracleUser{}).
		Custom(custom).
		Update(func(ub *xb.UpdateBuilder) {
			ub.Set("name", "李四").Set("age", 20)
		}).
		Eq("id", 123).
		Build()

	sql, args := built.SqlOfUpdate()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Oracle 的基础 UPDATE 与标准 SQL 一致
	if !strings.Contains(sql, "UPDATE users SET") {
		t.Errorf("Expected UPDATE, got: %s", sql)
	}

	t.Log("✅ Oracle Custom handles UPDATE (standard SQL)")
}

// ============================================================================
// Meta Map 测试（字段别名映射）- 业务场景
// ============================================================================

func TestOracleCustom_MetaMap_WithTablePrefix(t *testing.T) {
	custom := New()

	// ⭐ 业务场景：带表前缀的字段查询
	// Meta map 是 ORM 扫描结果时的字段映射依据
	built := xb.Of("users").
		As("u").
		Custom(custom).
		Select("u.id", "u.name", "u.age", "COUNT(*) AS total").
		Eq("u.status", "active").
		Paged(func(pb *xb.PageBuilder) {
			pb.Page(2).Rows(10)
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v (len=%d)", meta, len(meta))

	// ⭐ 验证 Meta 包含字段映射
	// u.id → 自动生成别名 c0
	// u.name → 自动生成别名 c1
	// u.age → 自动生成别名 c2
	// COUNT(*) AS total → total
	if len(meta) == 0 {
		t.Error("❌ Meta should not be empty for table.field queries")
	} else {
		t.Logf("✅ Meta contains %d mappings", len(meta))
		for k, v := range meta {
			t.Logf("  Meta[%s] = %s", k, v)
		}
	}

	// 验证自动生成的别名
	if _, ok := meta["c0"]; !ok {
		t.Error("Expected auto-generated alias 'c0' for u.id")
	}

	if _, ok := meta["total"]; !ok {
		t.Error("Expected alias 'total' for COUNT(*)")
	}

	// 验证 ROWNUM 分页
	if !strings.Contains(dataSQL, "ROWNUM") {
		t.Errorf("Expected ROWNUM, got: %s", dataSQL)
	}

	t.Log("✅ Oracle Custom with Meta map (table prefixes) works")
}

func TestOracleCustom_MetaMap_WithJoin(t *testing.T) {
	custom := New()

	// ⭐ 业务场景：多表 JOIN 查询
	// Meta map 记录字段到表的映射，用于 ORM 扫描
	built := xb.Of("users").
		As("u").
		Custom(custom).
		Select("u.id", "u.name", "o.order_id", "o.amount AS order_amount").
		FromX(func(fb *xb.FromBuilder) {
			fb.JOIN(xb.INNER).Of("orders").As("o").
				On("o.user_id = u.id")
		}).
		Eq("u.status", "active").
		Paged(func(pb *xb.PageBuilder) {
			pb.Page(1).Rows(20)
		}).
		Build()

	countSQL, dataSQL, args, meta := built.SqlOfPage()

	t.Logf("=== JOIN 查询 Meta Map 测试 ===")
	t.Logf("Count SQL: %s", countSQL)
	t.Logf("Data SQL: %s", dataSQL)
	t.Logf("Args: %v", args)
	t.Logf("Meta: %v (len=%d)", meta, len(meta))

	// ⭐ 验证 Meta 映射
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

	// 验证自动生成的别名
	if meta["c0"] != "u.id" {
		t.Errorf("Expected meta[c0] = u.id, got %s", meta["c0"])
	}

	// 验证显式别名
	if meta["order_amount"] != "order_amount" {
		t.Errorf("Expected meta[order_amount] = order_amount, got %s", meta["order_amount"])
	}

	// 验证 ROWNUM 分页正常工作
	if !strings.Contains(dataSQL, "ROWNUM") {
		t.Errorf("Expected ROWNUM in JOIN query, got: %s", dataSQL)
	}

	t.Log("✅ Oracle Custom with JOIN and Meta map works perfectly")
}
