package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
)

func (
	s *MatchService,
) Insights(
	owner, matchID domain.ID,
) (
	domain.MatchInsights,
	error,
) {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return insightError(err)
	}
	analysis, err :=
		s.Analyze(owner, matchID)
	if err != nil {
		return insightError(err)
	}
	report, err :=
		s.reportFor(owner, matchID)
	if err != nil {
		return insightError(err)
	}
	result := domain.MatchInsights{MatchID: matchID, GeneratedAt: time.Now().UTC(), Title: match.Name, Status: match.Status, Winner: report.Winner, HeadlineStats: insightStats(analysis, report), PlayerCards: playerInsights(analysis), ResourceCards: resourceInsights(analysis), Questions: append([]string(nil), report.Prompts...), Actions: analysis.Recommendations}
	result.Subtitle =
		insightSubtitle(
			match, analysis)
	result.LeadMessage =
		leadMessage(
			analysis, report)
	result.Story =
		insightStory(
			analysis, report)
	if len(result.PlayerCards) == 0 && len(seats) > 0 {
		result.PlayerCards =
			playerInsights(analysis)
	}
	return result, nil
}

func (
	s *MatchService,
) reportFor(
	owner, matchID domain.ID,
) (
	domain.MatchReport,
	error,
) {
	return NewReportService(s.store, s).Get(owner, matchID)
}

func insightSubtitle(
	match domain.Match,
	analysis matchAnalysis,
) string {
	if match.Status ==
		domain.MatchFinished {
		return fmt.Sprintf("完成了 %d 个回合的桌面记录，下面是这局故事的可读摘要。", analysis.Turns)
	}
	return fmt.Sprintf("这局还在%s，已记录 %d 个回合，摘要会随着行动继续更新。", match.Status, analysis.Turns)
}

func leadMessage(
	analysis matchAnalysis,
	report matchReport,
) string {
	if analysis.Events == 0 {
		return "目前还没有行动事件，先记录一条行动，系统才能开始组织这局故事。"
	}
	if analysis.Coverage < 0.5 {
		return fmt.Sprintf("%s暂时领先，但只有 %.0f%% 的回合留下记录，结论需要谨慎解读。", report.Winner, analysis.Coverage*100)
	}
	if analysis.ScoreSpread >= 8 {
		return fmt.Sprintf("%s以明显分差结束或暂时领先，建议回看第一次拉开差距的回合。", report.Winner)
	}
	return fmt.Sprintf("%s暂时占据第一，局势仍然具有复盘价值。", report.Winner)
}

func insightStats(
	analysis matchAnalysis,
	report matchReport,
) []domain.InsightStat {
	return []domain.InsightStat{
		{Label: "回合", Value: fmt.Sprint(analysis.Turns), Hint: "已推进的回合数"},
		{Label: "事件", Value: fmt.Sprint(analysis.Events), Hint: "行动、资源和旁白记录"},
		{Label: "覆盖率", Value: fmt.Sprintf("%.0f%%", analysis.Coverage*100), Hint: "有事件的回合占比"},
		{Label: "领先者", Value: report.Winner, Hint: "当前或最终得分最高的席位"},
		{Label: "关键时刻", Value: fmt.Sprint(len(report.TurningPoints)), Hint: "被标记的局势转折"},
	}
}

func playerInsights(
	analysis matchAnalysis,
) []domain.PlayerInsight {
	result :=
		make(
			[]domain.PlayerInsight, 0, len(analysis.Players))
	for _, player := range analysis.Players {
		oneLine := fmt.Sprintf("%s以%s完成了 %d 次行动，得到 %d 分。", player.SeatName, player.Style, player.ActionCount, player.Score)
		result = append(result, domain.PlayerInsight{SeatName: player.SeatName, Rank: player.Rank, Score: player.Score, Style: player.Style, OneLine: oneLine, Highlights: append([]string(nil), player.Strengths...), Watchouts: append([]string(nil), player.Risks...)})
	}
	sort.SliceStable(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].Rank
			right := result[j].Rank
			return left < right
		},
	)
	return result
}

func resourceInsights(
	analysis matchAnalysis,
) []domain.ResourceInsight {
	result :=
		make(
			[]domain.ResourceInsight, 0, len(analysis.ResourceTrails))
	for _, trail := range analysis.ResourceTrails {
		direction := "保持"
		if trail.NetChange > 0 {
			direction = "上升"
		} else if trail.NetChange < 0 {
			direction = "下降"
		}
		message := fmt.Sprintf("%s的%s从%d变化到%d，期间经历%d次正向、%d次负向调整。", trail.SeatName, trail.Resource, trail.Start, trail.End, trail.PositiveMoves, trail.NegativeMoves)
		result = append(result, domain.ResourceInsight{SeatName: trail.SeatName, Resource: trail.Resource, Direction: direction, Message: message, Peak: trail.Maximum, Low: trail.Minimum})
	}
	sort.Slice(
		result, func(i, j int) bool {
			if result[i].Resource == result[j].Resource {
				left := result[i].SeatName
				right := result[j].SeatName
				return left < right
			}
			left := result[i].Resource
			right := result[j].Resource
			return left < right
		})
	return result
}

func insightStory(
	analysis matchAnalysis,
	report matchReport,
) []domain.InsightChapter {
	chapters :=
		[]domain.InsightChapter{}
	chapters = append(chapters, domain.InsightChapter{Order: 1, Title: "开局", Summary: phaseSummary(analysis, domain.PeriodOpening), Evidence: phaseEvidence(analysis, domain.PeriodOpening), Tone: "neutral"})
	chapters = append(chapters, domain.InsightChapter{Order: 2, Title: "中局", Summary: phaseSummary(analysis, domain.PeriodMiddle), Evidence: phaseEvidence(analysis, domain.PeriodMiddle), Tone: "observe"})
	chapters = append(chapters, domain.InsightChapter{Order: 3, Title: "收官", Summary: closingSummary(analysis, report), Evidence: phaseEvidence(analysis, domain.PeriodClosing), Tone: closingTone(analysis)})
	for index := range chapters {
		chapters[index].TurnNumber = chapterTurn(analysis, chapters[index].Title)
	}
	return chapters
}

func phaseSummary(
	analysis matchAnalysis,
	phase analysisPeriod,
) string {
	for _, item := range analysis.TurnsByPhase {
		if item.Phase == phase {
			return item.Narrative
		}
	}
	return string(phase) + "还没有足够记录。"
}

func phaseEvidence(
	analysis matchAnalysis,
	phase analysisPeriod,
) []string {
	for _, item := range analysis.TurnsByPhase {
		if item.Phase == phase {
			return []string{fmt.Sprintf("回合 %d-%d", item.StartTurn, item.EndTurn), fmt.Sprintf("%d 个事件，影响级别合计 %d", item.EventCount, item.WeightTotal), "主导风格：" + emptyAs(item.DominantStyle, "尚未形成")}
		}
	}
	return []string{"没有可用的阶段数据"}
}

func closingSummary(
	analysis matchAnalysis,
	report matchReport,
) string {
	if report.Winner == "无人" {
		return "还没有形成可判断的收官结果。"
	}
	if analysis.ScoreSpread >= 8 {
		return fmt.Sprintf("%s在收官阶段建立了明显优势，值得回看分差第一次扩大时的选择。", report.Winner)
	}
	return fmt.Sprintf("%s在收官时保持领先，当前记录显示这是一局相对接近的对局。", report.Winner)
}

func closingTone(
	analysis matchAnalysis,
) string {
	if analysis.Pace.ClosingIntensity >= 0.5 {
		return "intense"
	}
	return "calm"
}

func chapterTurn(
	analysis matchAnalysis,
	title string,
) int {
	for _, phase := range analysis.TurnsByPhase {
		if string(phase.Phase) == title {
			return phase.StartTurn
		}
	}
	return 0
}

func emptyAs(
	value, fallback string,
) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
