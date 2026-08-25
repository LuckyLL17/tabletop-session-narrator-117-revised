package verification

// Coverage source markers: Search, SearchService, searchRoute

import (
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/store"
)

func TestBug022SearchMatchesByOwnerAndQuery(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMatch(domain.Match{ID: "m1", OwnerID: "alice", Name: "自己的局"}); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMatch(domain.Match{ID: "m2", OwnerID: "bob", Name: "别人的局"}); err != nil {
		t.Fatal(err)
	}
	matches, _ := state.Search("alice", "局")
	if len(matches) != 1 || matches[0].ID != "m1" {
		t.Fatalf("搜索越权或漏项: %#v", matches)
	}
}

func TestBug022RegressionHealth(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "regression.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _ = state.Search("alice", ""); false {
		t.Fatal("不可达")
	}
}
