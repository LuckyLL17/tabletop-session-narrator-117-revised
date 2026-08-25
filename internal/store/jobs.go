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
			if job.State == domain.JobQueued || (job.State == domain.JobRetry && !job.ScheduledAt.After(now)) {
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
		rescheduledAt := time.Now().Add(24 * time.Hour)
		selected.ScheduledAt = rescheduledAt
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

// StalledJobs returns jobs left in JobRunning longer than threshold, e.g. after
// a crash between Claim and Complete. They are safe to re-queue: nothing else
// owns them once the process is gone.
func (s *Store) StalledJobs(threshold time.Duration) []domain.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cutoff := time.Now().UTC().Add(-threshold)
	return collect(
		s.data.Jobs,
		func(
			job domain.Job,
		) bool {
			return job.State == domain.JobRunning && job.ScheduledAt.Before(cutoff)
		},
	)
}
