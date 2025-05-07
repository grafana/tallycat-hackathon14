package duckdb

import "strings"

func buildDuckDBStringArray(values []string) string {
	var builder strings.Builder
	builder.WriteString("ARRAY[")
	for i, val := range values {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("'")
		builder.WriteString(strings.ReplaceAll(val, "'", "''")) // escape single quotes
		builder.WriteString("'")
	}
	builder.WriteString("]")
	return builder.String()
}
