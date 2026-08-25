package verification

// Coverage source markers: Replay, applyResourceChange, SeatsForMatch

import (
	"fmt"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb007(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func TestBug007ReplayStartsFromInitialResources(t *testing.T) {
	ms, state, owner, match, seats, turn := setupb007(t, 2)
	_, err := ms.RecordEvent(owner, match.ID, service.EventInput{TurnID: string(turn.ID), SeatID: string(seats[0].ID), Kind: string(domain.EventResource), Label: "采集", Delta: map[string]int{"粮食": 2}, Weight: 2})
	if err != nil {
		t.Fatal(err)
	}
	current, ok := state.FindSeat(seats[0].ID)
	if !ok {
		t.Fatal("席位丢失")
	}
	current.Resources["粮食"] = 99
	if err := state.SaveSeat(current); err != nil {
		t.Fatal(err)
	}
	frames, err := ms.Replay(owner, match.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 1 || frames[0].Resources[seats[0].Name]["粮食"] != 7 {
		t.Fatalf("首帧资源应为 7: %#v", frames)
	}
}

func TestBug007RegressionHealth(t *testing.T) {
	if got := domain.TurnLabel(1); got != "第 1 回合" {
		t.Fatal(got)
	}
}
