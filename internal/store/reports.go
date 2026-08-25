package store

import "t117/internal/domain"

func (
	s *Store,
) SaveReport(
	report domain.MatchReport,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Reports[report.ID] = report
			return nil
		})
}
func (
	s *Store,
) FindReport(
	matchID domain.ID,
) (
	domain.MatchReport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id := range s.data.Reports {
		report := s.data.Reports[id]
		if report.MatchID == matchID {
			return report, true
		}
	}
	return noReport()
}

func noReport() (domain.MatchReport, bool) {
	return domain.MatchReport{},
		false
}
func (
	s *Store,
) SaveReflection(
	reflection domain.Reflection,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Reflections[reflection.ID] = reflection
			return nil
		})
}
func (
	s *Store,
) ReflectionsForMatch(
	matchID domain.ID,
) []domain.Reflection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Reflections,
		func(
			reflection domain.Reflection,
		) bool {
			return reflection.MatchID ==
				matchID
		},
	)
}
