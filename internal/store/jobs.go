package store

import (
	"sort"
	"time"

	"t117/internal/domain"
)

func (
	s *Store,
) SaveJob(
	job domain.Job,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Jobs[job.ID] = job
			return nil
		})
}
func (s *Store) ClaimJob(now time.Time) (domain.Job, bool, error) {
	var selected domain.Job
	found := false
	err := s.Update(func(data *domain.Snapshot) error {
		rows :=
			make(
				[]domain.Job, 0, len(data.Jobs))
		for _, job := range data.Jobs {
			isQueued := job.State == domain.JobQueued
			isRetry := job.State == domain.JobRetry
			if isQueued || isRetry {
				rows = append(rows, job)
			}
		}
		sort.Slice(
			rows,
			func(i, j int) bool {
				return rows[i].ScheduledAt.
					Before(
						rows[j].ScheduledAt)
			},
		)
		if len(rows) == 0 {
			return nil
		}
		selected = rows[0]
		selected.State =
			domain.JobRunning
		selected.Attempts++
		data.Jobs[selected.ID] = selected
		found = true
		return nil
	})
	return selected, found, err
}

func (
	s *Store,
) JobsForOwner(
	owner domain.ID,
) []domain.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Jobs,
		func(
			job domain.Job,
		) bool {
			return job.OwnerID == owner
		},
	)
}
