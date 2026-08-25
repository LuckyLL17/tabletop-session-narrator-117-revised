package service

import (
	"strings"
	"time"

	"t117/internal/domain"
	"t117/internal/store"
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
	match, _, err :=
		s.Get(owner, matchID)
	if err != nil {
		return eventError(err)
	}
	if !domain.CanWriteTimeline(match.Status) {
		return eventError(
			domain.ErrConflict,
		)
	}
	kind, err :=
		domain.ParseEventKind(
			input.Kind)
	if err != nil {
		return eventError(err)
	}
	turn, ok :=
		s.store.FindTurn(
			domain.ID(input.TurnID))
	if !ok {
		return eventError(
			domain.ErrMissing,
		)
	}
	seat, ok :=
		s.store.FindSeat(
			domain.ID(input.SeatID))
	if !ok {
		return eventError(
			domain.ErrMissing,
		)
	}
	event := domain.NewEvent(matchID, turn.ID, seat.ID, kind, strings.TrimSpace(input.Label), strings.TrimSpace(input.Detail), input.Delta, input.ScoreDelta, input.Weight, time.Now().UTC())
	event.ID =
		domain.ID(ids.New("event"))
	if err := domain.ValidateEvent(event, turn, seat); err != nil {
		return eventError(err)
	}
	// Resolve the resource projection up front so an invalid deduction fails
	// before any write; the same projection is reapplied inside the
	// transaction by RecordEventTx so it stays the source of truth.
	game, ok :=
		s.store.FindGame(match.GameID)
	if !ok {
		return eventError(domain.ErrMissing)
	}
	floors :=
		store.ResourceFloors(game)
	resources, err :=
		store.ApplyResourceDelta(
			seat.Resources, event.Delta, floors)
	if err != nil {
		return eventError(err)
	}
	// Build the milestone record before touching the store. Only when this
	// preparatory step succeeds does RecordEventTx commit the event, the
	// resource/score projection and the card together — atomically.
	var milestone *domain.Milestone
	if event.Kind == domain.EventMilestone || event.Weight >= 8 {
		item :=
			domain.Milestone{ID: domain.ID(ids.New("milestone")), MatchID: event.MatchID, TurnID: event.TurnID, EventID: event.ID, Title: event.Label, Explanation: event.Detail, Importance: event.Weight, CreatedAt: event.CreatedAt}
		milestone = &item
	}
	if commitErr := s.store.RecordEventTx(store.EventRecord{
		Seat: seat, Delta: event.Delta, Floors: floors, ScoreGap: event.ScoreDelta, Event: event, Milestone: milestone,
	}); commitErr != nil {
		return eventError(commitErr)
	}
	// Keep the in-memory seat copy consistent with the committed projection so
	// callers relying on the returned event see the post-event state.
	seat.Resources = resources
	seat.Score += event.ScoreDelta
	_ = s.store.AppendAudit("event.recorded", map[string]any{"match_id": matchID, "event_id": event.ID, "seat_id": seat.ID})
	return event, nil
}
