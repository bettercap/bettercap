package str

import (
	"strings"
)

// SplitBy splits by a separator a string and returns a
// list of the non empty parts.
func SplitBy(sv string, sep string) []string {
	filtered := make([]string, 0)
	for _, part := range strings.Split(sv, sep) {
		part = Trim(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}

// Comma splits by comma a string and returns a
// list of the non empty parts.
func Comma(csv string) []string {
	return SplitBy(csv, ",")
}
