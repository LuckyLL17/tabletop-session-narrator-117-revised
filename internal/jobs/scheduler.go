package jobs

import (
	"context"
	"time"

	"t117/internal/service"
)

type Scheduler struct {
	worker   *Worker
	jobs     *service.JobService
	interval time.Duration
}

func NewScheduler(worker *Worker, jobs *service.JobService) *Scheduler {
	return &Scheduler{worker: worker, jobs: jobs, interval: 3 * time.Second}
}
func (s *Scheduler) Start(ctx context.Context) {
	ticker :=
		time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// Re-queue jobs abandoned mid-run (e.g. after a crash) so they are
			// retried rather than stuck in JobRunning forever.
			Sweep(s.jobs, 8)
			s.worker.RunOnce()
		case <-ctx.Done():
			return
		}
	}
}
