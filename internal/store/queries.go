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
	// 对局没有自带标签字段，标签来自其所属桌游档案，因此随对局一起解析出桌游，
	// 让名称和标签筛选都在当前用户范围下用 AND 组合，避免或与和关系错位导致归属漂移。
	matches := collect(s.data.Matches, func(match domain.Match) bool {
		if match.OwnerID != owner {
			return false
		}
		game, ok := s.data.Games[match.GameID]
		if !ok {
			return containsMatch(match, nil, needle)
		}
		return containsMatch(match, &game, needle)
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
	match domain.Match, game *domain.Game, needle string,
) bool {
	if needle == "" {
		return true
	}
	if strings.Contains(strings.ToLower(match.Name), needle) || strings.Contains(strings.ToLower(string(match.Status)), needle) {
		return true
	}
	// 对局继承所属桌游的标签，标签筛选要落到这里，与名称条件用 OR 组合，
	// 但整段 OR 再受调用方的当前用户范围 AND 约束，确保归属稳定。
	if game != nil {
		for _, tag := range game.Tags {
			if strings.Contains(strings.ToLower(tag), needle) {
				return true
			}
		}
		for _, variant := range game.Variants {
			if strings.Contains(strings.ToLower(variant.Name), needle) {
				return true
			}
		}
	}
	return false
}
func (s *Store) CountEntities() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]int{"users": len(s.data.Users), "games": len(s.data.Games), "matches": len(s.data.Matches), "seats": len(s.data.Seats), "events": len(s.data.Events), "reports": len(s.data.Reports)}
}
