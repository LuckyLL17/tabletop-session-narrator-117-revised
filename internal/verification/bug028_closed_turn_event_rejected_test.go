package verification

// Coverage source markers: CloseTurn, RecordEvent, ValidateEvent

import (
	"fmt"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb028(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
	t.Helper()
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	owner := domain.ID("owner")
	gs := service.NewGameService(state)
	game, err := gs.Create(owner, service.GameInput{Name: "局势", Summary: "复盘", MinPlayers: 2, MaxPlayers: 4, DefaultResources: map[string]int{"粮食": 5, "金币": 2}})
	if err != nil {
		t.Fatal(err)
	}
	inputs := make([]service.PlayerInput, players)
	for n := range inputs {
		inputs[n] = service.PlayerInput{Name: fmt.Sprintf("玩家%d", n+1), Color: fmt.Sprintf("色%d", n+1)}
	}
	ms := service.NewMatchService(state, nil)
	match, seats, err := ms.Create(owner, service.MatchInput{GameID: string(game.ID), Name: "周末局", Players: inputs})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ms.Start(owner, match.ID); err != nil {
		t.Fatal(err)
	}
	turn, err := ms.OpenTurn(owner, match.ID, service.TurnInput{Focus: "开局"})
	if err != nil {
		t.Fatal(err)
	}
	match, _, _ = ms.Get(owner, match.ID)
	return ms, state, owner, match, seats, turn
}

func turnIsClosedb028(s *store.Store, id string) bool {
	turn, ok := s.FindTurn(domain.ID(id))
	return ok && turn.Status == domain.TurnClosed
}

func TestBug028ClosedTurnEventRejected(t *testing.T) {
	ms, _, owner, match, seats, turn := setupb028(t, 2)
	if err := ms.CloseTurn(owner, match.ID, turn.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := ms.RecordEvent(owner, match.ID, service.EventInput{TurnID: string(turn.ID), SeatID: string(seats[0].ID), Kind: string(domain.EventAction), Label: "关后", Weight: 1}); err == nil {
		t.Fatal("关闭回合应拒绝事件")
	}
}

func TestBug028RegressionHealth(t *testing.T) {
	if err := domain.Transition(domain.MatchLive, domain.MatchFinished); err != nil {
		t.Fatal(err)
	}
}
