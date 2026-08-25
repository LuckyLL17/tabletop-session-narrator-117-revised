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
	turns :=
		s.store.TurnsForMatch(
			matchID)
	events :=
		s.store.EventsForMatch(
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
	balances :=
		map[string]map[string]int{}
	scores := map[string]int{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		balances[seat.Name] = map[string]int{}
		for key, value := range seat.Resources {
			balances[seat.Name][key] = value
		}
		rangeData1 :=
			seat.Resources
		for key := range rangeData1 {
			value := rangeData1[key]
			balances[seat.Name][key] = value
		}
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
