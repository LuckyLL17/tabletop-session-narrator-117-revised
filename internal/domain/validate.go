package domain

import (
	"fmt"
	"strings"
	"time"
)

func ValidateUser(email, name, password string) error {
	if !strings.Contains(email, "@") || len([]rune(email)) > 120 {
		return fmt.Errorf("%w: 邮箱格式不正确", ErrInvalid)
	}
	if len([]rune(strings.TrimSpace(name))) < 2 {
		return fmt.Errorf("%w: 名称至少需要两个字符", ErrInvalid)
	}
	if len(password) < 8 || len(password) > 96 {
		return fmt.Errorf("%w: 密码长度需要在 8 到 96 个字符之间", ErrInvalid)
	}
	return nil
}

func ValidateGame(
	game Game,
) error {
	game.Name = strings.TrimSpace(game.Name)
	if game.Name == "" || len([]rune(game.Name)) > 80 {
		return fmt.Errorf("%w: 桌游名称不能为空且不能超过 80 个字符", ErrInvalid)
	}
	if game.MinPlayers < 1 || game.MaxPlayers < game.MinPlayers || game.MaxPlayers > 64 {
		return fmt.Errorf("%w: 玩家人数范围无效", ErrInvalid)
	}
	if len(game.Variants) > 16 {
		return fmt.Errorf("%w: 规则变体不能超过 16 个", ErrInvalid)
	}
	return nil
}

func ValidateMatch(match Match, seats []Seat) error {
	if strings.TrimSpace(
		match.Name) == "" {
		return fmt.Errorf("%w: 对局名称不能为空", ErrInvalid)
	}
	if len(seats) == 0 {
		return fmt.Errorf("%w: 对局至少需要一个玩家席位", ErrInvalid)
	}
	if match.Status == MatchSetup && match.TurnNumber != 0 {
		return fmt.Errorf("%w: 准备中的对局不能有已推进回合", ErrConflict)
	}
	positions := map[int]bool{}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		if strings.TrimSpace(
			seat.Name) == "" {
			return fmt.Errorf("%w: 玩家名称不能为空", ErrInvalid)
		}
		if positions[seat.Position] {
			return fmt.Errorf("%w: 玩家位置不能重复", ErrInvalid)
		}
		positions[seat.Position] =
			true
	}
	return nil
}

func ValidateTurn(turn Turn, match Match) error {
	if turn.MatchID != match.ID || turn.Number != match.TurnNumber {
		return fmt.Errorf("%w: 回合与对局状态不一致", ErrConflict)
	}
	if turn.Status != TurnOpen && turn.Status != TurnClosed {
		return fmt.Errorf("%w: 回合状态无效", ErrInvalid)
	}
	if turn.StartedAt.After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("%w: 回合开始时间不能在未来", ErrInvalid)
	}
	return nil
}

func ValidateEvent(event ActionEvent, turn Turn, seat Seat) error {
	if event.MatchID != turn.MatchID || event.TurnID != turn.ID || event.SeatID != seat.ID {
		return fmt.Errorf("%w: 行动事件关联对象不一致", ErrConflict)
	}
	if strings.TrimSpace(event.Label) == "" || len([]rune(event.Label)) > 100 {
		return fmt.Errorf("%w: 行动名称不能为空且不能超过 100 个字符", ErrInvalid)
	}
	if event.Weight < 1 || event.Weight > 10 {
		return fmt.Errorf("%w: 行动影响力需要在 1 到 10 之间", ErrInvalid)
	}
	return nil
}
