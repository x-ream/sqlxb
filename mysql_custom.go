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

import "strings"

// ============================================================================
// MySQLBuilder: Builder Pattern Configuration Builder
// ============================================================================

// MySQLBuilder MySQL configuration builder
// Uses Builder pattern to construct MySQLCustom configuration
type MySQLBuilder struct {
	custom *MySQLCustom
}

// NewMySQLBuilder creates a MySQL configuration builder
//
// Example:
//
//	xb.Of(...).Custom(
//	    xb.NewMySQLBuilder().
//	        UseUpsert(true).
//	        UseIgnore(false).
//	        Build(),
//	).Build()
func NewMySQLBuilder() *MySQLBuilder {
	return &MySQLBuilder{
		custom: newMySQLCustom(),
	}
}

// UseUpsert sets whether to use ON DUPLICATE KEY UPDATE
func (mb *MySQLBuilder) UseUpsert(use bool) *MySQLBuilder {
	mb.custom.UseUpsert = use
	return mb
}

// UseIgnore sets whether to use INSERT IGNORE
func (mb *MySQLBuilder) UseIgnore(use bool) *MySQLBuilder {
	mb.custom.UseIgnore = use
	return mb
}

// Placeholder sets the placeholder
func (mb *MySQLBuilder) Placeholder(placeholder string) *MySQLBuilder {
	mb.custom.Placeholder = placeholder
	return mb
}

// Build constructs and returns MySQLCustom configuration
func (mb *MySQLBuilder) Build() *MySQLCustom {
	return mb.custom
}

// ============================================================================
// MySQLCustom: MySQL/MariaDB-Specific Configuration (v1.1.0)
// ============================================================================

// MySQLCustom MySQL database-specific configuration
//
// Notes:
//   - xb defaults to MySQL-compatible SQL syntax (? placeholder, LIMIT/OFFSET)
//   - PostgreSQL/SQLite drivers automatically convert placeholders (? → $1, $2)
//   - Most scenarios don't need Custom, use default implementation directly
//
// Use cases:
//   - MySQL special syntax (ON DUPLICATE KEY UPDATE, INSERT IGNORE)
//   - Performance optimization (STRAIGHT_JOIN, FORCE INDEX)
//   - Extended functionality (user-defined)
//
// Example:
//
//	// Default configuration (using singleton)
//	built := xb.Of("users").Custom(xb.DefaultMySQLCustom()).Build()
//
//	// Custom configuration (using Builder)
//	custom := xb.NewMySQLBuilder().UseUpsert(true).Build()
//	built := xb.Of("users").Custom(custom).Build()
type MySQLCustom struct {
	// UseUpsert uses ON DUPLICATE KEY UPDATE (MySQL-specific)
	UseUpsert bool

	// UseIgnore uses INSERT IGNORE (ignores duplicate key errors)
	UseIgnore bool

	// Placeholder placeholder (default "?", compatible with MySQL/PostgreSQL/SQLite)
	Placeholder string
}

// ============================================================================
// Constructors
// ============================================================================

// newMySQLCustom internal function: creates default MySQL Custom
func newMySQLCustom() *MySQLCustom {
	return &MySQLCustom{
		Placeholder: "?", // MySQL placeholder
		UseUpsert:   false,
		UseIgnore:   false,
	}
}

// ============================================================================
// Usage Instructions
// ============================================================================
//
// Configuration methods:
//   1. Using singleton (default configuration): DefaultMySQLCustom()
//   2. Using Builder (recommended):
//      custom := NewMySQLBuilder().UseUpsert(true).Build()
//   3. Or use Built.SqlOfUpsert() method (recommended):
//      built := xb.Of(user).Insert(func(ib *InsertBuilder) {
//          ib.Set("name", "张三")
//      }).Build()
//      sql, args := built.SqlOfUpsert()  // Directly generates UPSERT SQL
//

// ============================================================================
// Implements Custom Interface
// ============================================================================

// Generate implements Custom interface
//
// Notes:
//   - Most scenarios use default SQL generation logic
//   - Special scenarios (UPSERT, IGNORE) use MySQL-specific syntax
//
// Parameters:
//   - built: Built object
//
// Returns:
//   - interface{}: *SQLResult
//   - error: error information
func (c *MySQLCustom) Generate(built *Built) (interface{}, error) {
	// ⭐ Insert scenario: may need UPSERT or IGNORE
	if built.Inserts != nil {
		return c.generateInsert(built)
	}

	// ⭐ Update scenario
	if built.Updates != nil {
		vs := []interface{}{}
		km := make(map[string]string)
		sql, _ := built.SqlData(&vs, km)
		return &SQLResult{SQL: sql, Args: vs, Meta: km}, nil
	}

	// ⭐ Select/Delete scenario (uses default implementation)
	// Note: MySQL DELETE syntax is consistent with standard SQL, no special handling needed
	vs := []interface{}{}
	km := make(map[string]string)
	sql, kmp := built.SqlData(&vs, km)
	return &SQLResult{
		SQL:  sql,
		Args: vs,
		Meta: kmp,
	}, nil
}

// ============================================================================
// Internal Implementation
// ============================================================================

// generateInsert generates MySQL INSERT statement
func (c *MySQLCustom) generateInsert(built *Built) (*SQLResult, error) {
	// Use default SQL generation logic
	vs := []interface{}{}
	sql := built.SqlInsert(&vs)

	// ⭐ MySQL special syntax: ON DUPLICATE KEY UPDATE
	if c.UseUpsert {
		sql = c.addUpsertClause(sql, built)
	}

	// ⭐ MySQL special syntax: INSERT IGNORE
	if c.UseIgnore {
		sql = c.addIgnoreClause(sql)
	}

	return &SQLResult{
		SQL:  sql,
		Args: vs,
	}, nil
}

// addUpsertClause adds ON DUPLICATE KEY UPDATE clause
func (c *MySQLCustom) addUpsertClause(sql string, built *Built) string {
	if built.Inserts == nil || len(*built.Inserts) == 0 {
		return sql
	}

	// Build ON DUPLICATE KEY UPDATE clause
	sql += "\nON DUPLICATE KEY UPDATE "

	inserts := *built.Inserts
	for i, bb := range inserts {
		sql += bb.Key + " = VALUES(" + bb.Key + ")"
		if i < len(inserts)-1 {
			sql += ", "
		}
	}

	return sql
}

// addIgnoreClause adds IGNORE keyword
func (c *MySQLCustom) addIgnoreClause(sql string) string {
	// Insert IGNORE after INSERT
	// INSERT INTO ... → INSERT IGNORE INTO ...
	return "INSERT IGNORE" + sql[6:] // Skip "INSERT"
}

// ============================================================================
// Default MySQL Custom (Global Singleton)
// ============================================================================

// defaultMySQLCustom default MySQL Custom instance
var defaultMySQLCustom = newMySQLCustom()

// DefaultMySQLCustom gets default MySQL Custom (singleton)
//
// Notes:
//   - Returns global singleton to avoid repeated creation
//   - xb defaults to MySQL-compatible SQL syntax, most scenarios don't need to set Custom
//   - Only use when MySQL special syntax (UPSERT, INSERT IGNORE) is needed
//
// Returns:
//   - *MySQLCustom
//
// Example:
//
//	// ⭐ Default scenario (recommended, most concise)
//	built := xb.Of("users").Eq("id", 1).Build()
//	sql, args, _ := built.SqlOfSelect()
//	// SELECT * FROM users WHERE id = ?
//
//	// ⭐ 需要 UPSERT 时
//	built := xb.Of("users").
//	    Custom(xb.NewMySQLBuilder().UseUpsert(true).Build()).
//	    Insert(func(ib *xb.InsertBuilder) {
//	        ib.Set("name", "张三").Set("age", 18)
//	    }).
//	    Build()
//	sql, args := built.SqlOfInsert()
//	// INSERT INTO users (name, age) VALUES (?, ?)
//	// ON DUPLICATE KEY UPDATE name = VALUES(name), age = VALUES(age)
func DefaultMySQLCustom() *MySQLCustom {
	return defaultMySQLCustom
}

// ============================================================================
// Built Convenience Methods (No Custom Required)
// ============================================================================

// SqlOfUpsert generates MySQL UPSERT SQL (INSERT ... ON DUPLICATE KEY UPDATE)
//
// Notes:
//   - No need to set Custom, call directly
//   - Automatically generates ON DUPLICATE KEY UPDATE clause
//
// Returns:
//   - string: SQL statement
//   - []interface{}: parameter list
//
// 示例：
//
//	built := xb.Of(user).Insert(func(ib *InsertBuilder) {
//	    ib.Set("id", 1).Set("name", "张三").Set("age", 18)
//	}).Build()
//
//	sql, args := built.SqlOfUpsert()
//	// INSERT INTO users (id, name, age) VALUES (?, ?, ?)
//	// ON DUPLICATE KEY UPDATE name = VALUES(name), age = VALUES(age)
func (built *Built) SqlOfUpsert() (string, []interface{}) {
	if built.Inserts == nil || len(*built.Inserts) == 0 {
		return "", nil
	}

	vs := []interface{}{}
	sql := built.SqlInsert(&vs)

	// Add ON DUPLICATE KEY UPDATE
	sql += " ON DUPLICATE KEY UPDATE "

	inserts := *built.Inserts
	updateParts := []string{}
	for _, bb := range inserts {
		// Skip primary key (usually id)
		if bb.Key == "id" {
			continue
		}
		updateParts = append(updateParts, bb.Key+" = VALUES("+bb.Key+")")
	}

	sql += strings.Join(updateParts, ", ")

	return sql, vs
}

// SqlOfInsertIgnore generates MySQL INSERT IGNORE SQL
//
// Notes:
//   - No need to set Custom, call directly
//   - Ignores duplicate key errors, doesn't throw exceptions
//
// Returns:
//   - string: SQL statement
//   - []interface{}: parameter list
//
// Example:
//
//	built := xb.Of(user).Insert(func(ib *InsertBuilder) {
//	    ib.Set("id", 1).Set("name", "张三")
//	}).Build()
//
//	sql, args := built.SqlOfInsertIgnore()
//	// INSERT IGNORE INTO users (id, name) VALUES (?, ?)
func (built *Built) SqlOfInsertIgnore() (string, []interface{}) {
	if built.Inserts == nil || len(*built.Inserts) == 0 {
		return "", nil
	}

	vs := []interface{}{}
	sql := built.SqlInsert(&vs)

	// Replace INSERT with INSERT IGNORE
	sql = strings.Replace(sql, "INSERT INTO", "INSERT IGNORE INTO", 1)

	return sql, vs
}
