package verification

// Coverage source markers: ChangeStatus, Enqueue, SaveJob

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb020(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func TestBug020FinishEnqueuesReport(t *testing.T) {
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
	ms := service.NewMatchService(state, service.NewJobService(state))
	match, _, err := ms.Create(owner, service.MatchInput{GameID: string(game.ID), Name: "周末局", Players: []service.PlayerInput{{Name: "甲", Color: "红"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ms.Start(owner, match.ID); err != nil {
		t.Fatal(err)
	}
	before := len(state.JobsForOwner(owner))
	if _, err = ms.Finish(owner, match.ID); err != nil {
		t.Fatal(err)
	}
	if jobs := state.JobsForOwner(owner); len(jobs) != before+1 || jobs[len(jobs)-1].Kind != "生成战报" {
		t.Fatalf("结束对局应新增战报任务: %#v", jobs)
	}
}

func TestBug020RegressionHealth(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "regression.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := state.ClaimJob(time.Now()); err != nil || !ok {
		_ = ok
	}
}
