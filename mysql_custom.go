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
// MySQLCustom：MySQL/MariaDB 专属配置（v1.1.0）
// ============================================================================

// MySQLCustom MySQL 数据库专属配置
//
// 说明：
//   - xb 默认使用 MySQL 兼容的 SQL 语法（? 占位符、LIMIT/OFFSET）
//   - PostgreSQL/SQLite 驱动会自动转换占位符（? → $1, $2）
//   - 大部分场景不需要 Custom，直接使用默认实现即可
//
// 使用场景：
//   - MySQL 特殊语法（ON DUPLICATE KEY UPDATE、INSERT IGNORE）
//   - 性能优化（STRAIGHT_JOIN、FORCE INDEX）
//   - 扩展功能（用户自定义）
//
// 示例：
//
//	// 默认配置
//	built := xb.Of("users").Custom(xb.NewMySQLCustom()).Build()
//
//	// 自定义配置
//	custom := &xb.MySQLCustom{
//	    UseUpsert: true,  // 启用 ON DUPLICATE KEY UPDATE
//	}
//	built := xb.Of("users").Custom(custom).Build()
type MySQLCustom struct {
	// UseUpsert 使用 ON DUPLICATE KEY UPDATE（MySQL 特有）
	UseUpsert bool

	// UseIgnore 使用 INSERT IGNORE（忽略重复键错误）
	UseIgnore bool

	// Placeholder 占位符（默认 "?"，兼容 MySQL/PostgreSQL/SQLite）
	Placeholder string
}

// ============================================================================
// 构造函数
// ============================================================================

// NewMySQLCustom 创建默认 MySQL Custom
//
// 默认配置：
//   - Placeholder: "?"（MySQL 兼容，PostgreSQL 驱动自动转换）
//   - UseUpsert: false
//   - UseIgnore: false
//
// 返回：
//   - *MySQLCustom
//
// 示例：
//
//	custom := xb.NewMySQLCustom()
//	built := xb.Of("users").Custom(custom).Build()
func NewMySQLCustom() *MySQLCustom {
	return &MySQLCustom{
		Placeholder: "?", // MySQL 占位符
		UseUpsert:   false,
		UseIgnore:   false,
	}
}

// ============================================================================
// 使用说明
// ============================================================================
//
// 配置方式：
//   1. 直接创建并使用默认值：NewMySQLCustom()
//   2. 手动设置字段启用功能：
//      custom := NewMySQLCustom()
//      custom.UseUpsert = true  // 启用 UPSERT
//   3. 或者使用 Built.SqlOfUpsert() 方法（推荐）：
//      built := xb.Of(user).Insert(func(ib *InsertBuilder) {
//          ib.Set("name", "张三")
//      }).Build()
//      sql, args := built.SqlOfUpsert()  // 直接生成 UPSERT SQL
//

// ============================================================================
// 实现 Custom 接口
// ============================================================================

// Generate 实现 Custom 接口
//
// 说明：
//   - 大部分场景使用默认 SQL 生成逻辑
//   - 特殊场景（UPSERT、IGNORE）时使用 MySQL 专属语法
//
// 参数：
//   - built: Built 对象
//
// 返回：
//   - interface{}: *SQLResult
//   - error: 错误信息
func (c *MySQLCustom) Generate(built *Built) (interface{}, error) {
	// ⭐ Insert 场景：可能需要 UPSERT 或 IGNORE
	if built.Inserts != nil {
		return c.generateInsert(built)
	}

	// ⭐ Update 场景
	if built.Updates != nil {
		vs := []interface{}{}
		km := make(map[string]string)
		sql, _ := built.SqlData(&vs, km)
		return &SQLResult{SQL: sql, Args: vs, Meta: km}, nil
	}

	// ⭐ Select/Delete 场景（使用默认实现）
	// 注意：MySQL DELETE 语法与标准 SQL 一致，无需特殊处理
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
// 内部实现
// ============================================================================

// generateInsert 生成 MySQL INSERT 语句
func (c *MySQLCustom) generateInsert(built *Built) (*SQLResult, error) {
	// 使用默认 SQL 生成逻辑
	vs := []interface{}{}
	sql := built.SqlInsert(&vs)

	// ⭐ MySQL 特殊语法：ON DUPLICATE KEY UPDATE
	if c.UseUpsert {
		sql = c.addUpsertClause(sql, built)
	}

	// ⭐ MySQL 特殊语法：INSERT IGNORE
	if c.UseIgnore {
		sql = c.addIgnoreClause(sql)
	}

	return &SQLResult{
		SQL:  sql,
		Args: vs,
	}, nil
}

// addUpsertClause 添加 ON DUPLICATE KEY UPDATE 子句
func (c *MySQLCustom) addUpsertClause(sql string, built *Built) string {
	if built.Inserts == nil || len(*built.Inserts) == 0 {
		return sql
	}

	// 构建 ON DUPLICATE KEY UPDATE 子句
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

// addIgnoreClause 添加 IGNORE 关键字
func (c *MySQLCustom) addIgnoreClause(sql string) string {
	// 在 INSERT 后面插入 IGNORE
	// INSERT INTO ... → INSERT IGNORE INTO ...
	return "INSERT IGNORE" + sql[6:] // 跳过 "INSERT"
}

// ============================================================================
// 默认 MySQL Custom（全局单例）
// ============================================================================

// defaultMySQLCustom 默认 MySQL Custom 实例
var defaultMySQLCustom = NewMySQLCustom()

// DefaultMySQLCustom 获取默认 MySQL Custom（单例）
//
// 说明：
//   - 返回全局单例，避免重复创建
//   - xb 默认使用 MySQL 兼容的 SQL 语法，大部分场景无需设置 Custom
//   - 只在需要 MySQL 特殊语法（UPSERT、INSERT IGNORE）时使用
//
// 返回：
//   - *MySQLCustom
//
// 示例：
//
//	// ⭐ 默认场景（推荐，最简洁）
//	built := xb.Of("users").Eq("id", 1).Build()
//	sql, args, _ := built.SqlOfSelect()
//	// SELECT * FROM users WHERE id = ?
//
//	// ⭐ 需要 UPSERT 时
//	built := xb.Of("users").
//	    Custom(xb.MySQLWithUpsert()).
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
// Built 便捷方法（无需 Custom）
// ============================================================================

// SqlOfUpsert 生成 MySQL UPSERT SQL（INSERT ... ON DUPLICATE KEY UPDATE）
//
// 说明：
//   - 无需设置 Custom，直接调用即可
//   - 自动生成 ON DUPLICATE KEY UPDATE 子句
//
// 返回：
//   - string: SQL 语句
//   - []interface{}: 参数列表
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

	// 添加 ON DUPLICATE KEY UPDATE
	sql += " ON DUPLICATE KEY UPDATE "

	inserts := *built.Inserts
	updateParts := []string{}
	for _, bb := range inserts {
		// 跳过主键（通常是 id）
		if bb.Key == "id" {
			continue
		}
		updateParts = append(updateParts, bb.Key+" = VALUES("+bb.Key+")")
	}

	sql += strings.Join(updateParts, ", ")

	return sql, vs
}

// SqlOfInsertIgnore 生成 MySQL INSERT IGNORE SQL
//
// 说明：
//   - 无需设置 Custom，直接调用即可
//   - 忽略重复键错误，不抛异常
//
// 返回：
//   - string: SQL 语句
//   - []interface{}: 参数列表
//
// 示例：
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

	// 将 INSERT 替换为 INSERT IGNORE
	sql = strings.Replace(sql, "INSERT INTO", "INSERT IGNORE INTO", 1)

	return sql, vs
}
