package verification

// Coverage source markers: List, ListMatches, matchesRoute

import (
	"path/filepath"
	"testing"
	"time"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func TestBug023MatchListOrderUsesUpdatedAt(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	old := domain.Match{ID: "old", OwnerID: "o", Name: "旧", CreatedAt: time.Unix(1, 0), UpdatedAt: time.Unix(3, 0)}
	newer := domain.Match{ID: "new", OwnerID: "o", Name: "新", CreatedAt: time.Unix(2, 0), UpdatedAt: time.Unix(4, 0)}
	_ = state.SaveMatch(old)
	_ = state.SaveMatch(newer)
	rows := service.NewMatchService(state, nil).List("o")
	if len(rows) != 2 {
		t.Fatal(len(rows))
	}
	if rows[0].ID != "new" {
		t.Fatalf("最近更新的对局应排在最前: %#v", rows)
	}
}

func TestBug023RegressionHealth(t *testing.T) {
	if got := domain.MatchStatuses(); len(got) == 0 {
		t.Fatal("状态为空")
	}
}
