package store

import (
	"fmt"

	"t117/internal/domain"
)

// ResourceFloors computes, for a game, the highest enabled-variant resource floor
// per resource name. It is the store-side mirror of the service projection so the
// floor check inside a transaction can be expressed against the same data.
func ResourceFloors(game domain.Game) map[string]int {
	floors := map[string]int{}
	for _, variant := range game.Variants {
		if !variant.Enabled {
			continue
		}
		for name, floor := range variant.ResourceFloor {
			if floor > floors[name] {
				floors[name] = floor
			}
		}
	}
	return floors
}

// ApplyResourceDelta returns the resulting resource map after applying delta to
// current and rejects any value that would drop below the variant floor. It is a
// pure function over its inputs so the same projection can be computed before a
// transaction (to fail fast) and reapplied inside it (as the source of truth).
func ApplyResourceDelta(current, delta, floors map[string]int) (map[string]int, error) {
	next := map[string]int{}
	for name, value := range current {
		next[name] = value
	}
	for name, change := range delta {
		next[name] += change
		if next[name] < floors[name] {
			return nil, fmt.Errorf("%w: %s 不能低于 %d", domain.ErrInvalid, name, floors[name])
		}
	}
	return next, nil
}

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
	return collect(
		s.data.Matches,
		func(
			match domain.Match,
		) bool {
			return match.OwnerID ==
				owner
		},
	)
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
	event = normalizeEventForStore(event)
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

// EventRecord is the fully-prepared record that RecordEventTx commits in one
// transaction. The seat, event and (optional) milestone fields are pre-built by
// the caller; floors and delta drive the resource projection applied inside the
// transaction. Committing all four writes through a single Store.Update makes
// the event, resource/score projection and key-moment card atomic: either the
// whole event lands or none of it does, so a later failure cannot leave the
// event on the timeline with resources already deducted and the milestone card
// missing, nor allow a retry to deduct resources a second time.
type EventRecord struct {
	Seat      domain.Seat
	Delta     map[string]int
	Floors    map[string]int
	ScoreGap  int
	Event     domain.ActionEvent
	Milestone *domain.Milestone
}

// RecordEventTx commits an action event and every side effect it triggers —
// resource deduction, score change, the event row and an optional key-moment
// card — inside one transactional Store.Update. It is the single point that
// guarantees atomicity between the event write and the state projection + card
// creation described in the problem statement.
func (s *Store) RecordEventTx(record EventRecord) error {
	return s.Update(func(data *domain.Snapshot) error {
		seat, ok := data.Seats[record.Seat.ID]
		if !ok {
			return domain.ErrMissing
		}
		resources, err := ApplyResourceDelta(seat.Resources, record.Delta, record.Floors)
		if err != nil {
			return err
		}
		seat.Resources = resources
		seat.Score += record.ScoreGap
		data.Seats[seat.ID] = seat

		data.Events[record.Event.ID] = normalizeEventForStore(record.Event)

		if record.Milestone != nil {
			data.Milestones[record.Milestone.ID] = *record.Milestone
		}
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

func normalizeEventForStore(event domain.ActionEvent) domain.ActionEvent {
	if event.Delta == nil {
		event.Delta = map[string]int{}
	}
	return event
}
