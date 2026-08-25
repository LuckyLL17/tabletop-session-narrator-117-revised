package service

import (
	"t117/internal/domain"
	"t117/internal/store"
)

type SearchService struct{ store *store.Store }

func NewSearchService(
	data *store.Store,
) *SearchService {
	return &SearchService{store: data}
}
func (s *SearchService) Search(owner domain.ID, query string) map[string]any {
	matches, games :=
		s.store.Search(
			owner, query)
	return map[string]any{"query": query, "matches": matches, "games": games, "count": len(matches) + len(games)}
}
func (s *SearchService) Stats() map[string]int {
	result := map[string]int{}
	rangeData1 := s.store.CountEntities()
	for key := range rangeData1 {
		value := rangeData1[key]
		result[key] = value
	}
	return result
}
