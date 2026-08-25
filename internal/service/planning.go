package service

import (
	"fmt"
	"sort"
	"strings"

	"t117/internal/domain"
)

func (
	s *MatchService,
) NextSteps(
	owner, matchID domain.ID,
) (
	[]domain.Recommendation, error) {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return nil, err
	}
	analysis, err :=
		s.Analyze(owner, matchID)
	if err != nil {
		return nil, err
	}
	result :=
		[]domain.Recommendation{}
	if match.Status ==
		domain.MatchSetup {
		result = append(result, domain.Recommendation{Priority: 1, Audience: "主持人", Title: "开始这局对局", Reason: "当前仍处于准备阶段，开始后才能推进回合和记录行动。", Steps: []string{"确认玩家席位", "检查规则变体", "点击开始对局"}, RelatedIDs: []domain.ID{matchID}})
		return result, nil
	}
	if match.Status ==
		domain.MatchPaused {
		result = append(result, domain.Recommendation{Priority: 1, Audience: "主持人", Title: "恢复或结束暂停", Reason: pauseReason(match), Steps: []string{"确认暂停原因已经解决", "恢复对局继续记录", "若不再继续则结束对局"}, RelatedIDs: []domain.ID{matchID}})
	}
	if match.Status ==
		domain.MatchLive {
		result = append(result, liveTurnSteps(match, seats, analysis)...)
	}
	if match.Status ==
		domain.MatchFinished {
		result = append(result, finishedSteps(analysis, matchID)...)
	}
	result = append(result, dataQualitySteps(analysis, matchID)...)
	sort.SliceStable(
		result, func(i, j int) bool {
			if result[i].Priority == result[j].Priority {
				left := result[i].Title
				right := result[j].Title
				return left < right
			}
			left := result[i].Priority
			right := result[j].Priority
			return left < right
		})
	return uniqueRecommendations(result), nil
}

func pauseReason(
	match domain.Match,
) string {
	if strings.TrimSpace(
		match.PauseReason) == "" {
		return "对局已暂停，但没有留下原因。"
	}
	return "暂停原因：" + match.PauseReason
}

func liveTurnSteps(match domain.Match, seats []domain.Seat, analysis domain.MatchAnalysis) []domain.Recommendation {
	result :=
		[]domain.Recommendation{}
	if len(seats) == 0 {
		return []domain.Recommendation{{Priority: 1, Audience: "主持人", Title: "补充玩家席位", Reason: "没有席位时无法推进回合。", Steps: []string{"返回桌游档案检查人数", "为每个玩家分配席位"}, RelatedIDs: []domain.ID{match.ID}}}
	}
	active := seats[match.CurrentSeat%len(seats)]
	result = append(result, domain.Recommendation{Priority: 1, Audience: "当前玩家", Title: "记录" + active.Name + "的第" + fmt.Sprint(match.TurnNumber+1) + "回合", Reason: "按席位顺序推进能让时间线保持完整。", Steps: []string{"写下本回合关注点", "记录行动和资源变化", "回合结束后切换下一位玩家"}, RelatedIDs: []domain.ID{active.ID, match.ID}})
	if analysis.Coverage < 0.75 {
		result = append(result, domain.Recommendation{Priority: 2, Audience: "记录者", Title: "优先补齐空白回合", Reason: fmt.Sprintf("当前覆盖率只有 %.0f%%", analysis.Coverage*100), Steps: []string{"回看上一回合", "补写一句发生了什么", "标记是否存在关键时刻"}, RelatedIDs: []domain.ID{match.ID}})
	}
	if len(analysis.ResourceTrails) == 0 {
		result = append(result, domain.Recommendation{Priority: 3, Audience: "记录者", Title: "启用资源变化记录", Reason: "当前还没有资源轨迹，后续难以解释策略选择。", Steps: []string{"在行动事件中填写资源增减", "保持资源名称前后一致", "在大幅变化时补充说明"}, RelatedIDs: []domain.ID{match.ID}})
	}
	return result
}

func finishedSteps(
	analysis matchAnalysis,
	matchID domain.ID,
) []domain.Recommendation {
	result := []domain.Recommendation{{Priority: 1, Audience: "复盘者", Title: "查看战报和时间线", Reason: fmt.Sprintf("本局包含 %d 个回合和 %d 个行动事件。", analysis.Turns, analysis.Events), Steps: []string{"先看玩家线", "再看关键时刻", "最后回答复盘问题"}, RelatedIDs: []domain.ID{matchID}}}
	if len(analysis.Signals) > 0 {
		result = append(result, domain.Recommendation{Priority: 2, Audience: "复盘者", Title: "处理分析提示", Reason: fmt.Sprintf("系统识别出 %d 个可讨论信号。", len(analysis.Signals)), Steps: signalSteps(analysis.Signals), RelatedIDs: []domain.ID{matchID}})
	}
	if analysis.Coverage >= 0.75 {
		result = append(result, domain.Recommendation{Priority: 3, Audience: "主持人", Title: "导出本局记录", Reason: "当前记录覆盖率足够，适合保存为战报。", Steps: []string{"导出 Markdown 便于分享", "导出 JSON 便于继续分析", "为下一局复制可用标签"}, RelatedIDs: []domain.ID{matchID}})
	}
	return result
}

func dataQualitySteps(
	analysis matchAnalysis,
	matchID domain.ID,
) []domain.Recommendation {
	result :=
		[]domain.Recommendation{}
	if analysis.Events == 0 && analysis.Turns > 0 {
		result = append(result, domain.Recommendation{Priority: 2, Audience: "记录者", Title: "确认事件是否漏记", Reason: "有回合但没有行动事件，可能是记录中断。", Steps: []string{"检查纸面或口头记录", "补录关键行动", "为无法还原的部分写旁白"}, RelatedIDs: []domain.ID{matchID}})
	}
	for _, player := range analysis.Players {
		if player.ActionCount == 0 {
			result = append(result, domain.Recommendation{Priority: 3, Audience: player.SeatName, Title: "补充个人行动线", Reason: "这个席位还没有任何行动事件。", Steps: []string{"确认是否参与了本局", "补写至少一个关键行动", "说明没有记录的原因"}, RelatedIDs: []domain.ID{player.SeatID}})
		}
	}
	return result
}

func signalSteps(
	signals strategySignalList,
) []string {
	result := []string{}
	for signalIndex := range signals {
		signal :=
			signals[signalIndex]
		if signal.SuggestedAsk != "" {
			result =
				append(
					result, signal.SuggestedAsk)
		} else {
			result =
				append(
					result, signal.Title)
		}
		if len(result) == 4 {
			break
		}
	}
	return result
}

func uniqueRecommendations(
	values recommendationList,
) []domain.Recommendation {
	seen := map[string]bool{}
	result :=
		make(
			[]domain.Recommendation,
			0,
			len(values),
		)
	for valueIndex := range values {
		value := values[valueIndex]
		key := value.Audience + "|" + value.Title
		if seen[key] {
			continue
		}
		seen[key] = true
		result =
			append(result, value)
	}
	return result
}
