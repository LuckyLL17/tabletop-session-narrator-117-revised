package service

import (
	"fmt"
	"time"

	"t117/internal/domain"
	"t117/pkg/ids"
)

func (
	s *MatchService,
) OpenTurn(
	owner, matchID domain.ID, input TurnInput,
) (domain.Turn, error) {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return turnError(err)
	}
	if !domain.CanWriteTimeline(match.Status) {
		return turnError(fmt.Errorf("%w: 当前状态不能推进回合", domain.ErrConflict))
	}
	if len(seats) == 0 {
		return turnError(
			domain.ErrCapacity,
		)
	}
	number := match.TurnNumber + 1
	if !turnNumberIsValid(number) {
		return turnError(fmt.Errorf("%w: 回合编号必须为正数", domain.ErrInvalid))
	}
	if number > 1 {
		previous :=
			s.lastTurn(matchID)
		if previous.ID != "" {
			now := time.Now().UTC()
			previous.Status =
				domain.TurnClosed
			previous.ClosedAt = &now
			_ = s.store.SaveTurn(previous)
		}
	}
	active := seats[match.CurrentSeat%len(seats)]
	turn := domain.NewTurn(matchID, active.ID, number, input.Focus, time.Now().UTC())
	turn.ID =
		domain.ID(ids.New("turn"))
	if err := domain.ValidateTurn(turn, domain.Match{ID: matchID, TurnNumber: number}); err != nil {
		return turnError(err)
	}
	match.TurnNumber = number
	match.Revision++
	match.UpdatedAt = time.Now().UTC()
	if saveErr :=
		s.store.SaveMatch(
			match); saveErr != nil {
		return turnError(saveErr)
	}
	if saveErr :=
		s.store.SaveTurn(
			turn); saveErr != nil {
		return turnError(saveErr)
	}
	return turn, nil
}
func (
	s *MatchService,
) CloseTurn(
	owner, matchID, turnID domain.ID,
) error {
	match, seats, err :=
		s.Get(owner, matchID)
	if err != nil {
		return err
	}
	turn, ok :=
		s.store.FindTurn(turnID)
	if !ok ||
		turn.MatchID != matchID {
		return domain.ErrMissing
	}
	if turn.Status ==
		domain.TurnClosed {
		return domain.ErrConflict
	}
	now := time.Now().UTC()
	turn.Status =
		domain.TurnClosed
	turn.ClosedAt = &now
	if saveErr :=
		s.store.SaveTurn(
			turn); saveErr != nil {
		return saveErr
	}
	match.CurrentSeat = domain.NextSeat(match.CurrentSeat, len(seats))
	match.Revision++
	match.UpdatedAt = now
	return s.store.SaveMatch(match)
}
func (
	s *MatchService,
) lastTurn(
	matchID domain.ID,
) domain.Turn {
	turns :=
		s.store.TurnsForMatch(
			matchID)
	var latest domain.Turn
	for turnIndex := range turns {
		turn := turns[turnIndex]
		if turn.Number > latest.Number {
			latest = turn
		}
	}
	return latest
}

func turnNumberIsValid(number int) bool {
	return number > 0
}
