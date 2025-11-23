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

// ============================================================================
// 结果类型定义
// ============================================================================

// SQLResult SQL 查询结果（SQL + 参数）
// 用于 SQL 数据库（PostgreSQL, MySQL, Oracle 等）
type SQLResult struct {
	SQL      string            // Data SQL（带占位符）
	CountSQL string            // Count SQL（可选，用于分页，Oracle/ClickHouse 等需要）
	Args     []interface{}     // 参数值
	Meta     map[string]string // 元数据（可选）
}

// ============================================================================
// Custom 接口：数据库专属配置（核心抽象）
// ============================================================================

// Custom 数据库专属配置接口
// 每个数据库实现自己的 Custom，通过接口多态实现不同的行为
//
// 设计原则（v1.1.0）：
//  - ✅ 统一返回类型：Generate() 返回 interface{}
//  - ✅ 类型灵活：可以是 string（JSON）或 *SQLResult（SQL）
//  - ✅ 性能无损：SQL 不需要包装成 JSON
//  - ✅ 接口极简：一个方法搞定所有数据库
//
// 返回值类型：
//  - string：      向量数据库 JSON（Qdrant/Milvus/Weaviate）
//  - *SQLResult：  SQL 数据库结果（PostgreSQL/Oracle/MySQL）
//
// 实现示例：
//
//	// Qdrant（返回 JSON string）
//	type QdrantCustom struct {
//	    DefaultHnswEf int
//	}
//
//	func (c *QdrantCustom) Generate(built *Built) (interface{}, error) {
//	    json, err := built.toQdrantJSON()
//	    return json, err  // ← 返回 string
//	}
//
//	// Oracle（返回 SQLResult）
//	type OracleCustom struct {
//	    UseRowNum bool
//	}
//
//	func (c *OracleCustom) Generate(built *Built) (interface{}, error) {
//	    sql, args, _ := built.toOracleSQL()
//	    return &SQLResult{SQL: sql, Args: args}, nil  // ← 返回 *SQLResult
//	}
//
// 使用示例：
//
//	// Qdrant
//	built := xb.Of("code_vectors").
//	    Custom(xb.NewQdrantBuilder().Build()).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ← 自动处理类型转换
//
//	// Oracle（示例：未来实现时使用 Builder 模式）
//	// built := xb.Of("users").
//	//     Custom(xb.NewOracleBuilder().Build()).
//	//     Build()
//	//
//	// sql, args, _ := built.SqlOfSelect()  // ← 自动处理类型转换
type Custom interface {
	// Generate 生成查询（统一接口）
	// 参数:
	//   - built: Built 对象（包含所有查询条件）
	// 返回:
	//   - interface{}: string（JSON）或 *SQLResult（SQL + Args）
	//   - error
	//
	// 说明:
	//   - 向量数据库：返回 string（JSON）
	//   - SQL 数据库：返回 *SQLResult（SQL + Args）
	//   - 调用者使用 JsonOfSelect() 或 SqlOfSelect() 自动处理类型转换
	Generate(built *Built) (interface{}, error)
}

// ============================================================================
// 说明和使用场景
// ============================================================================

// Custom 接口极简设计，只需一个 Generate() 方法
//
// 为什么 Generate() 返回 interface{}？
//  - 向量数据库：返回 string（JSON）
//  - SQL 数据库：返回 *SQLResult（SQL + Args）
//  - 未来：可以返回 GraphQL、Protobuf 等任意格式
//
// 为什么所有操作都用同一个方法？
//  - ClickHouse Insert：批量插入，FORMAT JSONEachRow
//  - ClickHouse Update：ALTER TABLE UPDATE（不是标准 UPDATE）
//  - ClickHouse Delete：ALTER TABLE DELETE（不是标准 DELETE）
//  - Oracle 分页：ROWNUM 或 FETCH FIRST（不是 LIMIT/OFFSET）
//  - TimescaleDB：超表特殊语法
//
// 示例：ClickHouse Insert
//
//	type ClickHouseCustom struct {
//	    UseJSONFormat bool
//	}
//
//	func (c *ClickHouseCustom) Generate(built *Built) (interface{}, error) {
//	    // 检查是 Insert 还是 Select
//	    if built.Inserts != nil {
//	        // ClickHouse 批量插入
//	        sql := "INSERT INTO t FORMAT JSONEachRow\n"
//	        return &SQLResult{SQL: sql, Args: nil}, nil
//	    }
//
//	    // ClickHouse 查询
//	    sql, args, _ := built.toSqlOfSelect()
//	    return &SQLResult{SQL: sql, Args: args}, nil
//	}
//
// 示例：Oracle 分页（需要提供 CountSQL）
//
//	type OracleCustom struct {
//	    UseRowNum bool
//	}
//
//	func (c *OracleCustom) Generate(built *Built) (interface{}, error) {
//	    if built.PageCondition != nil {
//	        // Oracle 分页（嵌套查询）
//	        dataSQL := `SELECT * FROM (
//	            SELECT a.*, ROWNUM rn FROM (
//	                SELECT * FROM users WHERE age > ?
//	            ) a WHERE ROWNUM <= 30
//	        ) WHERE rn > 20`
//
//	        // ⭐ 提供独立的 Count SQL
//	        countSQL := "SELECT COUNT(*) FROM users WHERE age > ?"
//
//	        return &SQLResult{
//	            SQL:      dataSQL,
//	            CountSQL: countSQL,  // ⭐ Oracle Custom 负责生成
//	            Args:     []interface{}{18},
//	        }, nil
//	    }
//
//	    // 普通查询
//	    sql, args, _ := built.toSqlOfSelect()
//	    return &SQLResult{SQL: sql, Args: args}, nil
//	}
//
// 这就是 Go 的哲学：简单、直接、实用、灵活
