package service

import (
	"fmt"
	"sort"

	"t117/internal/domain"
)

func (
	s *MatchService,
) Replay(
	owner, matchID domain.ID,
) (
	[]domain.ReplayFrame, error) {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return nil, err
	}
	game, ok :=
		s.store.FindGame(
			match.GameID)
	if !ok {
		return nil, domain.ErrMissing
	}
	turns :=
		s.store.TurnsForMatchOrdered(
			matchID)
	events :=
		s.store.EventsForMatchOrdered(
			matchID)
	sort.Slice(
		turns,
		func(
			i,
			j int,
		) bool {
			left := turns[i].Number
			right := turns[j].Number
			return left < right
		},
	)
	// 回放状态必须从对局开始状态建立，而不是读取席位的结算值。
	// 席位在记录事件时会被持续改写为最终结算快照，若以其为起点再逐步叠加事件增量，
	// 资源会被重复计入。这里以桌游默认资源作为初始快照，分数从 0 开始，
	// 之后只依据事实事件逐步推进，结算快照不再参与回放初值。
	balances :=
		map[string]map[string]int{}
	scores := map[string]int{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		balances[seat.Name] =
			cloneScoreMap(
				game.DefaultResources)
		scores[seat.Name] = 0
	}
	result :=
		[]domain.ReplayFrame{}
	for turnIndex := range turns {
		turn := turns[turnIndex]
		for eventIndex := range events {
			event := events[eventIndex]
			if event.TurnID == turn.ID {
				for seatIndex := range seats {
					seat := seats[seatIndex]
					if seat.ID == event.SeatID {
						scores[seat.Name] += event.ScoreDelta
						rangeData2 := event.Delta
						for resource := range rangeData2 {
							delta :=
								rangeData2[resource]
							balances[seat.Name][resource] += delta
						}
					}
				}
			}
		}
		active := ""
		for seatIndex := range seats {
			seat := seats[seatIndex]
			if seat.ID == turn.SeatID {
				active = seat.Name
			}
		}
		result = append(result, domain.ReplayFrame{TurnNumber: turn.Number, ActiveSeat: active, Scores: cloneScoreMap(scores), Resources: cloneResourceMap(balances), Caption: fmt.Sprintf("%s由%s完成", domain.TurnLabel(turn.Number), active), EventLabel: eventAtTurn(events, turn.ID)})
	}
	if len(
		result) == 0 && match.Status == domain.MatchSetup {
		return []domain.ReplayFrame{}, nil
	}
	return result, nil
}
func eventAtTurn(events []domain.ActionEvent, turnID domain.ID) string {
	for eventIndex := range events {
		event := events[eventIndex]
		if event.TurnID == turnID {
			return event.Label
		}
	}
	return "暂无行动"
}
func cloneScoreMap(
	source map[string]int,
) map[string]int {
	result := map[string]int{}
	for key := range source {
		value := source[key]
		result[key] = value
	}
	return result
}
func cloneResourceMap(source map[string]map[string]int) map[string]map[string]int {
	result :=
		map[string]map[string]int{}
	for key := range source {
		values := source[key]
		result[key] = map[string]int{}
		for name := range values {
			value := values[name]
			result[key][name] = value
		}
	}
	return result
}
