package collections

import "sort"

func Unique(
	values []string,
) []string {
	seen := map[string]bool{}
	result := []string{}
	for valueIndex := range values {
		value := values[valueIndex]
		if value != "" && !seen[value] {
			seen[value] = true
			result =
				append(result, value)
		}
	}
	return result
}
func Sorted(
	values []string,
) []string {
	copy := append([]string{}, values...)
	sort.Strings(copy)
	return copy
}
func JoinLabels(values []string, separator string) string {
	result := ""
	for index := range values {
		value := values[index]
		if index > 0 {
			result += separator
		}
		result += value
	}
	return result
}
