package jobs

import (
	"context"
	"time"
)

type Scheduler struct {
	worker   *Worker
	interval time.Duration
}

func NewScheduler(worker *Worker) *Scheduler {
	return &Scheduler{worker: worker, interval: 3 * time.Second}
}
func (s *Scheduler) Start(ctx context.Context) {
	ticker :=
		time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.worker.RunOnce()
		case <-ctx.Done():
			return
		}
	}
}
