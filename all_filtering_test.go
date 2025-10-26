// Copyright 2020 io.xream.sqlxb
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
package sqlxb

import (
	"testing"
)

// 综合测试：所有过滤机制
func TestAllFiltering_Comprehensive(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 模拟用户输入（很多字段可能为空/0）
	name := ""                  // 空字符串
	category := "electronics"   // 有效值
	minScore := 0.0             // 0
	maxScore := 0.9             // 有效值
	tags := []interface{}{}     // 空数组
	searchTerm := ""            // 空字符串
	role := ""                  // 空字符串
	department := "engineering" // 有效值

	built := Of(&CodeVectorForQdrant{}).
		// 单个条件过滤
		Eq("name", name).              // ⭐ 过滤：空字符串
		Eq("category", category).      // ✅ 保留
		Gt("min_score", minScore).     // ⭐ 过滤：0
		Lt("max_score", maxScore).     // ✅ 保留
		// IN 过滤
		In("tags", tags...).           // ⭐ 过滤：空数组
		// LIKE 过滤
		Like("description", searchTerm). // ⭐ 过滤：空字符串
		// 空 OR 过滤
		Or(func(cb *CondBuilder) {
			cb.Eq("role", role)         // ⭐ 过滤：空字符串
			cb.Gt("level", 0)           // ⭐ 过滤：0
		}). // ⭐ 整个 OR 被过滤
		// 部分有效的 AND
		And(func(cb *CondBuilder) {
			cb.Eq("department", department) // ✅ 保留
			cb.Gt("rank", 0)                // ⭐ 过滤：0
		}). // ✅ AND 保留（有 1 个有效条件）
		VectorSearch("embedding", queryVector, 20).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== 综合过滤测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	t.Logf("Args: %v", args)

	// 验证：应该只有 3 个有效条件
	// 1. category = 'electronics'
	// 2. max_score < 0.9
	// 3. AND (department = 'engineering')

	if containsString(sql, "name") {
		t.Errorf("name='' should be filtered")
	}
	if containsString(sql, "min_score") {
		t.Errorf("min_score=0 should be filtered")
	}
	if containsString(sql, "tags") {
		t.Errorf("empty tags should be filtered")
	}
	if containsString(sql, "description") {
		t.Errorf("empty LIKE should be filtered")
	}
	if containsString(sql, "role") {
		t.Errorf("role in empty OR should be filtered")
	}
	if containsString(sql, "level") {
		t.Errorf("level=0 in empty OR should be filtered")
	}
	if containsString(sql, "rank") {
		t.Errorf("rank=0 should be filtered")
	}

	// 应该包含的条件
	if !containsString(sql, "category") {
		t.Errorf("category should exist")
	}
	if !containsString(sql, "max_score") {
		t.Errorf("max_score should exist")
	}
	if !containsString(sql, "department") {
		t.Errorf("department should exist")
	}

	t.Logf("✅ 所有过滤机制正常工作")
}

// 测试用户场景：动态查询表单
func TestRealWorldScenario_SearchForm(t *testing.T) {
	// 模拟用户搜索表单（很多字段未填）
	type SearchForm struct {
		Keyword    string
		Category   string
		MinPrice   float64
		MaxPrice   float64
		Tags       []string
		Status     string
		StartDate  string
		EndDate    string
		Department string
		Role       string
	}

	// 用户只填了部分字段
	form := SearchForm{
		Keyword:    "laptop",
		Category:   "",         // 未填
		MinPrice:   0,          // 未填
		MaxPrice:   1500.0,     // 已填
		Tags:       []string{}, // 未填
		Status:     "active",
		StartDate:  "",         // 未填
		EndDate:    "",         // 未填
		Department: "sales",
		Role:       "",         // 未填
	}

	// 无需任何判断，直接构建查询
	tags := make([]interface{}, len(form.Tags))
	for i, tag := range form.Tags {
		tags[i] = tag
	}

	builder := Of(&Product{}).
		Like("name", form.Keyword).         // ✅ 保留
		Eq("category", form.Category).      // ⭐ 过滤
		Gte("price", form.MinPrice).        // ⭐ 过滤
		Lte("price", form.MaxPrice).        // ✅ 保留
		In("tag", tags...).                 // ⭐ 过滤
		Eq("status", form.Status).          // ✅ 保留
		And(func(cb *CondBuilder) {
			cb.Gte("created_at", form.StartDate) // ⭐ 过滤
			cb.Lte("created_at", form.EndDate)   // ⭐ 过滤
		}). // ⭐ 整个 AND 被过滤
		Or(func(cb *CondBuilder) {
			cb.Eq("department", form.Department) // ✅ 保留
			cb.Eq("role", form.Role)             // ⭐ 过滤
		}). // ✅ OR 保留（有 1 个有效条件）
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== 真实场景测试：搜索表单 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 应该只包含用户实际填写的字段
	if !containsString(sql, "LIKE") {
		t.Errorf("Keyword LIKE should exist")
	}
	if !containsString(sql, "status") {
		t.Errorf("Status should exist")
	}
	if !containsString(sql, "department") {
		t.Errorf("Department should exist")
	}
	if !containsString(sql, "price") {
		t.Errorf("MaxPrice should exist")
	}

	// 不应该包含未填写的字段
	if containsString(sql, "category") {
		t.Errorf("Empty category should be filtered")
	}
	if containsString(sql, "tag") {
		t.Errorf("Empty tags should be filtered")
	}
	if containsString(sql, "created_at") {
		t.Errorf("Empty date range should be filtered")
	}

	t.Logf("✅ 真实场景测试通过：只查询用户实际填写的字段")
}

// 测试：复杂的时间范围（用户提到的场景）
func TestTimeRangeFiltering(t *testing.T) {
	// 场景 1: 两个时间都为空
	t.Run("Both times empty", func(t *testing.T) {
		startTime := ""
		endTime := ""

		built := Of(&Order{}).
			Eq("status", "active").
			And(func(cb *CondBuilder) {
				cb.Gt("created_at", startTime)
				cb.Lt("created_at", endTime)
			}).
			Build()

		sql, _, _ := built.SqlOfSelect()

		t.Logf("SQL (both empty): %s", sql)

		// 整个 AND 应该被过滤
		if containsString(sql, "created_at") {
			t.Errorf("Empty time range AND should be filtered")
		}
	})

	// 场景 2: 只有开始时间
	t.Run("Only start time", func(t *testing.T) {
		startTime := "2024-01-01"
		endTime := ""

		built := Of(&Order{}).
			Eq("status", "active").
			And(func(cb *CondBuilder) {
				cb.Gt("created_at", startTime)
				cb.Lt("created_at", endTime)
			}).
			Build()

		sql, _, _ := built.SqlOfSelect()

		t.Logf("SQL (only start): %s", sql)

		// 应该只有 created_at > ?
		if !containsString(sql, "created_at >") {
			t.Errorf("Start time condition should exist")
		}
	})

	// 场景 3: 两个时间都有
	t.Run("Both times valid", func(t *testing.T) {
		startTime := "2024-01-01"
		endTime := "2024-12-31"

		built := Of(&Order{}).
			Eq("status", "active").
			And(func(cb *CondBuilder) {
				cb.Gt("created_at", startTime)
				cb.Lt("created_at", endTime)
			}).
			Build()

		sql, args, _ := built.SqlOfSelect()

		t.Logf("SQL (both valid): %s", sql)
		t.Logf("Args: %v", args)

		// 应该有完整的时间范围
		if !containsString(sql, "created_at >") {
			t.Errorf("Start time condition should exist")
		}
		if !containsString(sql, "created_at <") {
			t.Errorf("End time condition should exist")
		}

		// 应该有 3 个参数：status, startTime, endTime
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
	})

	t.Logf("✅ 时间范围过滤测试通过")
}

// 测试：Select 和 GroupBy 过滤
func TestSelectGroupByFiltering(t *testing.T) {
	built := X().
		From("products").
		Select("id", "", "name", "", "price"). // ⭐ 过滤空字符串
		GroupBy("").                            // ⭐ 过滤空字符串
		GroupBy("category").                    // ✅ 保留
		Agg("", "count").                       // ⭐ 过滤空函数名
		Agg("SUM", "price").                    // ✅ 保留
		Build()

	sql, _, _ := built.SqlOfSelect()

	t.Logf("=== Select/GroupBy 过滤测试 ===")
	t.Logf("SQL: %s", sql)

	// SELECT 应该只有 id, name, price
	if !containsString(sql, "SELECT") {
		t.Errorf("SELECT should exist")
	}

	// GROUP BY 应该只有 category
	if !containsString(sql, "GROUP BY") {
		t.Errorf("GROUP BY should exist")
	}
	if !containsString(sql, "category") {
		t.Errorf("category GROUP BY should exist")
	}

	// 应该有 SUM(price)
	if !containsString(sql, "SUM") {
		t.Errorf("SUM agg should exist")
	}

	t.Logf("✅ Select/GroupBy 过滤正常")
}

// 测试：Bool 条件过滤
func TestBoolFiltering(t *testing.T) {
	includeOptional := false
	includeAdvanced := true

	built := Of(&User{}).
		Eq("status", "active").
		Bool(func() bool { return includeOptional }, func(cb *CondBuilder) {
			cb.Eq("optional_field", "value")
		}).
		Bool(func() bool { return includeAdvanced }, func(cb *CondBuilder) {
			cb.Gt("score", 80)
		}).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("=== Bool 条件过滤测试 ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// optional_field 不应该存在
	if containsString(sql, "optional_field") {
		t.Errorf("Optional field should be filtered (includeOptional=false)")
	}

	// score 应该存在
	if !containsString(sql, "score") {
		t.Errorf("Score should exist (includeAdvanced=true)")
	}

	t.Logf("✅ Bool 条件过滤正常")
}

type Product struct {
	Id         int64   `db:"id"`
	Name       string  `db:"name"`
	Category   string  `db:"category"`
	Price      float64 `db:"price"`
	Status     string  `db:"status"`
	Department string  `db:"department"`
}

func (Product) TableName() string {
	return "products"
}

type Order struct {
	Id        int64  `db:"id"`
	Status    string `db:"status"`
	CreatedAt string `db:"created_at"`
}

func (Order) TableName() string {
	return "orders"
}

type User struct {
	Id     int64  `db:"id"`
	Status string `db:"status"`
	Score  int    `db:"score"`
}

func (User) TableName() string {
	return "users"
}

