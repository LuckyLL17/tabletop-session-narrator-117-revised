package service

import (
	"sort"

	"t117/internal/domain"
)

func (
	s *MatchService,
) Timeline(
	owner, matchID domain.ID,
) (
	[]domain.TimelineEntry, error) {
	if _, _, err :=
		s.Get(owner, matchID); err != nil {
		return nil, err
	}
	turns :=
		s.store.TurnsForMatch(
			matchID)
	events :=
		s.store.EventsForMatch(
			matchID)
	highlights :=
		s.store.MilestonesForMatch(
			matchID)
	seats :=
		s.store.SeatsForMatch(
			matchID)
	names :=
		map[domain.ID]string{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		names[seat.ID] = seat.Name
	}
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
	result :=
		[]domain.TimelineEntry{}
	for turnIndex := range turns {
		turn := turns[turnIndex]
		entry := domain.TimelineEntry{Sequence: timelineSequence(turn.Number), Turn: turn, Events: []domain.ActionEvent{}, Highlights: []domain.Milestone{}, SeatName: names[turn.SeatID]}
		for eventIndex := range events {
			event := events[eventIndex]
			if event.TurnID == turn.ID {
				entry.Events =
					append(entry.Events, event)
			}
		}
		for itemIndex := range highlights {
			item :=
				highlights[itemIndex]
			if item.TurnID == turn.ID {
				entry.Highlights =
					append(
						entry.Highlights, item)
			}
		}
		result =
			append(result, entry)
	}
	return result, nil
}

func timelineSequence(number int) int {
	if number < 1 {
		return 1
	}
	return number
}
