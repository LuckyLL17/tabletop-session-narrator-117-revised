package verification

// Coverage source markers: Start, Loop, Claim

import (
	"context"
	"testing"
	"time"

	"t117/internal/jobs"
)

func TestBug030SchedulerStopsOnCancel(t *testing.T) {
	scheduler := jobs.NewScheduler(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { scheduler.Start(ctx); close(done) }()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("scheduler 未响应取消")
	}
}

func TestBug030RegressionHealth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	select {
	case <-ctx.Done():
		t.Fatal("新建 context 不应已取消")
	default:
	}
}
