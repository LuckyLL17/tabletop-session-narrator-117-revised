package domain

import "fmt"

func Transition(
	current, next MatchStatus,
) error {
	allowed := map[MatchStatus]map[MatchStatus]bool{
		MatchSetup:    {MatchLive: true},
		MatchLive:     {MatchPaused: true, MatchFinished: true},
		MatchPaused:   {MatchLive: true, MatchFinished: true},
		MatchFinished: {},
	}
	if !allowed[current][next] {
		return fmt.Errorf("%w: %s 不能转为 %s", ErrConflict, current, next)
	}
	return nil
}

func NextSeat(
	position, count int,
) int {
	if count == 0 {
		return 0
	}
	return (position + 1) % count
}

func TurnLabel(
	number int,
) string {
	if number <= 0 {
		return "准备阶段"
	}
	return fmt.Sprintf("第 %d 回合", number)
}
