package service

import (
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
	// Events, turns and milestones come back from the store in map
	// iteration order, which Go randomizes per call. Two events in the
	// same turn would otherwise swap places between requests, and a
	// milestone created right after an action could land in the wrong
	// position. Use the ordered query variants (deterministic
	// CreatedAt/Number ordering with stable ID tiebreakers) so the
	// timeline is stable and milestones stay attached to their turn.
	turns :=
		s.store.TurnsForMatchOrdered(
			matchID)
	events :=
		s.store.EventsForMatchOrdered(
			matchID)
	highlights :=
		s.store.MilestonesForMatchOrdered(
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
	result :=
		[]domain.TimelineEntry{}
	for turnIndex := range turns {
		turn := turns[turnIndex]
		entry := domain.TimelineEntry{Sequence: turn.Number, Turn: turn, Events: []domain.ActionEvent{}, Highlights: []domain.Milestone{}, SeatName: names[turn.SeatID]}
		for eventIndex := range events {
			event := events[eventIndex]
			if event.TurnID == turn.ID {
				entry.Events =
					append(entry.Events, event)
			}
		}
		// Iterate every milestone. The previous code sliced the list in
		// half (len/2) and dropped the rest, which made highlight cards
		// disappear — especially when a milestone and an action were
		// created milliseconds apart and the surviving half landed on
		// the wrong side of the slice.
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
