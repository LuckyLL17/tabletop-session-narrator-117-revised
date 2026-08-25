package verification

// Coverage source markers: RecordEvent, ValidateEvent, EventsForMatch

import (
	"fmt"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func setupb003(t *testing.T, players int) (*service.MatchService, *store.Store, domain.ID, domain.Match, []domain.Seat, domain.Turn) {
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

func TestBug003ForeignEventRejected(t *testing.T) {
	ms, state, owner, match, _, _ := setupb003(t, 2)
	other, otherSeats, err := ms.Create(owner, service.MatchInput{GameID: string(match.GameID), Name: "第二局", Players: []service.PlayerInput{{Name: "乙", Color: "蓝"}, {Name: "丙", Color: "绿"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ms.Start(owner, other.ID); err != nil {
		t.Fatal(err)
	}
	otherTurn, err := ms.OpenTurn(owner, other.ID, service.TurnInput{Focus: "第二局"})
	if err != nil {
		t.Fatal(err)
	}
	_ = state
	_, err = ms.RecordEvent(owner, match.ID, service.EventInput{TurnID: string(otherTurn.ID), SeatID: string(otherSeats[0].ID), Kind: string(domain.EventAction), Label: "跨局行动", Weight: 3})
	if err == nil {
		t.Fatal("跨局事件应被拒绝")
	}
}

func TestBug003RegressionHealth(t *testing.T) {
	if _, err := domain.ParseEventKind(string(domain.EventAction)); err != nil {
		t.Fatal(err)
	}
}
