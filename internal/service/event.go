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
	eventMatchID := turn.MatchID
	if eventMatchID == "" {
		eventMatchID = turn.MatchID
	}
	event := domain.NewEvent(eventMatchID, turn.ID, seat.ID, kind, strings.TrimSpace(input.Label), strings.TrimSpace(input.Detail), input.Delta, input.ScoreDelta, input.Weight, time.Now().UTC())
	event.ID =
		domain.ID(ids.New("event"))
	if err := domain.ValidateEvent(event, turn, seat); err != nil {
		return eventError(err)
	}
	if err := s.applyResourceChange(seat, event.Delta, match.GameID); err != nil {
		return eventError(err)
	}
	seat.Score +=
		event.ScoreDelta
	if saveErr :=
		s.store.SaveSeat(
			seat); saveErr != nil {
		return eventError(saveErr)
	}
	if saveErr :=
		s.store.SaveEvent(
			event); saveErr != nil {
		return eventError(saveErr)
	}
	if event.Kind == domain.EventMilestone || event.Weight >= 8 {
		_ =
			s.createMilestone(event)
	}
	return event, nil
}
func (
	s *MatchService,
) createMilestone(
	event domain.ActionEvent,
) error {
	item := domain.Milestone{ID: domain.ID(ids.New("milestone")), MatchID: event.MatchID, TurnID: event.TurnID, EventID: event.ID, Title: event.Label, Explanation: event.Detail, Importance: event.Weight, CreatedAt: event.CreatedAt}
	return s.store.SaveMilestone(item)
}
