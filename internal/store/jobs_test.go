package store

import (
	"path/filepath"
	"testing"
	"time"

	"t117/internal/domain"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "jobs.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

// A retry job whose scheduled time is still in the future must not be claimed.
// This reproduces the duplicate-report / runaway-attempts bug where the poller
// re-claimed a still-waiting failure ahead of its schedule.
func TestClaimJob_SkipsRetryScheduledInTheFuture(t *testing.T) {
	store := newTestStore(t)

	now := time.Date(2026, 8, 25, 14, 56, 0, 0, time.UTC)
	earlyJob := domain.Job{ID: "job-demo-013", OwnerID: "owner", MatchID: "match", Kind: "report", State: domain.JobRetry, Attempts: 1, ScheduledAt: now.Add(29 * time.Second)}
	if err := store.SaveJob(earlyJob); err != nil {
		t.Fatalf("save early job: %v", err)
	}
	readyJob := domain.Job{ID: "job-demo-ready", OwnerID: "owner", MatchID: "match", Kind: "report", State: domain.JobRetry, Attempts: 1, ScheduledAt: now.Add(-time.Second)}
	if err := store.SaveJob(readyJob); err != nil {
		t.Fatalf("save ready job: %v", err)
	}

	got, ok, err := store.ClaimJob(now)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if !ok {
		t.Fatal("expected to claim the ready job")
	}
	if got.ID != readyJob.ID {
		t.Fatalf("expected ready job %q, got %q", readyJob.ID, got.ID)
	}

	// The still-waiting retry must remain unclaimed and unchanged.
	early, ok, err := store.ClaimJob(now)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if ok {
		t.Fatalf("future-scheduled retry must not be claimed, got job %q state=%q attempts=%d", early.ID, early.State, early.Attempts)
	}
}

// Once the scheduled time arrives, the retry becomes claimable.
func TestClaimJob_ClaimsRetryOnceScheduleArrives(t *testing.T) {
	store := newTestStore(t)

	now := time.Date(2026, 8, 25, 14, 56, 0, 0, time.UTC)
	future := now.Add(time.Minute)
	job := domain.Job{ID: "job-demo-013", OwnerID: "owner", MatchID: "match", Kind: "report", State: domain.JobRetry, Attempts: 1, ScheduledAt: future}
	if err := store.SaveJob(job); err != nil {
		t.Fatalf("save: %v", err)
	}

	if _, ok, err := store.ClaimJob(now); err != nil || ok {
		t.Fatalf("before schedule: expected no claim, got ok=%v err=%v", ok, err)
	}

	got, ok, err := store.ClaimJob(future)
	if err != nil {
		t.Fatalf("claim at schedule: %v", err)
	}
	if !ok || got.ID != job.ID {
		t.Fatalf("at schedule: expected to claim %q, got ok=%v job=%q", job.ID, ok, got.ID)
	}
	if got.State != domain.JobRunning {
		t.Fatalf("expected state %q, got %q", domain.JobRunning, got.State)
	}
	if got.Attempts != 2 {
		t.Fatalf("expected attempts incremented to 2, got %d", got.Attempts)
	}
}
