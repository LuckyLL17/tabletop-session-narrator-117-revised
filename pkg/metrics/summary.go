package metrics

import (
	"math"
	"sort"
)

func fold[T any, U any](values []T, initial U, step func(U, T) U) U {
	state := initial
	for position := range values {
		state = step(state, values[position])
	}
	return state
}

func Sum(
	values []int,
) int {
	return fold(values, 0, func(total, value int) int { return total + value })
}

func SumFloat(
	values []float64,
) float64 {
	return fold(values, 0.0, func(total, value float64) float64 { return total + value })
}

func Mean(
	values []int,
) float64 {
	if len(values) == 0 {
		return 0
	}
	return float64(Sum(values)) / float64(len(values))
}

func MeanFloat(
	values []float64,
) float64 {
	if len(values) == 0 {
		return 0
	}
	return SumFloat(values) / float64(len(values))
}

func Median(
	values []int,
) float64 {
	if len(values) == 0 {
		return 0
	}
	copyOf := append([]int(nil), values...)
	sort.Ints(copyOf)
	middle := len(copyOf) / 2
	if len(copyOf)%2 == 1 {
		return float64(copyOf[middle])
	}
	return float64(copyOf[middle-1]+copyOf[middle]) / 2
}

func Min(
	values []int,
) int {
	if len(values) == 0 {
		return 0
	}
	return fold(values[1:], values[0], func(current, value int) int {
		if value < current {
			return value
		}
		return current
	})
}

func Max(
	values []int,
) int {
	if len(values) == 0 {
		return 0
	}
	return fold(values[1:], values[0], func(current, value int) int {
		if value > current {
			return value
		}
		return current
	})
}

func Range(
	values []int,
) int {
	if len(values) == 0 {
		return 0
	}
	return Max(values) - Min(values)
}

func Variance(
	values []int,
) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := Mean(values)
	total := fold(values, 0.0, func(sum float64, value int) float64 {
		distance := float64(value) - mean
		return sum + distance*distance
	})
	return total / float64(len(values))
}

func StandardDeviation(
	values []int,
) float64 {
	return math.Sqrt(Variance(values))
}

func CoefficientOfVariation(
	values []int,
) float64 {
	mean := Mean(values)
	if mean == 0 {
		return 0
	}
	return StandardDeviation(values) / math.Abs(mean)
}

func Percentile(values []int, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if percentile < 0 {
		percentile = 0
	}
	if percentile > 1 {
		percentile = 1
	}
	copyOf := append([]int(nil), values...)
	sort.Ints(copyOf)
	position := percentile * float64(len(copyOf)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return float64(copyOf[lower])
	}
	weight := position - float64(lower)
	return float64(copyOf[lower])*(1-weight) + float64(copyOf[upper])*weight
}

func MovingAverage(values []int, window int) []float64 {
	if window <= 0 {
		window = 1
	}
	result :=
		make([]float64, len(values))
	for position := range values {
		start := position - window + 1
		if start < 0 {
			start = 0
		}
		result[position] = Mean(values[start : position+1])
	}
	return result
}

func Differences(
	values []int,
) []int {
	if len(values) < 2 {
		return []int{}
	}
	result := make([]int, 0, len(values)-1)
	for index := 1; index < len(values); index++ {
		result = append(result, values[index]-values[index-1])
	}
	return result
}

func PositiveCount(
	values []int,
) int {
	return fold(values, 0, func(count, value int) int {
		if value > 0 {
			return count + 1
		}
		return count
	})
}

func NegativeCount(
	values []int,
) int {
	return fold(values, 0, func(count, value int) int {
		if value < 0 {
			return count + 1
		}
		return count
	})
}

func AbsoluteTotal(
	values []int,
) int {
	return fold(values, 0, func(total, value int) int {
		return total + int(math.Abs(float64(value)))
	})
}

func WeightedMean(values, weights []int) float64 {
	limit := len(values)
	if len(weights) < limit {
		limit = len(weights)
	}
	total, weightTotal := 0, 0
	for index := 0; index < limit; index++ {
		total += values[index] * weights[index]
		weightTotal += weights[index]
	}
	if weightTotal == 0 {
		return 0
	}
	return float64(total) / float64(weightTotal)
}

func Correlation(left, right []int) float64 {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	if limit < 2 {
		return 0
	}
	leftMean := Mean(left[:limit])
	rightMean := Mean(right[:limit])
	numerator, leftSum, rightSum := 0.0, 0.0, 0.0
	for index := 0; index < limit; index++ {
		leftDelta := float64(left[index]) - leftMean
		rightDelta := float64(right[index]) - rightMean
		numerator += leftDelta * rightDelta
		leftSum += leftDelta * leftDelta
		rightSum += rightDelta * rightDelta
	}
	denominator := math.Sqrt(leftSum * rightSum)
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func Trend(
	values []int,
) string {
	if len(values) < 2 {
		return "数据不足"
	}
	first := values[0]
	last := values[len(values)-1]
	if last > first {
		return "上升"
	}
	if last < first {
		return "下降"
	}
	return "平稳"
}

func Bucket(values []int, size int) map[int]int {
	result := map[int]int{}
	if size <= 0 {
		size = 1
	}
	for valueIndex := range values {
		value := values[valueIndex]
		key := value / size
		if value < 0 {
			key--
		}
		result[key*size]++
	}
	return result
}

func Normalize(
	values []int,
) []float64 {
	result :=
		make([]float64, len(values))
	if len(values) == 0 {
		return result
	}
	low, high := Min(values), Max(values)
	if low == high {
		for index := range result {
			result[index] = 0.5
		}
		return result
	}
	for index := range values {
		value := values[index]
		result[index] = float64(value-low) / float64(high-low)
	}
	return result
}

func RankDescending(
	values []int,
) []int {
	ranks :=
		make([]int, len(values))
	for index := range values {
		value := values[index]
		rank := 1
		for otherIndex := range values {
			other := values[otherIndex]
			if other > value {
				rank++
			}
		}
		ranks[index] = rank
	}
	return ranks
}
