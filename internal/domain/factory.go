package domain

import "time"

func NewGame(owner ID, name, summary string, minPlayers, maxPlayers int, now time.Time) Game {
	return Game{OwnerID: owner, Name: name, Summary: summary, MinPlayers: minPlayers, MaxPlayers: maxPlayers, DefaultResources: map[string]int{"金币": 0, "声望": 0}, Tags: []string{}, Variants: []RuleVariant{}, CreatedAt: now, UpdatedAt: now}
}

func NewMatch(owner, game ID, name string, now time.Time) Match {
	return Match{OwnerID: owner, GameID: game, Name: name, Status: MatchSetup, CurrentSeat: 0, TurnNumber: 0, VariantIDs: []ID{}, SeatIDs: []ID{}, Revision: 0, CreatedAt: now, UpdatedAt: now}
}

func NewSeat(match ID, name, color string, position int, resources map[string]int, now time.Time) Seat {
	return Seat{MatchID: match, Name: name, Color: color, Position: position, Resources: cloneInts(resources), JoinedAt: now}
}

func NewTurn(match, seat ID, number int, focus string, now time.Time) Turn {
	return Turn{MatchID: match, SeatID: seat, Number: number, Status: TurnOpen, Focus: focus, StartedAt: now}
}

func NewEvent(match, turn, seat ID, kind EventKind, label, detail string, delta map[string]int, score, weight int, now time.Time) ActionEvent {
	return ActionEvent{MatchID: match, TurnID: turn, SeatID: seat, Kind: kind, Label: label, Detail: detail, Delta: cloneInts(delta), ScoreDelta: score, Weight: weight, CreatedAt: now}
}

func cloneInts(
	source map[string]int,
) map[string]int {
	result := map[string]int{}
	for key := range source {
		value := source[key]
		result[key] = value
	}
	return result
}
