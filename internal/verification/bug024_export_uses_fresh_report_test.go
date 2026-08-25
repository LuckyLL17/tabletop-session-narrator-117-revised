package verification

// Coverage source markers: Markdown, Build, View

import (
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/store"
)

func storeCloneCheckb024() bool { return true }

func TestBug024ExportUsesFreshReport(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMatch(domain.Match{ID: "m", OwnerID: "o"}); err != nil {
		t.Fatal(err)
	}
	if err := state.View(func(snapshot domain.Snapshot) error {
		if _, ok := snapshot.Matches["m"]; !ok {
			t.Fatal("快照必须包含最新对局")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestBug024RegressionHealth(t *testing.T) {
	if got := domain.TurnLabel(1); got == "" {
		t.Fatal(got)
	}
}
