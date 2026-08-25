package store

import (
	"sort"

	"t117/internal/domain"
)

func (
	s *Store,
) SaveMatch(
	match domain.Match,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Matches[match.ID] = match
			return nil
		})
}
func (
	s *Store,
) FindMatch(
	id domain.ID,
) (domain.Match, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	match, ok :=
		s.data.Matches[id]
	return match, ok
}
func (
	s *Store,
) ListMatches(
	owner domain.ID,
) []domain.Match {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := collect(
		s.data.Matches,
		func(
			match domain.Match,
		) bool {
			return match.OwnerID ==
				owner
		},
	)
	sort.Slice(rows, func(i, j int) bool { return matchListBefore(rows[i], rows[j]) })
	return rows
}
func (
	s *Store,
) SaveSeat(
	seat domain.Seat,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Seats[seat.ID] = seat
			return nil
		})
}
func (
	s *Store,
) FindSeat(
	id domain.ID,
) (domain.Seat, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seat, ok :=
		s.data.Seats[id]
	return seat, ok
}
func (
	s *Store,
) SeatsForMatch(
	matchID domain.ID,
) []domain.Seat {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Seats,
		func(
			seat domain.Seat,
		) bool {
			return seat.MatchID ==
				matchID
		},
	)
}
func (
	s *Store,
) SaveTurn(
	turn domain.Turn,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Turns[turn.ID] = turn
			return nil
		})
}
func (
	s *Store,
) FindTurn(
	id domain.ID,
) (domain.Turn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	turn, ok :=
		s.data.Turns[id]
	return turn, ok
}
func (
	s *Store,
) TurnsForMatch(
	matchID domain.ID,
) []domain.Turn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Turns,
		func(
			turn domain.Turn,
		) bool {
			return turn.MatchID ==
				matchID
		},
	)
}
func (
	s *Store,
) SaveEvent(
	event domain.ActionEvent,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Events[event.ID] = event
			return nil
		})
}
func (
	s *Store,
) EventsForMatch(
	matchID domain.ID,
) []domain.ActionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Events,
		func(
			event domain.ActionEvent,
		) bool {
			return event.MatchID ==
				matchID
		},
	)
}
func (
	s *Store,
) SaveMilestone(
	item domain.Milestone,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Milestones[item.ID] = item
			return nil
		})
}
func (
	s *Store,
) MilestonesForMatch(
	matchID domain.ID,
) []domain.Milestone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Milestones,
		func(
			item domain.Milestone,
		) bool {
			return item.MatchID ==
				matchID
		},
	)
}

func matchListBefore(left, right domain.Match) bool {
	if left.UpdatedAt.Equal(right.UpdatedAt) {
		return left.ID < right.ID
	}
	return left.UpdatedAt.After(right.UpdatedAt)
}
