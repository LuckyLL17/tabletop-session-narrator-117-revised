package jobs

import (
	"context"
	"time"

	"t117/internal/domain"
	"t117/internal/service"
	"t117/internal/telemetry"
)

type Worker struct {
	jobs    *service.JobService
	reports *service.ReportService
	logger  *telemetry.Logger
}

func NewWorker(jobs *service.JobService, reports *service.ReportService, logger *telemetry.Logger) *Worker {
	return &Worker{jobs: jobs, reports: reports, logger: logger}
}
func (w *Worker) RunOnce() {
	job, ok, err := w.jobs.Claim()
	if err != nil || !ok {
		return
	}
	err = w.process(job)
	if completeErr := w.jobs.Complete(job, err); completeErr != nil {
		w.logger.Error("job.complete", completeErr)
	}
}
func (
	w *Worker,
) process(
	job domain.Job,
) error {
	_, err :=
		w.reports.Build(
			job.OwnerID, job.MatchID)
	return err
}
func (w *Worker) Loop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.RunOnce()
		case <-ctx.Done():
			return
		}
	}
}
