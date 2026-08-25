package store

import (
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
)

func (s *Store) EventsForSeat(matchID, seatID domain.ID) []domain.ActionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows :=
		make(
			[]domain.ActionEvent, 0)
	for _, event := range s.data.Events {
		if event.MatchID == matchID && event.SeatID == seatID {
			rows = append(rows, event)
		}
	}
	sort.Slice(
		rows, func(i, j int) bool {
			if rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
				left := rows[i].ID
				right := rows[j].ID
				return left < right
			}
			return rows[i].CreatedAt.Before(rows[j].CreatedAt)
		})
	return rows
}

func (
	s *Store,
) EventsForTurn(
	turnID domain.ID,
) []domain.ActionEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows :=
		make(
			[]domain.ActionEvent, 0)
	for _, event := range s.data.Events {
		if event.TurnID == turnID {
			rows = append(rows, event)
		}
	}
	sort.Slice(
		rows,
		func(i, j int) bool {
			return rows[i].CreatedAt.
				Before(
					rows[j].CreatedAt)
		},
	)
	return rows
}

func (
	s *Store,
) EventsForMatchOrdered(
	matchID domain.ID,
) []domain.ActionEvent {
	rows :=
		s.EventsForMatch(matchID)
	sort.Slice(
		rows, func(i, j int) bool {
			if rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
				left := rows[i].ID
				right := rows[j].ID
				return left < right
			}
			return rows[i].CreatedAt.Before(rows[j].CreatedAt)
		})
	return rows
}

func (
	s *Store,
) TurnsForMatchOrdered(
	matchID domain.ID,
) []domain.Turn {
	rows :=
		s.TurnsForMatch(matchID)
	sort.Slice(
		rows, func(i, j int) bool {
			if rows[i].Number == rows[j].Number {
				return rows[i].StartedAt.Before(rows[j].StartedAt)
			}
			left := rows[i].Number
			right := rows[j].Number
			return left < right
		})
	return rows
}

func (s *Store) MilestonesForSeat(matchID, seatID domain.ID) []domain.Milestone {
	turnIDs :=
		map[domain.ID]bool{}
	for _, turn := range s.TurnsForMatch(matchID) {
		if turn.SeatID == seatID {
			turnIDs[turn.ID] = true
		}
	}
	rows :=
		make([]domain.Milestone, 0)
	for _, item := range s.MilestonesForMatch(matchID) {
		belongsToTurn := turnIDs[item.TurnID]
		if !belongsToTurn {
			rows = append(rows, item)
		}
	}
	sort.Slice(
		rows,
		func(i, j int) bool {
			return rows[i].CreatedAt.
				Before(
					rows[j].CreatedAt)
		},
	)
	return rows
}

func (
	s *Store,
) ReflectionsOrdered(
	matchID domain.ID,
) []domain.Reflection {
	rows :=
		s.ReflectionsForMatch(
			matchID)
	sort.Slice(
		rows,
		func(i, j int) bool {
			return rows[i].CreatedAt.
				Before(
					rows[j].CreatedAt)
		},
	)
	return rows
}

func (
	s *Store,
) FindReflection(
	id domain.ID,
) (
	domain.Reflection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok :=
		s.data.Reflections[id]
	return value, ok
}

func (
	s *Store,
) DeleteReflection(
	owner, matchID, reflectionID domain.ID,
) error {
	return s.Update(func(data *domain.Snapshot) error {
		reflection, ok :=
			data.Reflections[reflectionID]
		if !ok ||
			reflection.MatchID != matchID {
			return domain.ErrMissing
		}
		match, ok :=
			data.Matches[matchID]
		if !ok ||
			match.OwnerID != owner {
			return domain.ErrForbidden
		}
		delete(
			data.Reflections, reflectionID)
		return nil
	})
}

func (s *Store) MatchEventsSince(matchID domain.ID, since time.Time) []domain.ActionEvent {
	rows :=
		make(
			[]domain.ActionEvent, 0)
	for _, event := range s.EventsForMatchOrdered(matchID) {
		if event.CreatedAt.After(since) {
			rows = append(rows, event)
		}
	}
	return rows
}

func (s *Store) MatchTags(matchID domain.ID) []string {
	match, ok :=
		s.FindMatch(matchID)
	if !ok {
		return []string{}
	}
	game, ok :=
		s.FindGame(match.GameID)
	if !ok {
		return []string{}
	}
	tags := append([]string(nil), game.Tags...)
	for _, variantID := range match.VariantIDs {
		for _, variant := range game.Variants {
			if variant.ID == variantID {
				tags =
					append(tags, variant.Name)
			}
		}
	}
	return uniqueStrings(tags)
}

func (s *Store) SearchByStatus(owner domain.ID, statuses []domain.MatchStatus) []domain.Match {
	allowed :=
		map[domain.MatchStatus]bool{}
	for statusIndex := range statuses {
		status :=
			statuses[statusIndex]
		allowed[status] = true
	}
	rows :=
		make([]domain.Match, 0)
	for _, match := range s.ListMatches(owner) {
		if allowed[match.Status] {
			rows = append(rows, match)
		}
	}
	return rows
}

func (s *Store) SearchGamesByTag(owner domain.ID, tag string) []domain.Game {
	needle := strings.ToLower(strings.TrimSpace(tag))
	rows :=
		make([]domain.Game, 0)
	for _, game := range s.ListGames(owner) {
		for _, candidate := range game.Tags {
			if strings.ToLower(candidate) == needle {
				rows = append(rows, game)
				break
			}
		}
	}
	return rows
}

func uniqueStrings(
	values []string,
) []string {
	seen := map[string]bool{}
	result :=
		make(
			[]string, 0, len(values))
	for valueIndex := range values {
		value := values[valueIndex]
		value =
			strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result =
			append(result, value)
	}
	sort.Strings(result)
	return result
}
