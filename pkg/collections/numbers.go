package collections

import "sort"

func Sum(
	values []int,
) int {
	total := 0
	for valueIndex := range values {
		value := values[valueIndex]
		total += value
	}
	return total
}
func Average(
	values []int,
) float64 {
	if len(values) == 0 {
		return 0
	}
	return float64(Sum(values)) / float64(len(values))
}
func TopByScore[T any](values []T, score func(T) int, limit int) []T {
	copy := append([]T{}, values...)
	sort.Slice(copy, func(i, j int) bool { return score(copy[i]) > score(copy[j]) })
	if limit > 0 && len(copy) > limit {
		copy = copy[:limit]
	}
	return copy
}
