package verification

// Coverage source markers: applyResourceChange, AddVariant, SaveSeat

import (
	"fmt"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb004(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func TestBug004VariantFloorActivation(t *testing.T) {
	_, state, owner, match, _, _ := setupb004(t, 2)
	gs := service.NewGameService(state)
	game, err := gs.Get(owner, match.GameID)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := gs.AddVariant(owner, game.ID, service.VariantInput{Name: "下限", ResourceFloor: map[string]int{"粮食": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Variants) != 1 || !updated.Variants[0].Enabled {
		t.Fatalf("新增变体必须启用: %#v", updated.Variants)
	}
}

func TestBug004RegressionHealth(t *testing.T) {
	if err := domain.ValidateGame(domain.Game{ID: "g", OwnerID: "o", Name: "局", MinPlayers: 1, MaxPlayers: 2}); err != nil {
		t.Fatal(err)
	}
}
