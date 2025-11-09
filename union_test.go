package xb

import (
	"strings"
	"testing"
)

func TestBuilderX_UnionAll(t *testing.T) {
	built := Of("users").
		Select("id").
		UNION(ALL, func(sb *BuilderX) {
			sb.From("archived_users").
				Select("id")
		}).
		Build()

	sql, args, _ := built.SqlOfSelect()

	if !strings.Contains(sql, "UNION ALL") {
		t.Fatalf("expected UNION ALL, got: %s", sql)
	}
	if !strings.HasPrefix(sql, "SELECT id FROM users") {
		t.Fatalf("missing base SELECT, sql: %s", sql)
	}
	if !strings.Contains(sql, "SELECT id FROM archived_users") {
		t.Fatalf("missing union SELECT, sql: %s", sql)
	}
	if len(args) != 0 {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestBuilderX_UnionDefaultDistinct(t *testing.T) {
	built := Of("orders").
		Select("id").
		UNION(nil, func(sb *BuilderX) {
			sb.From("orders_history").
				Select("id")
		}).
		Sort("id", ASC).
		Build()

	sql, args, _ := built.SqlOfSelect()

	if strings.Contains(sql, "UNION ALL") {
		t.Fatalf("should not contain UNION ALL, sql: %s", sql)
	}
	if !strings.Contains(sql, "UNION") {
		t.Fatalf("expected UNION keyword, sql: %s", sql)
	}
	if !strings.HasSuffix(sql, "ORDER BY id ASC") {
		t.Fatalf("ORDER BY should apply after UNION, sql: %s", sql)
	}
	if len(args) != 0 {
		t.Fatalf("unexpected args: %v", args)
	}
}
