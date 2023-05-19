package utils

import (
	"regexp"
)

// Returns a string with all non-word characters removed.
func ReplaceSymbols(s string) string {
	m := regexp.MustCompile("[^a-zA-Z0-9]")
	return m.ReplaceAllString(s, "")
}
