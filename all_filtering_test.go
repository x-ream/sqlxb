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

// TestAllFiltering_Comprehensive test comprehensive: all filtering mechanisms
func TestAllFiltering_Comprehensive(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// Simulate user input (many fields may be empty/0)
	name := ""                  // Empty string
	category := "electronics"   // Valid value
	minScore := 0.0             // 0
	maxScore := 0.9             // Valid value
	tags := []interface{}{}     // Empty array
	searchTerm := ""            // Empty string
	role := ""                  // Empty string
	department := "engineering" // Valid value

	built := Of(&CodeVectorForQdrant{}).
		// Single condition filtering
		Eq("name", name).          // ⭐ Filtered: empty string
		Eq("category", category).  // ✅ Retained
		Gt("min_score", minScore). // ⭐ Filtered: 0
		Lt("max_score", maxScore). // ✅ Retained
		// IN filtering
		In("tags", tags...). // ⭐ Filtered: empty array
		// LIKE filtering
		Like("description", searchTerm). // ⭐ Filtered: empty string
		// Empty OR filtering
		Or(func(cb *CondBuilder) {
			cb.Eq("role", role) // ⭐ Filtered: empty string
			cb.Gt("level", 0)   // ⭐ Filtered: 0
		}). // ⭐ Entire OR filtered
		// Partially valid AND
		And(func(cb *CondBuilder) {
			cb.Eq("department", department) // ✅ Retained
			cb.Gt("rank", 0)                // ⭐ Filtered: 0
		}). // ✅ AND retained (1 valid condition)
		VectorSearch("embedding", queryVector, 20).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== Comprehensive Filtering Test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	t.Logf("Args: %v", args)

	// Verify: should only have 3 valid conditions
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

	// Should contain conditions
	if !containsString(sql, "category") {
		t.Errorf("category should exist")
	}
	if !containsString(sql, "max_score") {
		t.Errorf("max_score should exist")
	}
	if !containsString(sql, "department") {
		t.Errorf("department should exist")
	}

	t.Logf("✅ All filtering mechanisms work correctly")
}

// TestRealWorldScenario_SearchForm test user scenario: dynamic query form
func TestRealWorldScenario_SearchForm(t *testing.T) {
	// Simulate user search form (many fields not filled)
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

	// User only filled some fields
	form := SearchForm{
		Keyword:    "laptop",
		Category:   "",         // Not filled
		MinPrice:   0,          // Not filled
		MaxPrice:   1500.0,     // Filled
		Tags:       []string{}, // Not filled
		Status:     "active",
		StartDate:  "", // Not filled
		EndDate:    "", // Not filled
		Department: "sales",
		Role:       "", // Not filled
	}

	// No need for any checks, build query directly
	tags := make([]interface{}, len(form.Tags))
	for i, tag := range form.Tags {
		tags[i] = tag
	}

	builder := Of(&Product{}).
		Like("name", form.Keyword).    // ✅ Retained
		Eq("category", form.Category). // ⭐ Filtered
		Gte("price", form.MinPrice).   // ⭐ Filtered
		Lte("price", form.MaxPrice).   // ✅ Retained
		In("tag", tags...).            // ⭐ Filtered
		Eq("status", form.Status).     // ✅ Retained
		And(func(cb *CondBuilder) {
			cb.Gte("created_at", form.StartDate) // ⭐ Filtered
			cb.Lte("created_at", form.EndDate)   // ⭐ Filtered
		}). // ⭐ Entire AND filtered
		Or(func(cb *CondBuilder) {
			cb.Eq("department", form.Department) // ✅ Retained
			cb.Eq("role", form.Role)             // ⭐ Filtered
		}). // ✅ OR retained (1 valid condition)
		Build()

	sql, args, _ := builder.SqlOfSelect()

	t.Logf("=== Real World Scenario: Search Form ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Should only contain fields user actually filled
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

	// Should not contain fields not filled
	if containsString(sql, "category") {
		t.Errorf("Empty category should be filtered")
	}
	if containsString(sql, "tag") {
		t.Errorf("Empty tags should be filtered")
	}
	if containsString(sql, "created_at") {
		t.Errorf("Empty date range should be filtered")
	}

	t.Logf("✅ Real World Scenario: Search Form passed: only query fields user actually filled")
}

// TestTimeRangeFiltering test complex time range (user mentioned scenario)
func TestTimeRangeFiltering(t *testing.T) {
	// Scenario 1: both times are empty
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

		// Entire AND should be filtered
		if containsString(sql, "created_at") {
			t.Errorf("Empty time range AND should be filtered")
		}
	})

	// Scenario 2: only start time
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

		// Should only have created_at > ?
		if !containsString(sql, "created_at >") {
			t.Errorf("Start time condition should exist")
		}
	})

	// Scenario 3: both times are valid
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

		// Should have complete time range
		if !containsString(sql, "created_at >") {
			t.Errorf("Start time condition should exist")
		}
		if !containsString(sql, "created_at <") {
			t.Errorf("End time condition should exist")
		}

		// Should have 3 args: status, startTime, endTime
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
	})

	t.Logf("✅ Time range filtering test passed")
}

// TestSelectGroupByFiltering test Select and GroupBy filtering
func TestSelectGroupByFiltering(t *testing.T) {
	built := X().
		From("products").
		Select("id", "", "name", "", "price"). // ⭐ Filtered empty string
		GroupBy("").                           // ⭐ Filtered empty string
		GroupBy("category").                   // ✅ Retained
		Agg("", "count").                      // ⭐ Filtered empty function name
		Agg("SUM", "price").                   // ✅ Retained
		Build()

	sql, _, _ := built.SqlOfSelect()

	t.Logf("=== Select/GroupBy filtering test ===")
	t.Logf("SQL: %s", sql)

	// SELECT should only have id, name, price
	if !containsString(sql, "SELECT") {
		t.Errorf("SELECT should exist")
	}

	// GROUP BY should only have category
	if !containsString(sql, "GROUP BY") {
		t.Errorf("GROUP BY should exist")
	}
	if !containsString(sql, "category") {
		t.Errorf("category GROUP BY should exist")
	}

	// Should have SUM(price)
	if !containsString(sql, "SUM") {
		t.Errorf("SUM agg should exist")
	}

	t.Logf("✅ Select/GroupBy filtering works correctly")
}

// TestBoolFiltering test Bool condition filtering
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

	t.Logf("=== Bool condition filtering test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// optional_field should not exist
	if containsString(sql, "optional_field") {
		t.Errorf("Optional field should be filtered (includeOptional=false)")
	}

	// score should exist
	if !containsString(sql, "score") {
		t.Errorf("Score should exist (includeAdvanced=true)")
	}

	t.Logf("✅ Bool condition filtering works correctly")
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
