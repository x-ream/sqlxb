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
// Result Type Definitions
// ============================================================================

// SQLResult SQL query result (SQL + parameters)
// Used for SQL databases (PostgreSQL, MySQL, Oracle, etc.)
type SQLResult struct {
	SQL      string            // Data SQL (with placeholders)
	CountSQL string            // Count SQL (optional, for pagination, required by Oracle/ClickHouse, etc.)
	Args     []interface{}     // Parameter values
	Meta     map[string]string // Metadata (optional)
}

// ============================================================================
// Custom Interface: Database-Specific Configuration (Core Abstraction)
// ============================================================================

// Custom database-specific configuration interface
// Each database implements its own Custom, achieving different behaviors through interface polymorphism
//
// Design principles (v1.1.0):
//   - ✅ Unified return type: Generate() returns interface{}
//   - ✅ Type flexibility: can be string (JSON) or *SQLResult (SQL)
//   - ✅ Zero performance overhead: SQL doesn't need to be wrapped in JSON
//   - ✅ Minimal interface: one method handles all databases
//
// Return value types:
//   - string:      Vector database JSON (Qdrant/Milvus/Weaviate)
//   - *SQLResult:  SQL database result (PostgreSQL/Oracle/MySQL)
//
// Implementation examples:
//
//	// Qdrant (returns JSON string)
//	type QdrantCustom struct {
//	    DefaultHnswEf int
//	}
//
//	func (c *QdrantCustom) Generate(built *Built) (interface{}, error) {
//	    json, err := built.toQdrantJSON()
//	    return json, err  // ← returns string
//	}
//
//	// Oracle (returns SQLResult)
//	type OracleCustom struct {
//	    UseRowNum bool
//	}
//
//	func (c *OracleCustom) Generate(built *Built) (interface{}, error) {
//	    sql, args, _ := built.toOracleSQL()
//	    return &SQLResult{SQL: sql, Args: args}, nil  // ← returns *SQLResult
//	}
//
// Usage examples:
//
//	// Qdrant
//	built := xb.Of("code_vectors").
//	    Custom(xb.NewQdrantBuilder().Build()).
//	    Build()
//
//	json, _ := built.JsonOfSelect()  // ← automatic type conversion
//
//	// Oracle (example: future implementation using Builder pattern)
//	// built := xb.Of("users").
//	//     Custom(xb.NewOracleBuilder().Build()).
//	//     Build()
//	//
//	// sql, args, _ := built.SqlOfSelect()  // ← automatic type conversion
type Custom interface {
	// Generate generates query (unified interface)
	// Parameters:
	//   - built: Built object (contains all query conditions)
	// Returns:
	//   - interface{}: string (JSON) or *SQLResult (SQL + Args)
	//   - error
	//
	// Notes:
	//   - Vector databases: returns string (JSON)
	//   - SQL databases: returns *SQLResult (SQL + Args)
	//   - Callers use JsonOfSelect() or SqlOfSelect() for automatic type conversion
	Generate(built *Built) (interface{}, error)
}

// ============================================================================
// Notes and Use Cases
// ============================================================================

// Custom interface has minimal design, only one Generate() method
//
// Why does Generate() return interface{}?
//  - Vector databases: returns string (JSON)
//  - SQL databases: returns *SQLResult (SQL + Args)
//  - Future: can return GraphQL, Protobuf, or any other format
//
// Why use the same method for all operations?
//  - ClickHouse Insert: batch insert, FORMAT JSONEachRow
//  - ClickHouse Update: ALTER TABLE UPDATE (not standard UPDATE)
//  - ClickHouse Delete: ALTER TABLE DELETE (not standard DELETE)
//  - Oracle pagination: ROWNUM or FETCH FIRST (not LIMIT/OFFSET)
//  - TimescaleDB: hypertable special syntax
//
// Example: ClickHouse Insert
//
//	type ClickHouseCustom struct {
//	    UseJSONFormat bool
//	}
//
//	func (c *ClickHouseCustom) Generate(built *Built) (interface{}, error) {
//	    // Check if Insert or Select
//	    if built.Inserts != nil {
//	        // ClickHouse batch insert
//	        sql := "INSERT INTO t FORMAT JSONEachRow\n"
//	        return &SQLResult{SQL: sql, Args: nil}, nil
//	    }
//
//	    // ClickHouse query
//	    sql, args, _ := built.toSqlOfSelect()
//	    return &SQLResult{SQL: sql, Args: args}, nil
//	}
//
// Example: Oracle pagination (requires CountSQL)
//
//	type OracleCustom struct {
//	    UseRowNum bool
//	}
//
//	func (c *OracleCustom) Generate(built *Built) (interface{}, error) {
//	    if built.PageCondition != nil {
//	        // Oracle pagination (nested query)
//	        dataSQL := `SELECT * FROM (
//	            SELECT a.*, ROWNUM rn FROM (
//	                SELECT * FROM users WHERE age > ?
//	            ) a WHERE ROWNUM <= 30
//	        ) WHERE rn > 20`
//
//	        // ⭐ Provide independent Count SQL
//	        countSQL := "SELECT COUNT(*) FROM users WHERE age > ?"
//
//	        return &SQLResult{
//	            SQL:      dataSQL,
//	            CountSQL: countSQL,  // ⭐ Oracle Custom is responsible for generation
//	            Args:     []interface{}{18},
//	        }, nil
//	    }
//
//	    // Normal query
//	    sql, args, _ := built.toSqlOfSelect()
//	    return &SQLResult{SQL: sql, Args: args}, nil
//	}
//
// This is Go's philosophy: simple, direct, practical, flexible
