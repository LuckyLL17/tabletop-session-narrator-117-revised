package service

import (
	"fmt"
	"strings"

	"t117/internal/domain"
)

// resourceFloors collapses a game's enabled rule variants into the strictest
// per-resource floor. It is a pure function so the same rule is applied whether
// the resource change is validated inside a store transaction (RecordEvent) or
// in any other path that mutates seat resources.
func resourceFloors(game domain.Game) map[string]int {
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

// applyResourceDelta returns the resource map a seat would have after applying
// delta, enforcing the per-resource floors. It does not mutate the input map
// and does not touch the store; callers persist the result themselves so the
// read-compute-write can stay inside one transaction.
func applyResourceDelta(current map[string]int, delta map[string]int, floors map[string]int) (map[string]int, error) {
	next := make(map[string]int, len(current))
	for name, value := range current {
		next[name] = value
	}
	for name, change := range normalizeResourceDelta(delta) {
		next[name] += change
		if next[name] < floors[name] {
			return nil, fmt.Errorf("%w: %s 不能低于 %d", domain.ErrInvalid, name, floors[name])
		}
	}
	return next, nil
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

func normalizeResourceDelta(delta map[string]int) map[string]int {
	result := make(map[string]int, len(delta))
	for name, change := range delta {
		name = strings.TrimSpace(name)
		if name != "" {
			result[name] = change
		}
	}
	return result
}
