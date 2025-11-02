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
	"fmt"

	"github.com/fndome/xb"
)

// ============================================================================
// OracleCustom：Oracle 数据库专属配置（v1.1.0）
// ============================================================================

// OracleCustom Oracle 数据库专属配置
//
// 说明：
//   - Oracle 分页语法与 MySQL/PostgreSQL 完全不同（ROWNUM 或 FETCH FIRST）
//   - Oracle 序列语法特殊（NEXTVAL）
//   - xb 默认使用 MySQL 兼容语法，Oracle 需要 Custom
//
// 使用场景：
//   - Oracle 分页（ROWNUM 或 FETCH FIRST）
//   - Oracle 序列（INSERT ... VALUES (seq.NEXTVAL, ...)）
//
// 示例：
//
//	// Oracle 11g 及以下（ROWNUM）
//	built := xb.Of("users").
//	    Custom(oracle_custom.New()).
//	    Eq("age", 18).
//	    Paged(func(pb *xb.PageBuilder) {
//	        pb.Page(3).Rows(20)
//	    }).
//	    Build()
//
//	// Oracle 12c+（FETCH FIRST）
//	built := xb.Of("users").
//	    Custom(oracle_custom.WithFetchFirst()).
//	    Eq("age", 18).
//	    Paged(func(pb *xb.PageBuilder) {
//	        pb.Page(3).Rows(20)
//	    }).
//	    Build()
type OracleCustom struct {
	// UseFetchFirst 使用 FETCH FIRST 语法（Oracle 12c+）
	// false: 使用 ROWNUM 语法（Oracle 11g 及以下）
	UseFetchFirst bool

	// Placeholder 占位符（Oracle 使用 :1, :2 或 ?）
	// 默认使用 ? 以兼容 Oracle 驱动自动转换
	Placeholder string
}

// ============================================================================
// 构造函数
// ============================================================================

// New 创建默认 Oracle Custom（使用 ROWNUM）
//
// 默认配置：
//   - UseFetchFirst: false（使用 ROWNUM，兼容 Oracle 11g）
//   - Placeholder: "?"
//
// 返回：
//   - *OracleCustom
//
// 示例：
//
//	custom := oracle_custom.New()
//	built := xb.Of("users").Custom(custom).Build()
func New() *OracleCustom {
	return &OracleCustom{
		UseFetchFirst: false, // 默认使用 ROWNUM（兼容性最好）
		Placeholder:   "?",
	}
}

// WithFetchFirst 创建使用 FETCH FIRST 的 Oracle Custom（Oracle 12c+）
//
// 说明：
//   - FETCH FIRST 是 SQL 标准语法（Oracle 12c+ 支持）
//   - 性能通常优于 ROWNUM
//   - 语法更简洁
//
// 返回：
//   - *OracleCustom
//
// 示例：
//
//	custom := oracle_custom.WithFetchFirst()
//	built := xb.Of("users").
//	    Custom(custom).
//	    Paged(func(pb *xb.PageBuilder) {
//	        pb.Page(2).Rows(10)
//	    }).
//	    Build()
func WithFetchFirst() *OracleCustom {
	return &OracleCustom{
		UseFetchFirst: true,
		Placeholder:   "?",
	}
}

// WithRowNum 创建使用 ROWNUM 的 Oracle Custom（显式声明）
//
// 说明：
//   - 等价于 New()
//   - 显式声明使用 ROWNUM 语法
//
// 返回：
//   - *OracleCustom
func WithRowNum() *OracleCustom {
	return New()
}

// ============================================================================
// 实现 Custom 接口
// ============================================================================

// Generate 实现 xb.Custom 接口
//
// 说明：
//   - 分页场景：使用 Oracle 特殊的分页语法（ROWNUM 或 FETCH FIRST）
//   - 其他场景：使用默认 SQL 生成逻辑
//
// 参数：
//   - built: xb.Built 对象
//
// 返回：
//   - interface{}: *xb.SQLResult
//   - error: 错误信息
func (c *OracleCustom) Generate(built *xb.Built) (interface{}, error) {
	// ⭐ 分页场景：Oracle 特殊语法
	if built.PageCondition != nil {
		return c.generatePageSQL(built)
	}

	// ⭐ Insert/Update/Select 场景（使用默认实现）
	// 注意：Oracle 的基础 CRUD 语法与标准 SQL 一致
	if built.Inserts != nil {
		vs := []interface{}{}
		sql := built.SqlInsert(&vs)
		return &xb.SQLResult{SQL: sql, Args: vs}, nil
	}

	if built.Updates != nil {
		vs := []interface{}{}
		km := make(map[string]string)
		sql, _ := built.SqlData(&vs, km)
		return &xb.SQLResult{SQL: sql, Args: vs, Meta: km}, nil
	}

	// Select（默认）
	vs := []interface{}{}
	km := make(map[string]string)
	sql, kmp := built.SqlData(&vs, km)
	return &xb.SQLResult{
		SQL:  sql,
		Args: vs,
		Meta: kmp,
	}, nil
}

// ============================================================================
// 内部实现
// ============================================================================

// generatePageSQL 生成 Oracle 分页 SQL
func (c *OracleCustom) generatePageSQL(built *xb.Built) (*xb.SQLResult, error) {
	if built.PageCondition == nil {
		return nil, fmt.Errorf("PageCondition is nil")
	}

	page := built.PageCondition.Page
	rows := built.PageCondition.Rows

	offset := (page - 1) * rows
	limit := rows

	// ⭐ 临时移除 PageCondition，生成不含分页的基础 SQL
	originalPage := built.PageCondition
	built.PageCondition = nil

	// 生成基础查询 SQL（不含分页）
	vs := []interface{}{}
	km := make(map[string]string)
	baseSQL, _ := built.SqlData(&vs, km)

	// ⭐ 恢复 PageCondition
	built.PageCondition = originalPage

	var dataSQL string
	var countSQL string

	if c.UseFetchFirst {
		// ⭐ FETCH FIRST 方式（Oracle 12c+）
		dataSQL = fmt.Sprintf(`%s
OFFSET %d ROWS
FETCH NEXT %d ROWS ONLY`, baseSQL, offset, limit)

		// Count SQL（提取 WHERE 部分）
		countSQL = c.buildCountSQL(built)
	} else {
		// ⭐ ROWNUM 方式（Oracle 11g 及以下）
		dataSQL = fmt.Sprintf(`SELECT * FROM (
  SELECT a.*, ROWNUM rn FROM (
    %s
  ) a WHERE ROWNUM <= %d
) WHERE rn > %d`, baseSQL, offset+limit, offset)

		// Count SQL
		countSQL = c.buildCountSQL(built)
	}

	return &xb.SQLResult{
		SQL:      dataSQL,
		CountSQL: countSQL,
		Args:     vs,
		Meta:     km,
	}, nil
}

// buildCountSQL 生成 Count SQL
func (c *OracleCustom) buildCountSQL(built *xb.Built) string {
	// 使用 built 的 SqlCount() 方法生成 COUNT SQL
	return built.SqlCount()
}

// ============================================================================
// 默认 Oracle Custom（全局单例）
// ============================================================================

// defaultOracleCustom 默认 Oracle Custom 实例
var defaultOracleCustom = New()

// Default 获取默认 Oracle Custom（单例）
//
// 说明：
//   - 返回全局单例，避免重复创建
//   - 默认使用 ROWNUM 语法（兼容 Oracle 11g）
//
// 返回：
//   - *OracleCustom
//
// 示例：
//
//	// 使用默认配置
//	built := xb.Of("users").Custom(oracle_custom.Default()).Build()
//
//	// 或者使用预设模式
//	built := xb.Of("users").Custom(oracle_custom.WithFetchFirst()).Build()
func Default() *OracleCustom {
	return defaultOracleCustom
}

