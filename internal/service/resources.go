package service

import (
	"fmt"

	"t117/internal/domain"
)

func (s *MatchService) applyResourceChange(seat domain.Seat, delta map[string]int, gameID domain.ID) (domain.Seat, error) {
	game, ok :=
		s.store.FindGame(gameID)
	if !ok {
		return seat, domain.ErrMissing
	}
	floors := map[string]int{}
	for _, variant := range game.Variants {
		if variant.Enabled {
			rangeData2 :=
				variant.ResourceFloor
			for name := range rangeData2 {
				floor := rangeData2[name]
				if floor > floors[name] {
					floors[name] = floor
				}
			}
		}
	}
	next := map[string]int{}
	rangeData3 :=
		seat.Resources
	for name := range rangeData3 {
		value := rangeData3[name]
		next[name] = value
	}
	for name := range delta {
		change := delta[name]
		next[name] += change
		if next[name] < floors[name] {
			return seat, fmt.Errorf("%w: %s 不能低于 %d", domain.ErrInvalid, name, floors[name])
		}
	}
	seat.Resources = next
	return seat, nil
}
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
