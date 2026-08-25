package verification

// Coverage source markers: ClaimJob, Claim, RunOnce

import (
	"path/filepath"
	"testing"
	"time"

	"t117/internal/domain"
	"t117/internal/jobs"
	"t117/internal/store"
)

func TestBug013JobClaimRespectsSchedule(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(2 * time.Hour)
	job := domain.Job{ID: "j1", OwnerID: "o", MatchID: "m", State: domain.JobRetry, ScheduledAt: future}
	if err := state.SaveJob(job); err != nil {
		t.Fatal(err)
	}
	type claimResult struct {
		job domain.Job
		ok  bool
		err error
	}
	results := make(chan claimResult, 1)
	go func() {
		job, ok, err := state.ClaimJob(time.Now())
		results <- claimResult{job: job, ok: ok, err: err}
	}()
	result := <-results
	got, ok, err := result.job, result.ok, result.err
	if err != nil {
		t.Fatal(err)
	}
	if ok || got.ID != "" {
		t.Fatalf("未到时间的重试不应被领取: %#v %v", got, ok)
	}
}

func TestBug013RegressionHealth(t *testing.T) {
	if got := jobs.RetryWindow(1); got <= 0 {
		t.Fatal(got)
	}
}
