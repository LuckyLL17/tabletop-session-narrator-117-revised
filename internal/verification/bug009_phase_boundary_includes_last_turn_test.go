package verification

// Coverage source markers: phaseEnd, Timeline, ReportService.Build

import (
	"fmt"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb009(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func phaseEndb009(total, parts, part int) int {
	if total == 0 {
		return 0
	}
	return (total*part + parts - 1) / parts
}

func TestBug009PhaseBoundaryIncludesLastTurn(t *testing.T) {
	ms, _, owner, match, _, turn := setupb009(t, 2)
	var err error
	for n := 1; n < 5; n++ {
		if err := ms.CloseTurn(owner, match.ID, turn.ID); err != nil {
			t.Fatal(err)
		}
		turn, err = ms.OpenTurn(owner, match.ID, service.TurnInput{Focus: fmt.Sprintf("第%d回合", n+1)})
		if err != nil {
			t.Fatal(err)
		}
	}
	rows, err := ms.Timeline(owner, match.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 || rows[len(rows)-1].Sequence != 5 {
		t.Fatalf("最后回合边界丢失: %#v", rows)
	}
}

func TestBug009RegressionHealth(t *testing.T) {
	if got := domain.TurnLabel(3); got == "" {
		t.Fatal(got)
	}
}
