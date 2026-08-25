package service

import (
	"strings"
	"time"

	"t117/internal/domain"
	"t117/pkg/ids"
)

func (
	s *MatchService,
) RecordEvent(
	owner, matchID domain.ID, input EventInput,
) (
	domain.ActionEvent,
	error,
) {
	// The whole read-compute-write runs inside one store transaction so that two
	// events submitted in the same second are serialized. Each call reads the
	// match, turn, seat and game from a single coherent snapshot, computes the
	// new seat/resources and the new event, and writes them all back together.
	// Previously the read (FindTurn/FindSeat) and the writes (SaveSeat/SaveEvent)
	// happened across separate lock acquisitions, so a concurrent submit could
	// read the other player's stale seat snapshot and overwrite it — losing one
	// player's resource update or one of the two events.
	var created domain.ActionEvent
	err := s.store.Update(func(data *domain.Snapshot) error {
		match, ok := data.Matches[matchID]
		if !ok || match.OwnerID != owner {
			return domain.ErrMissing
		}
		if !domain.CanWriteTimeline(match.Status) {
			return domain.ErrConflict
		}
		kind, err := domain.ParseEventKind(input.Kind)
		if err != nil {
			return err
		}
		turn, ok := data.Turns[domain.ID(input.TurnID)]
		if !ok {
			return domain.ErrMissing
		}
		seat, ok := data.Seats[domain.ID(input.SeatID)]
		if !ok {
			return domain.ErrMissing
		}
		event := domain.NewEvent(matchID, turn.ID, seat.ID, kind, strings.TrimSpace(input.Label), strings.TrimSpace(input.Detail), input.Delta, input.ScoreDelta, input.Weight, time.Now().UTC())
		event.ID = domain.ID(ids.New("event"))
		if err := domain.ValidateEvent(event, turn, seat); err != nil {
			return err
		}
		seat, event = prepareEventWrite(seat, event)
		game, ok := data.Games[match.GameID]
		if !ok {
			return domain.ErrMissing
		}
		resources, err := applyResourceDelta(seat.Resources, event.Delta, resourceFloors(game))
		if err != nil {
			return err
		}
		seat.Resources = resources
		seat.Score += event.ScoreDelta
		data.Seats[seat.ID] = seat
		data.Events[event.ID] = event
		if event.Kind == domain.EventMilestone || event.Weight >= 8 {
			item := domain.Milestone{ID: domain.ID(ids.New("milestone")), MatchID: event.MatchID, TurnID: event.TurnID, EventID: event.ID, Title: event.Label, Explanation: event.Detail, Importance: event.Weight, CreatedAt: event.CreatedAt}
			data.Milestones[item.ID] = item
		}
		created = event
		return nil
	})
	if err != nil {
		return eventError(err)
	}
	return created, nil
}

func prepareEventWrite(seat domain.Seat, event domain.ActionEvent) (domain.Seat, domain.ActionEvent) {
	if event.Delta == nil {
		return seat, event
	}
	delta := make(map[string]int, len(event.Delta))
	for key, value := range event.Delta {
		delta[key] = value
	}
	event.Delta = delta
	return seat, event
}
