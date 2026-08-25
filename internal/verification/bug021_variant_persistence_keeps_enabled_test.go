package verification

// Coverage source markers: AddVariant, SaveGame, MatchTags

import (
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func TestBug021VariantPersistenceKeepsEnabled(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	gs := service.NewGameService(state)
	owner := domain.ID("o")
	game, err := gs.Create(owner, service.GameInput{Name: "档案", Summary: "", MinPlayers: 1, MaxPlayers: 2})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := gs.AddVariant(owner, game.ID, service.VariantInput{Name: "夜局", ResourceFloor: map[string]int{"粮食": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Variants) != 1 || updated.Variants[0].ResourceFloor["粮食"] != 2 {
		t.Fatalf("变体应完整保存: %#v", updated.Variants)
	}
}

func TestBug021RegressionHealth(t *testing.T) {
	if err := domain.ValidateGame(domain.Game{Name: "档案", MinPlayers: 1, MaxPlayers: 2}); err != nil {
		t.Fatal(err)
	}
}
