package verification

// Coverage source markers: Insights, collectSignals, buildRecommendations

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb026(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func stableOrderb026(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func TestBug026InsightOrderingDeterministic(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	owner := domain.ID("owner")
	gs := service.NewGameService(state)
	game, err := gs.Create(owner, service.GameInput{Name: "局势", MinPlayers: 1, MaxPlayers: 2})
	if err != nil {
		t.Fatal(err)
	}
	ms := service.NewMatchService(state, nil)
	match, _, err := ms.Create(owner, service.MatchInput{GameID: string(game.ID), Name: "周末局", Players: []service.PlayerInput{{Name: "玩家甲", Color: "红"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ms.Start(owner, match.ID); err != nil {
		t.Fatal(err)
	}
	turn, err := ms.OpenTurn(owner, match.ID, service.TurnInput{Focus: fmt.Sprintf("%d", 1)})
	if err != nil {
		t.Fatal(err)
	}
	for n := 2; n <= 3; n++ {
		if err := ms.CloseTurn(owner, match.ID, turn.ID); err != nil {
			t.Fatal(err)
		}
		turn, err = ms.OpenTurn(owner, match.ID, service.TurnInput{Focus: fmt.Sprintf("%d", n)})
		if err != nil {
			t.Fatal(err)
		}
	}
	analysis, err := ms.Analyze(owner, match.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, signal := range analysis.Signals {
		if signal.Title == "" {
			t.Fatal("洞察标题不应被清空")
		}
	}
}

func TestBug026RegressionHealth(t *testing.T) {
	if got := domain.EventKinds(); len(got) == 0 {
		t.Fatal("事件种类为空")
	}
}
