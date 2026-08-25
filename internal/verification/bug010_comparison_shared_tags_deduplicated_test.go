package verification

// Coverage source markers: sharedTags, MatchTags, compareRoute

import (
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func TestBug010ComparisonSharedTagsDeduplicated(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	owner := domain.ID("owner")
	gs := service.NewGameService(state)
	game, err := gs.Create(owner, service.GameInput{Name: "同类", MinPlayers: 1, MaxPlayers: 2, Tags: []string{"策略"}})
	if err != nil {
		t.Fatal(err)
	}
	game.Tags = []string{"策略", "策略"}
	if err := state.SaveGame(game); err != nil {
		t.Fatal(err)
	}
	ms := service.NewMatchService(state, nil)
	players := []service.PlayerInput{{Name: "甲", Color: "红"}}
	first, _, err := ms.Create(owner, service.MatchInput{GameID: string(game.ID), Name: "一", Players: players})
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := ms.Create(owner, service.MatchInput{GameID: string(game.ID), Name: "二", Players: players})
	if err != nil {
		t.Fatal(err)
	}
	cmp, err := ms.Compare(owner, []domain.ID{first.ID, second.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmp.SharedTags) != 1 || cmp.SharedTags[0] != "策略" {
		t.Fatalf("相同标签应只出现一次: %#v", cmp.SharedTags)
	}
}

func TestBug010RegressionHealth(t *testing.T) {
	if got := domain.EventKinds(); len(got) == 0 {
		t.Fatal("事件种类为空")
	}
}
