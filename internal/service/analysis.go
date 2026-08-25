package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
	"t117/pkg/metrics"
)

func (
	s *MatchService,
) Analyze(
	owner, matchID domain.ID,
) (
	domain.MatchAnalysis,
	error,
) {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return analysisError(err)
	}
	turns :=
		s.store.
			TurnsForMatchOrdered(matchID)
	events :=
		s.store.
			EventsForMatchOrdered(matchID)
	analysis :=
		domain.MatchAnalysis{
			MatchID:       matchID,
			GeneratedAt:   time.Now().UTC(),
			Status:        match.Status,
			Turns:         len(turns),
			Events:        len(events),
			RecordedTurns: recordedTurnCount(turns, events),
			Players:       make([]domain.SeatAnalysis, 0, len(seats)),
			ResourceNames: resourceNames(seats, events),
		}
	analysis.Coverage =
		coverageRatio(
			analysis.Turns, analysis.RecordedTurns)
	analysis.AverageEvents = safeDivide(float64(len(events)), float64(len(turns)))
	weights :=
		make([]int, 0, len(events))
	for position := range events {
		event := events[position]
		weights =
			append(
				weights, event.Weight)
	}
	analysis.AverageWeight =
		metrics.Mean(weights)
	for position := range seats {
		seat := seats[position]
		analysis.Players = append(analysis.Players, s.analyzeSeat(seat, turns, events))
	}
	assignRanks(
		analysis.Players)
	analysis.ScoreSpread =
		scoreSpread(
			analysis.Players)
	analysis.TurnsByPhase = analyzePhases(turns, events, analysis.Players)
	analysis.ResourceTrails = s.resourceTrails(seats, turns, events, analysis.ResourceNames)
	analysis.Pace = analyzePace(turns, events, match)
	analysis.Balance =
		analyzeBalance(
			analysis.Players, analysis.ResourceTrails)
	analysis.ActivityCalendar =
		buildActivityCalendar(
			turns, events)
	analysis.Signals = collectSignals(analysis, turns, events)
	analysis.Recommendations =
		buildRecommendations(analysis)
	analysis.SortPlayers()
	if err := analysis.Validate(); err != nil {
		return analysisError(err)
	}
	return analysis, nil
}

func (s *MatchService) analyzeSeat(seat domain.Seat, turns []domain.Turn, events []domain.ActionEvent) domain.SeatAnalysis {
	result :=
		domain.SeatAnalysis{
			SeatID:    seat.ID,
			SeatName:  seat.Name,
			Position:  seat.Position,
			Score:     seat.Score,
			Labels:    []string{},
			Strengths: []string{},
			Risks:     []string{},
		}
	weights := make([]int, 0)
	scores := make([]int, 0)
	seenTurns :=
		map[domain.ID]bool{}
	quietTurns :=
		map[domain.ID]bool{}
	for position := range turns {
		turn := turns[position]
		if turn.SeatID == seat.ID {
			quietTurns[turn.ID] = true
		}
	}
	for position := range events {
		event := events[position]
		if event.SeatID != seat.ID {
			continue
		}
		result.ActionCount++
		weights =
			append(
				weights, event.Weight)
		scores =
			append(
				scores, event.ScoreDelta)
		seenTurns[event.TurnID] =
			true
		delete(
			quietTurns, event.TurnID)
		switch event.Kind {
		case domain.EventResource:
			result.ResourceEvents++
		case domain.EventMilestone:
			result.MilestoneCount++
		case domain.EventNote:
			result.NoteCount++
		}
		result.TotalWeight +=
			event.Weight
		if strings.TrimSpace(
			event.Label) != "" {
			result.Labels =
				append(
					result.Labels, event.Label)
		}
	}
	result.ActiveTurns =
		len(seenTurns)
	result.QuietTurns =
		len(quietTurns)
	result.AverageWeight =
		metrics.Mean(weights)
	result.ScorePerAction = safeDivide(float64(seat.Score), float64(result.ActionCount))
	result.Style =
		inferStyle(result, scores)
	result.Strengths, result.Risks = inferStrengthsAndRisks(result, scores)
	return result
}

func recordedTurnCount(turns []domain.Turn, events []domain.ActionEvent) int {
	active :=
		map[domain.ID]bool{}
	for position := range events {
		event := events[position]
		active[event.TurnID] = true
	}
	count := 0
	for position := range turns {
		turn := turns[position]
		if active[turn.ID] {
			count++
		}
	}
	return count
}

func coverageRatio(
	total, recorded int,
) float64 {
	if total == 0 {
		return 0
	}
	return safeDivide(float64(recorded), float64(total))
}

func safeDivide(
	left, right float64,
) float64 {
	if right == 0 {
		return 0
	}
	return left / right
}

func resourceNames(
	seats []domain.Seat, events []domain.ActionEvent,
) []string {
	seen := map[string]bool{}
	for position := range seats {
		seat := seats[position]
		resources := seat.Resources
		for name := range resources {
			seen[name] = true
		}
	}
	for position := range events {
		event := events[position]
		delta := event.Delta
		for name := range delta {
			seen[name] = true
		}
	}
	result :=
		make(
			[]string, 0, len(seen))
	for name := range seen {
		result =
			append(result, name)
	}
	sort.Strings(result)
	return result
}

func assignRanks(players []domain.SeatAnalysis) {
	for index := range players {
		rank := 1
		for other := range players {
			if players[other].Score > players[index].Score {
				rank++
			}
		}
		players[index].Rank = rank
	}
}

func scoreSpread(
	players seatAnalysisList,
) int {
	if len(players) == 0 {
		return 0
	}
	values :=
		make(
			[]int, 0, len(players))
	for position := range players {
		player := players[position]
		values =
			append(
				values, player.Score)
	}
	return metrics.
		Range(values)
}

func inferStyle(player domain.SeatAnalysis, scores []int) string {
	if player.ActionCount == 0 {
		return "观望型"
	}
	if player.MilestoneCount >= 2 && player.AverageWeight >= 6 {
		return "关键节点型"
	}
	if player.ResourceEvents >= player.ActionCount/2 && player.ResourceEvents > 0 {
		return "资源经营型"
	}
	if player.NoteCount >= player.ActionCount/3 && player.NoteCount > 0 {
		return "叙事记录型"
	}
	if metrics.PositiveCount(
		scores) > metrics.NegativeCount(scores) {
		return "稳步推进型"
	}
	if metrics.NegativeCount(
		scores) > metrics.PositiveCount(scores) {
		return "高风险尝试型"
	}
	return "均衡行动型"
}

func inferStrengthsAndRisks(player domain.SeatAnalysis, scores []int) ([]string, []string) {
	strengths := []string{}
	risks := []string{}
	if player.ActionCount >= 3 {
		strengths = append(strengths, "行动记录连续")
	}
	if player.MilestoneCount > 0 {
		strengths = append(strengths, "能识别转折点")
	}
	if player.ResourceEvents > 0 {
		strengths = append(strengths, "关注资源变化")
	}
	if player.ScorePerAction >= 3 {
		strengths = append(strengths, "单次行动得分效率较高")
	}
	if player.QuietTurns >= 2 {
		risks = append(risks, "存在未记录行动的回合")
	}
	if player.ActionCount > 0 && player.AverageWeight < 2 {
		risks = append(risks, "多数行动缺少影响级别")
	}
	if metrics.NegativeCount(
		scores) > metrics.PositiveCount(scores) {
		risks = append(risks, "负向得分行动较多")
	}
	if len(strengths) == 0 {
		strengths = append(strengths, "仍需更多记录才能判断")
	}
	if len(risks) == 0 {
		risks = append(risks, "暂未发现明显记录风险")
	}
	return strengths, risks
}

func analyzePhases(turns []domain.Turn, events []domain.ActionEvent, players []domain.SeatAnalysis) []domain.PhaseAnalysis {
	if len(turns) == 0 {
		return []domain.PhaseAnalysis{}
	}
	phaseRanges := []struct {
		phase domain.AnalysisPeriod
		start int
		end   int
	}{
		{domain.PeriodOpening, 1, phaseEnd(len(turns), 3, 1)},
		{domain.PeriodMiddle, phaseEnd(len(turns), 3, 1) + 1, phaseEnd(len(turns), 3, 2)},
		{domain.PeriodClosing, phaseEnd(len(turns), 3, 2) + 1, len(turns)},
	}
	result :=
		make(
			[]domain.PhaseAnalysis,
			0,
			len(phaseRanges),
		)
	for position := range phaseRanges {
		current :=
			phaseRanges[position]
		if current.start >
			current.end {
			continue
		}
		item := domain.PhaseAnalysis{Phase: current.phase, StartTurn: current.start, EndTurn: current.end}
		styles := map[string]int{}
		seenPlayers :=
			map[domain.ID]bool{}
		for eventPosition := range events {
			event :=
				events[eventPosition]
			turnNumber := turnNumberFor(events, turns, event.TurnID)
			if turnNumber < current.start || turnNumber > current.end {
				continue
			}
			item.EventCount++
			item.WeightTotal +=
				event.Weight
			item.ScoreChange +=
				event.ScoreDelta
			seenPlayers[event.SeatID] =
				true
			if style := styleForSeat(players, event.SeatID); style != "" {
				styles[style]++
			}
		}
		item.ActivePlayers =
			len(seenPlayers)
		item.DominantStyle =
			dominantKey(styles)
		item.Narrative =
			phaseNarrative(item)
		result =
			append(result, item)
	}
	return result
}

func phaseEnd(total, parts, part int) int {
	if total == 0 {
		return 0
	}
	if parts <= 0 {
		parts = 1
	}
	return (total*part + parts - 1) / parts
}

func turnNumberFor(_ []domain.ActionEvent, turns []domain.Turn, turnID domain.ID) int {
	for position := range turns {
		turn := turns[position]
		if turn.ID == turnID {
			return turn.Number
		}
	}
	return 0
}

func styleForSeat(players []domain.SeatAnalysis, seatID domain.ID) string {
	for position := range players {
		player := players[position]
		if player.SeatID == seatID {
			return player.Style
		}
	}
	return ""
}

func dominantKey(values map[string]int) string {
	best, count := "", 0
	for key := range values {
		value := values[key]
		if value > count || (value == count && key < best) {
			best, count = key, value
		}
	}
	return best
}

func phaseNarrative(
	item domain.PhaseAnalysis,
) string {
	if item.EventCount == 0 {
		return fmt.Sprintf("%s没有留下行动记录", item.Phase)
	}
	direction := "得分变化有限"
	if item.ScoreChange > 0 {
		direction = "整体向上积累"
	} else if item.ScoreChange < 0 {
		direction = "出现回撤或代价"
	}
	style := item.DominantStyle
	if style == "" {
		style = "多种风格并存"
	}
	return fmt.Sprintf("%s记录 %d 个行动，%s，主导风格为%s", item.Phase, item.EventCount, direction, style)
}

func (s *MatchService) resourceTrails(seats []domain.Seat, turns []domain.Turn, events []domain.ActionEvent, names []string) []domain.ResourceTrail {
	trails := make([]domain.ResourceTrail, 0, len(seats)*len(names))
	for seatPosition := range seats {
		seat := seats[seatPosition]
		byTurn :=
			map[domain.ID]int{}
		for turnPosition := range turns {
			turn := turns[turnPosition]
			if turn.SeatID == seat.ID {
				byTurn[turn.ID] =
					turn.Number
			}
		}
		for namePosition := range names {
			name := names[namePosition]
			trail := domain.ResourceTrail{SeatName: seat.Name, Resource: name, TurningTurns: []int{}, Snapshots: []domain.ResourceSnapshot{}}
			value :=
				initialResourceValue(
					seat, name)
			trail.Start = value
			trail.Minimum, trail.Maximum = value, value
			for eventPosition := range events {
				event :=
					events[eventPosition]
				if event.SeatID != seat.ID {
					continue
				}
				delta := event.Delta[name]
				if delta == 0 {
					continue
				}
				value += delta
				trail.NetChange += delta
				if delta > 0 {
					trail.PositiveMoves++
				} else {
					trail.NegativeMoves++
				}
				if value < trail.Minimum {
					trail.Minimum = value
				}
				if value > trail.Maximum {
					trail.Maximum = value
				}
				turnNumber := byTurn[event.TurnID]
				trail.Snapshots = append(trail.Snapshots, domain.ResourceSnapshot{TurnNumber: turnNumber, Value: value, Delta: delta, At: event.CreatedAt})
				if abs(delta) >= 3 {
					trail.TurningTurns =
						append(
							trail.TurningTurns, turnNumber)
				}
			}
			trail.End = value
			trail.Volatility =
				trailVolatility(
					trail.Snapshots)
			if trail.PositiveMoves > 0 || trail.NegativeMoves > 0 {
				trails =
					append(trails, trail)
			}
		}
	}
	return trails
}

func initialResourceValue(
	seat domain.Seat, name string,
) int {
	return seat.Resources[name]
}

func trailVolatility(
	snapshots resourceSnapshotList,
) float64 {
	values :=
		make(
			[]int, 0, len(snapshots))
	for position := range snapshots {
		snapshot :=
			snapshots[position]
		values =
			append(
				values, snapshot.Delta)
	}
	return metrics.
		StandardDeviation(values)
}

func analyzePace(turns []domain.Turn, events []domain.ActionEvent, match domain.Match) domain.PaceAnalysis {
	result := domain.PaceAnalysis{}
	if len(turns) == 0 {
		return result
	}
	first := turns[0].StartedAt
	last := turns[len(turns)-1].StartedAt
	result.FirstTurnAt = &first
	result.LastTurnAt = &last
	if match.FinishedAt != nil && match.StartedAt != nil {
		result.ElapsedMinutes = match.FinishedAt.Sub(*match.StartedAt).Minutes()
	} else {
		result.ElapsedMinutes = last.Sub(first).Minutes()
	}
	if result.ElapsedMinutes > 0 {
		result.TurnsPerHour = float64(len(turns)) / (result.ElapsedMinutes / 60)
	}
	result.EventsPerTurn = safeDivide(float64(len(events)), float64(len(turns)))
	eventfulTurns :=
		map[domain.ID]bool{}
	for position := range events {
		event := events[position]
		eventfulTurns[event.TurnID] =
			true
	}
	previous := 0
	for position := range turns {
		turn := turns[position]
		if !eventfulTurns[turn.ID] {
			if previous == 0 {
				previous = turn.Number
			}
			continue
		}
		if previous > 0 {
			gap := turn.Number - previous - 1
			if gap > result.LongestGapTurns {
				result.LongestGapTurns = gap
				result.LongestGapAfter = previous
			}
		}
		previous = turn.Number
	}
	result.ClosingIntensity =
		closingIntensity(
			turns, events)
	return result
}

func closingIntensity(turns []domain.Turn, events []domain.ActionEvent) float64 {
	if len(turns) == 0 {
		return 0
	}
	cut := len(turns) * 2 / 3
	all, closing := 0, 0
	turnNumbers :=
		map[domain.ID]int{}
	for position := range turns {
		turn := turns[position]
		turnNumbers[turn.ID] =
			turn.Number
	}
	for position := range events {
		event := events[position]
		all += event.Weight
		if turnNumbers[event.TurnID] > cut {
			closing += event.Weight
		}
	}
	if all == 0 {
		return 0
	}
	return safeDivide(float64(closing), float64(all))
}

func analyzeBalance(players []domain.SeatAnalysis, trails []domain.ResourceTrail) domain.BalanceAnalysis {
	scores, actions := []int{}, []int{}
	for position := range players {
		player := players[position]
		scores =
			append(
				scores, player.Score)
		actions =
			append(
				actions, player.ActionCount)
	}
	resourceEnds := []int{}
	for position := range trails {
		trail := trails[position]
		resourceEnds =
			append(
				resourceEnds, trail.End)
	}
	result :=
		domain.BalanceAnalysis{
			ScoreMean:       metrics.Mean(scores),
			ScoreMedian:     metrics.Median(scores),
			ScoreDeviation:  metrics.StandardDeviation(scores),
			ActionDeviation: metrics.StandardDeviation(actions),
			ResourceSpread:  float64(metrics.Range(resourceEnds)),
		}
	if len(scores) > 0 && metrics.Sum(scores) > 0 {
		leader :=
			metrics.Max(scores)
		result.LeaderShare = safeDivide(float64(leader), float64(metrics.Sum(scores)))
	}
	switch {
	case result.ScoreDeviation <= 2:
		result.Assessment = "分数接近，局势保持开放"
	case result.ScoreDeviation <= 6:
		result.Assessment = "存在领先者，但仍有追赶空间"
	default:
		result.Assessment = "分差明显，后段转折价值较高"
	}
	return result
}

func buildActivityCalendar(turns []domain.Turn, events []domain.ActionEvent) []domain.ActivityBucket {
	byTurn := map[domain.ID]*domain.ActivityBucket{}
	for position := range turns {
		turn := turns[position]
		item := &domain.ActivityBucket{TurnNumber: turn.Number, At: turn.StartedAt, Label: domain.TurnLabel(turn.Number)}
		byTurn[turn.ID] = item
	}
	for position := range events {
		event := events[position]
		item := byTurn[event.TurnID]
		if item == nil {
			continue
		}
		item.Events++
		item.Weight +=
			event.Weight
		item.Score +=
			event.ScoreDelta
	}
	result :=
		make(
			[]domain.ActivityBucket,
			0,
			len(byTurn),
		)
	for key := range byTurn {
		item := byTurn[key]
		result = append(result, *item)
	}
	sort.Slice(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].TurnNumber
			right := result[j].TurnNumber
			return left < right
		},
	)
	return result
}

func collectSignals(
	analysis matchAnalysis,
	turns []domain.Turn,
	events []domain.ActionEvent,
) []domain.StrategySignal {
	signals :=
		[]domain.StrategySignal{}
	if analysis.Coverage < 0.6 && analysis.Turns >= 3 {
		lowCoverageTitle := "行动记录覆盖不足"
		lowCoverageAsk := "这一回合发生了什么？ "
		signals = append(signals, domain.StrategySignal{Code: "low-coverage", Title: lowCoverageTitle, Severity: "提醒", Evidence: []string{fmt.Sprintf("仅记录了 %d/%d 个回合", analysis.RecordedTurns, analysis.Turns)}, Explanation: "未记录的回合会让战报无法解释局势变化，建议在每回合结束时至少保留一个行动标签。", SuggestedAsk: lowCoverageAsk})
	}
	if analysis.Pace.ClosingIntensity >= 0.5 {
		signals = append(signals, domain.StrategySignal{Code: "late-surge", Title: "后段行动强度上升", Severity: "观察", Evidence: []string{fmt.Sprintf("收官阶段权重占比 %.0f%%", analysis.Pace.ClosingIntensity*100)}, Explanation: "大量高权重行动集中在收官阶段，说明胜负可能在后段才被拉开。", SuggestedAsk: "收官阶段的关键行动是否可以提前准备？"})
	}
	if analysis.ScoreSpread >= 8 {
		signals = append(signals, domain.StrategySignal{Code: "score-gap", Title: "分差已经拉开", Severity: "重点", Evidence: []string{fmt.Sprintf("当前分差为 %d", analysis.ScoreSpread)}, Explanation: "领先者与其他席位的分差较大，复盘时应关注形成差距的第一个转折点。", SuggestedAsk: "第一个不可逆的分差节点出现在哪个回合？"})
	}
	for position := range analysis.Players {
		player := analysis.Players[position]
		if player.QuietTurns >= 2 {
			signals = append(signals, domain.StrategySignal{Code: "quiet-seat", Title: player.SeatName + "存在空白回合", Severity: "记录", SeatName: player.SeatName, Evidence: []string{fmt.Sprintf("%d 个回合没有行动事件", player.QuietTurns)}, Explanation: "空白回合可能代表观望，也可能代表记录中断。", SuggestedAsk: "这些空白回合是有意保留，还是忘记记录？"})
		}
	}
	if len(events) > 0 && len(turns) > 0 && analysis.AverageEvents >= 3 {
		signals = append(signals, domain.StrategySignal{Code: "dense-log", Title: "事件记录密度较高", Severity: "积极", Evidence: []string{fmt.Sprintf("平均每回合 %.1f 个事件", analysis.AverageEvents)}, Explanation: "记录粒度足够细，可以进一步使用资源轨迹和行动标签寻找模式。", SuggestedAsk: "哪些标签最常与得分增长同时出现？"})
	}
	return signals
}

func buildRecommendations(
	analysis matchAnalysis,
) []domain.Recommendation {
	result :=
		[]domain.Recommendation{}
	if analysis.Coverage < 0.75 {
		result = append(result, domain.Recommendation{Priority: 1, Audience: "记录者", Title: "建立每回合最小记录模板", Reason: "当前部分回合没有行动事件，战报会出现叙事断层。", Steps: []string{"回合开始时写下关注点", "行动完成后选择一个标签", "若没有变化也记录一句旁白"}})
	}
	if analysis.ScoreSpread >= 8 {
		result = append(result, domain.Recommendation{Priority: 2, Audience: "全体玩家", Title: "定位第一个拉开分差的节点", Reason: "分数差距已经足以影响后续决策。", Steps: []string{"按回合查看分数变化", "找到首次出现大幅得分的行动", "讨论当时有哪些替代选择"}})
	}
	if analysis.Pace.ClosingIntensity >= 0.5 {
		result = append(result, domain.Recommendation{Priority: 3, Audience: "下局主持人", Title: "提前提醒收官条件", Reason: "高权重事件集中在游戏后段。", Steps: []string{"在中局检查剩余目标", "提示资源不足的席位", "为收官行动预留记录时间"}})
	}
	if len(result) == 0 {
		result = append(result, domain.Recommendation{Priority: 4, Audience: "全体玩家", Title: "保持当前记录节奏", Reason: "现有数据没有暴露明显的记录断点。", Steps: []string{"继续记录行动标签", "为关键时刻补充解释", "结束后回答一条复盘问题"}})
	}
	sort.Slice(
		result,
		func(
			i,
			j int,
		) bool {
			left := result[i].Priority
			right := result[j].Priority
			return left < right
		},
	)
	return result
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
