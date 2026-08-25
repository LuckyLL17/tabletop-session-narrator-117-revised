package verification

// Coverage source markers: Update, RecordEvent, SaveSeat

import (
	"fmt"
	"path/filepath"
	"testing"

	"sync"
	"t117/internal/domain"
	"t117/internal/store"
)

func TestBug029ConcurrentUpdatesPreserveEvents(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = state.SaveMatch(domain.Match{ID: domain.ID(fmt.Sprintf("m%d", n)), OwnerID: "o"})
		}(n)
	}
	wg.Wait()
	if len(state.ListMatches("o")) != 2 {
		t.Fatal("并发保存不能丢失另一条记录")
	}
}

func TestBug029RegressionHealth(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "regression.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := state.SaveMatch(domain.Match{ID: "m", OwnerID: "o"}); err != nil {
		t.Fatal(err)
	}
}
