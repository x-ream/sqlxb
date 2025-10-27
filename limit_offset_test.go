package sqlxb

import (
	"strings"
	"testing"
)

// TestLimit 测试 Limit API
func TestLimit(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Sort("created_at", nil).  // DESC
		Limit(10).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 包含 LIMIT
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("Expected SQL to contain 'LIMIT 10', got: %s", sql)
	}

	// 验证参数
	if len(args) != 1 { // status = 1
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
}

// TestOffset 测试 Offset API
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

	// 验证 SQL 包含 LIMIT 和 OFFSET
	if !strings.Contains(sql, "LIMIT 10") {
		t.Errorf("Expected SQL to contain 'LIMIT 10', got: %s", sql)
	}
	if !strings.Contains(sql, "OFFSET 5") {
		t.Errorf("Expected SQL to contain 'OFFSET 5', got: %s", sql)
	}
}

// TestLimitOnly 测试只有 Limit 没有 Offset
func TestLimitOnly(t *testing.T) {
	built := Of(&Product{}).
		Limit(20).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 包含 LIMIT 但不包含 OFFSET
	if !strings.Contains(sql, "LIMIT 20") {
		t.Errorf("Expected SQL to contain 'LIMIT 20', got: %s", sql)
	}
	if strings.Contains(sql, "OFFSET") {
		t.Errorf("Expected SQL to NOT contain 'OFFSET', got: %s", sql)
	}
}

// TestOffsetOnly 测试只有 Offset 没有 Limit
func TestOffsetOnly(t *testing.T) {
	built := Of(&Product{}).
		Offset(10).
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 包含 OFFSET
	if !strings.Contains(sql, "OFFSET 10") {
		t.Errorf("Expected SQL to contain 'OFFSET 10', got: %s", sql)
	}
}

// TestLimitWithPaged 测试 Limit 不与 Paged 冲突（Paged 优先）
func TestLimitWithPaged(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Limit(50). // 应该被 Paged 覆盖
		Paged(func(pb *PageBuilder) {
			pb.Page(2).Rows(20)
		}).
		Build()

	_, dataSql, args, _ := built.SqlOfPage()

	t.Logf("Data SQL: %s", dataSql)
	t.Logf("Args: %v", args)

	// 验证使用 Paged 的值（LIMIT 20 OFFSET 20），而不是 Limit 的值
	if !strings.Contains(dataSql, "LIMIT 20") {
		t.Errorf("Expected SQL to contain 'LIMIT 20' (from Paged), got: %s", dataSql)
	}
	if !strings.Contains(dataSql, "OFFSET 20") {
		t.Errorf("Expected SQL to contain 'OFFSET 20' (from Paged), got: %s", dataSql)
	}
}

// TestLimitZeroIgnored 测试 Limit(0) 被自动忽略
func TestLimitZeroIgnored(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Limit(0). // 应该被忽略
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 不包含 LIMIT
	if strings.Contains(sql, "LIMIT") {
		t.Errorf("Expected Limit(0) to be ignored, but SQL contains LIMIT: %s", sql)
	}
}

// TestOffsetZeroIgnored 测试 Offset(0) 被自动忽略
func TestOffsetZeroIgnored(t *testing.T) {
	built := Of(&Product{}).
		Eq("status", 1).
		Offset(0). // 应该被忽略
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// 验证 SQL 不包含 OFFSET
	if strings.Contains(sql, "OFFSET") {
		t.Errorf("Expected Offset(0) to be ignored, but SQL contains OFFSET: %s", sql)
	}
}

