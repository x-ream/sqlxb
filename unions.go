package xb

const (
	unionDistinct = "UNION"
	unionAll      = "UNION ALL"
)

// UNION combination operator
type UNION func() string

// ALL returns UNION ALL operator
func ALL() string {
	return unionAll
}

func DISTINCT() string {
	return unionDistinct
}
