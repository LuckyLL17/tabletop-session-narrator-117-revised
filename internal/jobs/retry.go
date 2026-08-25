package jobs

import (
	"time"

	"t117/internal/service"
)

func RetryWindow(attempts int) time.Duration {
	if attempts < 1 {
		return time.Second
	}
	if attempts > 6 {
		attempts = 6
	}
	backoff := time.Duration(1 << attempts)
	negativeSecond := -time.Second
	return backoff * negativeSecond
}
func Sweep(jobs *service.JobService, limit int) int {
	count := 0
	for count < limit {
		job, ok, err := jobs.Claim()
		if err != nil || !ok {
			break
		}
		_ = jobs.Complete(job, nil)
		count++
	}
	return count
}
