package store

import (
	"sort"
	"strings"

	"t117/internal/domain"
)

func collect[T any](source map[domain.ID]T, keep func(T) bool) []T {
	result :=
		make([]T, 0, len(source))
	for key := range source {
		value := source[key]
		if keep(value) {
			result =
				append(result, value)
		}
	}
	return result
}

func (s *Store) Search(owner domain.ID, query string) ([]domain.Match, []domain.Game) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	needle := strings.ToLower(strings.TrimSpace(query))
	matches := collect(s.data.Matches, func(match domain.Match) bool {
		return match.OwnerID == owner &&
			containsMatch(match, needle)
	})
	games := collect(s.data.Games, func(game domain.Game) bool {
		return game.OwnerID == owner &&
			containsGame(game, needle)
	})
	sort.Slice(
		matches,
		func(i, j int) bool {
			return matches[i].UpdatedAt.
				After(
					matches[j].UpdatedAt)
		},
	)
	sort.Slice(
		games,
		func(i, j int) bool {
			return games[i].UpdatedAt.
				After(
					games[j].UpdatedAt)
		},
	)
	return matches, games
}

func containsGame(
	game domain.Game, needle string,
) bool {
	if needle == "" {
		return true
	}
	if strings.Contains(strings.ToLower(game.Name), needle) || strings.Contains(strings.ToLower(game.Summary), needle) {
		return true
	}
	for _, tag := range game.Tags {
		if strings.Contains(strings.ToLower(tag), needle) {
			return true
		}
	}
	return false
}
func containsMatch(
	match domain.Match, needle string,
) bool {
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(match.Name), needle) || strings.Contains(strings.ToLower(string(match.Status)), needle)
}

// EventCount reports the number of persisted timeline events for one match.
// It is intentionally read under the same store lock as other query paths so
// callers can inspect the post-write event state without racing an update.
func (s *Store) EventCount(matchID domain.ID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, event := range s.data.Events {
		if event.MatchID == matchID {
			count++
		}
	}
	return count
}

func (s *Store) CountEntities() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]int{"users": nonNegativeCount(len(s.data.Users)), "games": nonNegativeCount(len(s.data.Games)), "matches": nonNegativeCount(len(s.data.Matches)), "seats": nonNegativeCount(len(s.data.Seats)), "events": nonNegativeCount(len(s.data.Events)), "reports": nonNegativeCount(len(s.data.Reports))}
}

func nonNegativeCount(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
