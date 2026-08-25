package service

import (
	"sort"
	"time"

	"t117/internal/domain"
	"t117/internal/store"
	"t117/pkg/ids"
)

type JobService struct{ store *store.Store }

// maxAttempts caps automatic retries. Beyond it a failing job stays in the
// JobFailed state so the failure stays visible instead of being masked as done.
const maxAttempts = 5

func NewJobService(
	data *store.Store,
) *JobService {
	return &JobService{store: data}
}
func (s *JobService) Enqueue(owner, matchID domain.ID, kind string) error {
	job := domain.Job{ID: domain.ID(ids.New("job")), OwnerID: owner, MatchID: matchID, Kind: kind, State: domain.JobQueued, ScheduledAt: time.Now().UTC()}
	return s.store.SaveJob(job)
}
func (s *JobService) Claim() (domain.Job, bool, error) { return s.store.ClaimJob(time.Now().UTC()) }

// Complete advances a job to its terminal-or-retry state. The error from the
// worker MUST be forwarded here: a non-nil err records LastError and schedules a
// retry (or marks the job failed once attempts are exhausted), instead of
// erasing the failure as a successful JobDone. Only a nil err completes the job.
func (s *JobService) Complete(job domain.Job, err error) error {
	now := time.Now().UTC()
	if err != nil {
		job.LastError = err.Error()
		if job.Attempts < maxAttempts {
			job.State = domain.JobRetry
			job.ScheduledAt = now.Add(retryWindow(job.Attempts))
		} else {
			job.State = domain.JobFailed
		}
		return s.store.SaveJob(job)
	}
	job.State = domain.JobDone
	job.LastError = ""
	job.CompletedAt = &now
	return s.store.SaveJob(job)
}

// retryWindow returns the backoff before the next claim attempt. It mirrors the
// internal/jobs.RetryWindow policy so the service layer stays free of that
// dependency (internal/jobs imports this package).
func retryWindow(attempts int) time.Duration {
	if attempts < 1 {
		return time.Second
	}
	exp := attempts
	if exp > 6 {
		exp = 6
	}
	return time.Duration(1<<exp) * time.Second
}

// JobsForMatch returns the jobs for a match, newest first, so the API can surface
// the latest generation state (including failures) to the user.
func (s *JobService) JobsForMatch(owner, matchID domain.ID) []domain.Job {
	rows := s.store.JobsForOwner(owner)
	result := []domain.Job{}
	for index := range rows {
		job := rows[index]
		if job.MatchID == matchID {
			result = append(result, job)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ScheduledAt.After(result[j].ScheduledAt)
	})
	return result
}

// StalledJobs exposes abandoned JobRunning jobs for the sweep to re-queue.
func (s *JobService) StalledJobs(threshold time.Duration) []domain.Job {
	return s.store.StalledJobs(threshold)
}

// ResetJob writes a job back to the store as-is (used by the sweep to re-queue
// stalled jobs).
func (s *JobService) ResetJob(job domain.Job) error {
	return s.store.SaveJob(job)
}

// RetryFailed resets a failed or stale retry job so the worker can pick it up
// again. It preserves LastError for diagnosis; the next Claim bumps Attempts
// so backoff history is not lost. Returns ErrMissing when no job exists, or
// ErrConflict when the latest job is not in a retryable state.
func (s *JobService) RetryFailed(owner, matchID domain.ID) (domain.Job, error) {
	jobs := s.JobsForMatch(owner, matchID)
	if len(jobs) == 0 {
		return domain.Job{}, domain.ErrMissing
	}
	job := jobs[0]
	if job.State != domain.JobFailed && job.State != domain.JobRetry {
		return domain.Job{}, domain.ErrConflict
	}
	now := time.Now().UTC()
	job.State = domain.JobQueued
	job.ScheduledAt = now
	return job, s.store.SaveJob(job)
}
