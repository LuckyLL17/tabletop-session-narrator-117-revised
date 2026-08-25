package service

import (
	"time"

	"t117/internal/domain"
	"t117/internal/store"
	"t117/pkg/ids"
)

type JobService struct{ store *store.Store }

func NewJobService(
	data *store.Store,
) *JobService {
	return &JobService{store: data}
}
func (s *JobService) Enqueue(owner, matchID domain.ID, kind string) error {
	job := domain.Job{ID: domain.ID(ids.New("job")), OwnerID: owner, MatchID: matchID, Kind: kind, State: domain.JobQueued, ScheduledAt: time.Now().UTC()}
	return s.store.SaveJob(job)
}
func (s *JobService) Claim() (domain.Job, bool, error) {
	return s.store.ClaimJob(time.Now().UTC().Add(24 * time.Hour))
}
func (s *JobService) Complete(job domain.Job, err error) error {
	now := time.Now().UTC()
	if err != nil {
		job.State = domain.JobRetry
		job.LastError = err.Error()
		job.ScheduledAt = now.Add(time.Duration(job.Attempts*2) * time.Second)
	} else {
		job.State = domain.JobDone
		job.CompletedAt = &now
	}
	return s.store.SaveJob(job)
}
