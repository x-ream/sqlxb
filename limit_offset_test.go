package xb

import (
	"strings"
	"testing"
)

// TestLimit test Limit API
func TestLimit(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Sort("created_at", nil). // DESC
		Limit(10).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL contains LIMIT
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("Expected SQL to contain 'LIMIT 10', got: %s", sql)
	}

	// Verify args
	if len(args) != 1 { // status = 1
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

// TestOffset test Offset API
func TestOffset(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Sort("id", nil).
		Limit(10).
		Offset(5).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL contains LIMIT and OFFSET
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("Expected SQL to contain 'LIMIT 10', got: %s", sql)
	}
	if !strings.Contains(sql, "OFFSET 5") {
		t.Errorf("Expected SQL to contain 'OFFSET 5', got: %s", sql)
	}
}

// TestLimitOnly test only Limit without Offset
func TestLimitOnly(t *testing.T) {
	built := Of(&Product{}).
		Limit(20).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL contains LIMIT but not OFFSET
	if !strings.Contains(sql, "LIMIT 20") {
		t.Errorf("Expected SQL to contain 'LIMIT 20', got: %s", sql)
	}
	if strings.Contains(sql, "OFFSET") {
		t.Errorf("Expected SQL to NOT contain 'OFFSET', got: %s", sql)
	}
}

// TestOffsetOnly test only Offset without Limit
func TestOffsetOnly(t *testing.T) {
	built := Of(&Product{}).
		Offset(10).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL contains OFFSET
	if !strings.Contains(sql, "OFFSET 10") {
		t.Errorf("Expected SQL to contain 'OFFSET 10', got: %s", sql)
	}
}

// TestLimitWithPaged test Limit not conflict with Paged (Paged priority)
func TestLimitWithPaged(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Limit(50). // Should be covered by Paged
		Paged(func(pb *PageBuilder) {
			pb.Page(2).Rows(20)
		}).
		Build()

	_, dataSql, args, _ := built.SqlOfPage()

	t.Logf("Data SQL: %s", dataSql)
	t.Logf("Args: %v", args)

	// Verify using Paged value (LIMIT 20 OFFSET 20), not Limit value
	if !strings.Contains(dataSql, "LIMIT 20") {
		t.Errorf("Expected SQL to contain 'LIMIT 20' (from Paged), got: %s", dataSql)
	}
	if !strings.Contains(dataSql, "OFFSET 20") {
		t.Errorf("Expected SQL to contain 'OFFSET 20' (from Paged), got: %s", dataSql)
	}
}

// TestLimitZeroIgnored test Limit(0) is ignored automatically
func TestLimitZeroIgnored(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Limit(0). // Should be ignored
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL does not contain LIMIT
	if strings.Contains(sql, "LIMIT") {
		t.Errorf("Expected Limit(0) to be ignored, but SQL contains LIMIT: %s", sql)
	}
}

// TestOffsetZeroIgnored test Offset(0) is ignored automatically
func TestOffsetZeroIgnored(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Offset(0). // Should be ignored
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// Verify SQL does not contain OFFSET
	if strings.Contains(sql, "OFFSET") {
		t.Errorf("Expected Offset(0) to be ignored, but SQL contains OFFSET: %s", sql)
	}
}
