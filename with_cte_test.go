package xb

import (
	"strings"
	"testing"
)

func TestBuilderX_WithClauseBasic(t *testing.T) {
	built := Of("recent_orders").As("ro").
		With("recent_orders", func(sb *BuilderX) {
			sb.From("orders o").
				Select("o.id", "o.user_id").
				Gt("o.created_at", "2025-01-01")
		}).
		Select("ro.id").
		Build()

	sql, args, _ := built.SqlOfSelect()

	t.Logf("sql: %s", sql)
	t.Logf("args: %v", args)

	if !strings.HasPrefix(sql, "WITH ") {
		t.Fatalf("expected SQL to start with WITH, got: %s", sql)
	}
	if !strings.Contains(sql, "recent_orders AS (SELECT o.id AS c0, o.user_id AS c1 FROM orders o WHERE o.created_at > ?)") {
		t.Fatalf("unexpected WITH clause: %s", sql)
	}
	if !strings.Contains(sql, "SELECT ro.id AS c0 FROM recent_orders ro") {
		t.Fatalf("unexpected main SELECT: %s", sql)
	}

	if len(args) != 1 || args[0] != "2025-01-01" {
		t.Fatalf("unexpected args: %v", args)
	}
}

func TestBuilderX_WithRecursiveMixed(t *testing.T) {
	built := Of("team_hierarchy").As("th").
		With("base_users", func(sb *BuilderX) {
			sb.From("users u").
				Select("u.id", "u.manager_id").
				Eq("u.role", "manager")
		}).
		WithRecursive("team_hierarchy", func(sb *BuilderX) {
			sb.From("users u").
				Select("u.id", "u.manager_id").
				Eq("u.active", true)
		}).
		Select("th.id").
		Build()

	sql, args, _ := built.SqlOfSelect()

	if !strings.HasPrefix(sql, "WITH RECURSIVE ") {
		t.Fatalf("expected WITH RECURSIVE prefix, got: %s", sql)
	}

	if !strings.Contains(sql, "base_users AS (SELECT u.id AS c0, u.manager_id AS c1 FROM users u WHERE u.role = ?)") {
		t.Fatalf("missing base_users CTE, sql: %s", sql)
	}
	if !strings.Contains(sql, "team_hierarchy AS (SELECT u.id AS c0, u.manager_id AS c1 FROM users u WHERE u.active = ?)") {
		t.Fatalf("missing team_hierarchy CTE, sql: %s", sql)
	}

	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(args), args)
	}
	if args[0] != "manager" || args[1] != true {
		t.Fatalf("unexpected args order: %v", args)
	}
}
