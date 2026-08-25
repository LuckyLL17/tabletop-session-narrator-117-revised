package verification

// Coverage source markers: process, Complete, SaveJob

import (
	"errors"
	"path/filepath"
	"testing"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/store"
)

func TestBug014WorkerCompletesFailureAsRetry(t *testing.T) {
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	job := domain.Job{ID: "j1", OwnerID: "o", MatchID: "m", State: domain.JobRetry, Attempts: 1}
	if err := state.SaveJob(job); err != nil {
		t.Fatal(err)
	}
	if err := service.NewJobService(state).Complete(job, errors.New("报告失败")); err != nil {
		t.Fatal(err)
	}
	jobs := state.JobsForOwner(job.OwnerID)
	if len(jobs) != 1 {
		t.Fatalf("任务丢失: %#v", jobs)
	}
	got := jobs[0]
	if got.State != domain.JobRetry || got.ScheduledAt.IsZero() {
		t.Fatalf("失败任务应保留重试状态: %#v", got)
	}
}

func TestBug014RegressionHealth(t *testing.T) {
	if got := domain.EventKinds(); len(got) < 2 {
		t.Fatal("事件类型过少")
	}
}
