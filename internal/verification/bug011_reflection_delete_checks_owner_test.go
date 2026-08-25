package verification

// Coverage source markers: Delete, DeleteReflection, reflectionRoute

import (
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/store"
)

func TestBug011ReflectionDeleteChecksOwner(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	owner, other := domain.ID("owner"), domain.ID("other")
	if err := state.SaveReflection(domain.Reflection{ID: "r1", MatchID: "m1", Prompt: "问题", Answer: "回答"}); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMatch(domain.Match{ID: "m1", OwnerID: owner}); err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMatch(domain.Match{ID: "m2", OwnerID: other}); err != nil {
		t.Fatal(err)
	}
	if err := state.DeleteReflection(other, "m2", "r1"); err == nil {
		t.Fatal("反思必须校验所属对局")
	}
}

func TestBug011RegressionHealth(t *testing.T) {
	if got := domain.TurnLabel(1); got == "" {
		t.Fatal(got)
	}
}
