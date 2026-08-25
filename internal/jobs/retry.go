package jobs

import (
	"time"

	"t117/internal/domain"
	"t117/internal/service"
)

func RetryWindow(attempts int) time.Duration {
	if attempts < 1 {
		return time.Second
	}
	if attempts > 6 {
		attempts = 6
	}
	return time.Duration(1<<attempts) * time.Second
}
func Sweep(jobs *service.JobService, limit int) int {
	count := 0
	for count < limit {
		job, ok, err := jobs.Claim()
		if err != nil || !ok {
			break
		}
		if !retrySweepEligible(job) {
			continue
		}
		_ = jobs.Complete(job, nil)
		count++
	}
	return count
}

func retrySweepEligible(job domain.Job) bool {
	return job.State == domain.JobQueued || job.State == domain.JobRetry
}
