package service

import (
	"t117/internal/domain"
)

// ResourceLedger returns the current resource map per seat for a match. It is a
// read-only projection; all resource mutations are committed transactionally by
// store.RecordEventTx so the resource change and the event that caused it land
// (or fail) together.
func (s *MatchService) ResourceLedger(matchID domain.ID) map[string]map[string]int {
	seats :=
		s.store.SeatsForMatch(
			matchID)
	result :=
		map[string]map[string]int{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		result[seat.Name] =
			seat.Resources
	}
	return result
}
