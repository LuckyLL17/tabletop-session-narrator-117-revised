package verification

// Coverage source markers: Prompts, Analyze, reflectionPrompts

import (
	"fmt"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb012(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func reflectSummaryb012(entries []domain.Reflection) domain.ReflectionSummary {
	r := domain.ReflectionSummary{Entries: entries, ByCategory: map[string]int{}}
	for _, e := range entries {
		r.Total++
		if e.Answer != "" {
			r.Answered++
		} else {
			r.OpenQuestions = append(r.OpenQuestions, e.Prompt)
		}
	}
	return r
}

func TestBug012ReflectionPromptDedup(t *testing.T) {
	ms, state, owner, match, _, _ := setupb012(t, 2)
	refs := service.NewReflectionService(state, ms)
	const prompt = "哪一个回合最早改变了你的计划？"
	if _, err := refs.Save(owner, match.ID, domain.ReflectionInput{Prompt: prompt, Answer: "已回答"}); err != nil {
		t.Fatal(err)
	}
	recs, err := refs.Prompts(owner, match.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range recs {
		for _, step := range rec.Steps {
			if step == prompt {
				t.Fatalf("已回答问题不应再次推荐: %#v", rec)
			}
		}
	}
}

func TestBug012RegressionHealth(t *testing.T) {
	if got := domain.TurnLabel(1); got == "" {
		t.Fatal(got)
	}
}
