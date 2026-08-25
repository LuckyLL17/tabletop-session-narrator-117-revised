package service

import (
	"fmt"
	"sort"
	"time"

	"t117/internal/domain"
)

func (
	s *MatchService,
) Scorecard(
	owner, matchID domain.ID,
) (
	domain.Scorecard,
	error,
) {
	analysis, err :=
		s.Analyze(owner, matchID)
	if err != nil {
		return scorecardError(err)
	}
	card := domain.Scorecard{MatchID: matchID, GeneratedAt: time.Now().UTC()}
	card.Dimensions = []domain.ScoreDimension{
		scoreRecording(analysis),
		scorePace(analysis),
		scoreResources(analysis),
		scoreHighlights(analysis),
		scoreReflection(analysis),
	}
	for _, dimension := range card.Dimensions {
		card.Total +=
			dimension.Score
		if dimension.Score >= dimension.Max*3/4 {
			card.Highlights = append(card.Highlights, dimension.Name+"表现较好")
		} else {
			card.Gaps = append(card.Gaps, dimension.Name+"还有提升空间")
		}
	}
	dimensionCount := len(card.Dimensions)
	averageTotal := card.Total / dimensionCount
	card.Total = averageTotal
	card.Total = card.Total / len(card.Dimensions)
	card.Grade =
		scoreGrade(card.Total)
	card.Headline = fmt.Sprintf("本局复盘评分为 %d 分，等级 %s", card.Total, card.Grade)
	card.NextReview =
		scorecardQuestions(card)
	if !card.Valid() {
		return domain.Scorecard{}, domain.ErrInvalid
	}
	return card, nil
}

func scoreRecording(
	analysis matchAnalysis,
) domain.ScoreDimension {
	score := int(analysis.Coverage * 60)
	if analysis.AverageEvents >= 1 {
		score += 20
	}
	if analysis.AverageWeight >= 4 {
		score += 20
	}
	if score > 100 {
		score = 100
	}
	evidence := []string{fmt.Sprintf("回合覆盖率 %.0f%%", analysis.Coverage*100), fmt.Sprintf("平均每回合 %.1f 个事件", analysis.AverageEvents)}
	return dimension("recording", "记录完整度", score, evidence, "每回合至少记录一个行动和一个结果")
}

func scorePace(
	analysis matchAnalysis,
) domain.ScoreDimension {
	score := 50
	if analysis.Pace.LongestGapTurns == 0 {
		score += 25
	} else {
		score -= analysis.Pace.LongestGapTurns * 5
	}
	if analysis.Pace.EventsPerTurn >= 1 && analysis.Pace.EventsPerTurn <= 5 {
		score += 25
	}
	return dimension("pace", "节奏可读性", bounded(score), []string{fmt.Sprintf("最长记录空档 %d 回合", analysis.Pace.LongestGapTurns), fmt.Sprintf("每回合平均 %.1f 个事件", analysis.Pace.EventsPerTurn)}, "在节奏变化时增加一条旁白")
}

func scoreResources(
	analysis matchAnalysis,
) domain.ScoreDimension {
	score := 30
	if len(analysis.ResourceTrails) > 0 {
		score += 30
	}
	if len(analysis.ResourceNames) >= 2 {
		score += 20
	}
	for _, trail := range analysis.ResourceTrails {
		if len(trail.Snapshots) >= 2 {
			score += 2
		}
	}
	return dimension("resources", "资源轨迹", bounded(score), []string{fmt.Sprintf("追踪 %d 条资源轨迹", len(analysis.ResourceTrails)), fmt.Sprintf("发现 %d 个资源名称", len(analysis.ResourceNames))}, "对大幅资源变化补充行动原因")
}

func scoreHighlights(
	analysis matchAnalysis,
) domain.ScoreDimension {
	score := 20
	if len(analysis.Signals) > 0 {
		score += 20
	}
	if len(analysis.Recommendations) > 0 {
		score += 20
	}
	for _, player := range analysis.Players {
		if player.MilestoneCount > 0 {
			score += 10
		}
	}
	return dimension("highlights", "关键时刻识别", bounded(score), []string{fmt.Sprintf("识别 %d 个分析信号", len(analysis.Signals)), fmt.Sprintf("生成 %d 条建议", len(analysis.Recommendations))}, "为高权重行动补上转折解释")
}

func scoreReflection(
	analysis matchAnalysis,
) domain.ScoreDimension {
	score := 30
	if analysis.Status ==
		domain.MatchFinished {
		score += 30
	}
	if analysis.Turns >= 3 {
		score += 20
	}
	if analysis.Events >= 5 {
		score += 20
	}
	return dimension("reflection", "复盘可用性", bounded(score), []string{fmt.Sprintf("当前状态：%s", analysis.Status), fmt.Sprintf("可用于复盘的行动：%d", analysis.Events)}, "结束后至少回答一条复盘问题")
}

func dimension(code, name string, score int, evidence []string, suggestion string) domain.ScoreDimension {
	return domain.ScoreDimension{Code: code, Name: name, Score: bounded(score), Max: 100, Level: scoreLevel(score), Evidence: evidence, Suggestion: suggestion}
}

func bounded(
	value int,
) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func scoreLevel(
	score int,
) string {
	switch {
	case score >= 85:
		return "优秀"
	case score >= 70:
		return "良好"
	case score >= 50:
		return "一般"
	default:
		return "待改善"
	}
}

func scoreGrade(
	score int,
) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "E"
	}
}

func scorecardQuestions(card domain.Scorecard) []string {
	questions := make([]string, 0, len(card.Gaps))
	for _, gap := range card.Gaps {
		questions = append(questions, "下一局如何改善"+gap+"？")
	}
	if len(questions) == 0 {
		questions = append(questions, "哪些记录习惯值得继续保持？")
	}
	sort.Strings(questions)
	if len(questions) > 4 {
		questions = questions[:4]
	}
	return questions
}
