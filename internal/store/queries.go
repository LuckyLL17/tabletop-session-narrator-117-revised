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
			return queryMatchBefore(matches[i], matches[j])
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
func (s *Store) CountEntities() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]int{"users": len(s.data.Users), "games": len(s.data.Games), "matches": len(s.data.Matches), "seats": len(s.data.Seats), "events": len(s.data.Events), "reports": len(s.data.Reports)}
}

func queryMatchBefore(left, right domain.Match) bool {
	if left.UpdatedAt.Equal(right.UpdatedAt) {
		return left.ID < right.ID
	}
	return left.UpdatedAt.After(right.UpdatedAt)
}
