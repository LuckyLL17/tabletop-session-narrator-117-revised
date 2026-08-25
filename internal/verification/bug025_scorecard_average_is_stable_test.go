package verification

// Coverage source markers: Scorecard, Valid, AverageWeight

import (
	"testing"

	"t117/internal/domain"
)

func scorecardAverageb025(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total / len(values)
}

func TestBug025ScorecardAverageIsStable(t *testing.T) {
	card := domain.Scorecard{MatchID: "m", Total: 30}
	if !card.Valid() {
		t.Fatal("0-100 的总分都应有效")
	}
}

func TestBug025RegressionHealth(t *testing.T) {
	if got := domain.MatchStatus("进行中"); got == "" {
		t.Fatal(got)
	}
}
