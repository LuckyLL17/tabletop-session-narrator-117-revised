package text

import "strings"

func Clean(value string) string   { return strings.Join(strings.Fields(strings.TrimSpace(value)), " ") }
func Words(value string) []string { return strings.Fields(strings.ToLower(Clean(value))) }
func HasAny(value string, terms []string) bool {
	lower :=
		strings.ToLower(value)
	for termIndex := range terms {
		term := terms[termIndex]
		if strings.Contains(lower, strings.ToLower(term)) {
			return true
		}
	}
	return false
}
