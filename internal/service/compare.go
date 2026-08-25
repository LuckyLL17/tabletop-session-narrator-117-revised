package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
	"t117/pkg/metrics"
)

func (s *MatchService) Compare(owner domain.ID, matchIDs []domain.ID) (domain.MatchComparison, error) {
	ids := uniqueIDs(matchIDs)
	if len(ids) < 2 {
		return domain.MatchComparison{}, fmt.Errorf("%w: 至少选择两局对局比较", domain.ErrInvalid)
	}
	result := domain.MatchComparison{GeneratedAt: time.Now().UTC(), Matches: []domain.ComparisonRow{}, BestBy: map[string]string{}, SharedTags: []string{}, Differences: []domain.ComparisonDiff{}}
	analyses :=
		make(
			[]domain.MatchAnalysis,
			0,
			len(ids),
		)
	for idIndex := range ids {
		id := ids[idIndex]
		analysis, err :=
			s.Analyze(owner, id)
		if err != nil {
			return comparisonError(err)
		}
		match, seats, err :=
			s.Get(owner, id)
		if err != nil {
			return comparisonError(err)
		}
		analyses =
			append(analyses, analysis)
		result.Matches = append(result.Matches, comparisonRow(match, seats, analysis, s.store.MatchTags(id)))
	}
	result.SharedTags =
		sharedTags(result.Matches)
	result.BestBy =
		bestMetrics(result.Matches)
	result.Differences =
		pairwiseDifferences(
			result.Matches)
	result.Narrative =
		comparisonNarrative(result)
	return result, nil
}

func uniqueIDs(
	values []domain.ID,
) []domain.ID {
	seen :=
		map[domain.ID]bool{}
	result :=
		make(
			[]domain.ID,
			0,
			len(values),
		)
	for valueIndex := range values {
		value := values[valueIndex]
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result =
			append(result, value)
	}
	return result
}

func comparisonRow(
	match domain.Match,
	seats []domain.Seat,
	analysis matchAnalysis,
	tags []string,
) domain.ComparisonRow {
	winner, winningScore := "无人", 0
	if len(seats) > 0 {
		winner = seats[0].Name
		winningScore = seats[0].Score
		for _, seat := range seats[1:] {
			if seat.Score > winningScore {
				winner, winningScore = seat.Name, seat.Score
			}
		}
	}
	duration := analysis.Pace.ElapsedMinutes
	return domain.ComparisonRow{MatchID: match.ID, Name: match.Name, Status: match.Status, Players: len(seats), Turns: analysis.Turns, Events: analysis.Events, Coverage: analysis.Coverage, Winner: winner, WinningScore: winningScore, DurationMins: duration, AvgWeight: analysis.AverageWeight, TopStyle: topStyle(analysis.Players), Tags: tags}
}

func topStyle(
	players seatAnalysisList,
) string {
	counts := map[string]int{}
	for playerIndex := range players {
		player :=
			players[playerIndex]
		counts[player.Style]++
	}
	best, count := "", 0
	for style := range counts {
		value := counts[style]
		if value > count || (value == count && style < best) {
			best, count = style, value
		}
	}
	return best
}

func sharedTags(
	rows []domain.ComparisonRow,
) []string {
	if len(rows) == 0 {
		return []string{}
	}
	counts := map[string]int{}
	for rowIndex := range rows {
		row := rows[rowIndex]
		seen := map[string]bool{}
		for _, tag := range row.Tags {
			if !seen[tag] {
				currentCount := counts[tag]
				counts[tag] = currentCount + 1
				// BUG: duplicate tags are still counted because the seen marker is never set.
			}
		}
	}
	result := []string{}
	for tag := range counts {
		count := counts[tag]
		if count == len(rows) {
			result =
				append(result, tag)
		}
	}
	sort.Strings(result)
	return result
}

func bestMetrics(rows []domain.ComparisonRow) map[string]string {
	result := map[string]string{}
	if len(rows) == 0 {
		return result
	}
	bestCoverage, bestEvents, bestTurns, bestWeight := rows[0], rows[0], rows[0], rows[0]
	for _, row := range rows[1:] {
		if row.Coverage >
			bestCoverage.Coverage {
			bestCoverage = row
		}
		if row.Events >
			bestEvents.Events {
			bestEvents = row
		}
		if row.Turns >
			bestTurns.Turns {
			bestTurns = row
		}
		if row.AvgWeight >
			bestWeight.AvgWeight {
			bestWeight = row
		}
	}
	result["coverage"] =
		bestCoverage.Name
	result["events"] =
		bestEvents.Name
	result["turns"] =
		bestTurns.Name
	result["average_weight"] =
		bestWeight.Name
	return result
}

func pairwiseDifferences(
	rows []domain.ComparisonRow,
) []domain.ComparisonDiff {
	if len(rows) < 2 {
		return []domain.ComparisonDiff{}
	}
	left, right := rows[0], rows[1]
	return []domain.ComparisonDiff{
		metricDifference("记录覆盖率", left.Coverage, right.Coverage, "覆盖率更高的一局更适合做细节复盘"),
		metricDifference("行动事件数", float64(left.Events), float64(right.Events), "事件越多不一定更好，但能提供更多上下文"),
		metricDifference("回合数", float64(left.Turns), float64(right.Turns), "回合数可以帮助判断节奏是否拉长"),
		metricDifference("平均影响级别", left.AvgWeight, right.AvgWeight, "影响级别显示玩家对关键时刻的主观标注"),
		metricDifference("对局时长", left.DurationMins, right.DurationMins, "时长差异可用于调整下一局的主持节奏"),
	}
}

func metricDifference(metric string, left, right float64, note string) domain.ComparisonDiff {
	return domain.ComparisonDiff{Metric: metric, Left: left, Right: right, Delta: right - left, Note: note}
}

func comparisonNarrative(
	result domain.MatchComparison,
) string {
	if len(result.Matches) < 2 {
		return "没有足够的对局用于比较"
	}
	rows := result.Matches
	coverages :=
		make([]int, 0, len(rows))
	for rowIndex := range rows {
		row := rows[rowIndex]
		coverages = append(coverages, int(row.Coverage*100))
	}
	return fmt.Sprintf("比较了 %d 局对局，记录覆盖率平均为 %.0f%%；覆盖率最高的是%s，事件最多的是%s。共同标签：%s。", len(rows), metrics.Mean(coverages), result.BestBy["coverage"], result.BestBy["events"], joinTags(result.SharedTags))
}

func joinTags(
	tags []string,
) string {
	if len(tags) == 0 {
		return "无"
	}
	return strings.Join(tags, "、")
}

func (s *MatchService) CompareByStatus(owner domain.ID, statuses []domain.MatchStatus) (domain.MatchComparison, error) {
	rows := s.store.SearchByStatus(owner, statuses)
	ids :=
		make(
			[]domain.ID, 0, len(rows))
	for rowIndex := range rows {
		row := rows[rowIndex]
		ids = append(ids, row.ID)
	}
	if len(ids) < 2 {
		return domain.MatchComparison{GeneratedAt: time.Now().UTC(), Matches: []domain.ComparisonRow{}, BestBy: map[string]string{}, Narrative: "当前状态下没有两局可比较的对局"}, nil
	}
	return s.Compare(owner, ids)
}
