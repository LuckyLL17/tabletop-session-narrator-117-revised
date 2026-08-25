package jobs

import (
	"time"

	"t117/internal/domain"
	"t117/internal/service"
)

// RetryWindow mirrors service.retryWindow: the backoff before a failed job is
// re-claimable. Kept here as the public policy helper referenced by ops tooling.
func RetryWindow(attempts int) time.Duration {
	if attempts < 1 {
		return time.Second
	}
	if attempts > 6 {
		attempts = 6
	}
	return time.Duration(1<<attempts) * time.Second
}

// stalledThreshold is how long a JobRunning job may be owned before it is
// considered abandoned (process crash between Claim and Complete) and reset.
const stalledThreshold = 5 * time.Minute

// Sweep re-queues jobs abandoned in JobRunning (e.g. after a crash between Claim
// and Complete) so the worker gets another attempt. It deliberately does NOT
// touch JobQueued/JobRetry — those are reclaimed by Claim on their own schedule
// — and never marks a job done without running it.
func Sweep(jobs *service.JobService, limit int) int {
	stalled := jobs.StalledJobs(stalledThreshold)
	if len(stalled) > limit {
		stalled = stalled[:limit]
	}
	count := 0
	for index := range stalled {
		job := stalled[index]
		job.State = domain.JobQueued
		job.ScheduledAt = time.Now().UTC()
		if jobs.ResetJob(job) == nil {
			count++
		}
	}
	return count
}
