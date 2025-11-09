package xb

const (
	unionDistinct = "UNION"
	unionAll      = "UNION ALL"
)

// UNION 组合操作符
type UNION func() string

// ALL 返回 UNION ALL 操作符
func ALL() string {
	return unionAll
}

func DISTINCT() string {
	return unionDistinct
}
