package textnorm

import "strings"

func LowerTrim(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}
