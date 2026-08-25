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
func (s *JobService) Claim() (domain.Job, bool, error) { return s.store.ClaimJob(time.Now().UTC()) }
func (s *JobService) Complete(job domain.Job, err error) error {
	now := time.Now().UTC()
	if err != nil {
		job.LastError = err.Error()
		// 重试次数已经达到上限：稳定停在终态，不再被 ClaimJob 重新领取。
		if job.Attempts >= maxRetries {
			job.State = domain.JobFailed
			job.CompletedAt = &now
			return s.store.SaveJob(job)
		}
		// 计算下一次重试的等待时长。必须为正值，否则 ScheduledAt 会落到
		// now 之前，ClaimJob 的 "!ScheduledAt.After(now)" 条件恒为真，
		// 调度器会高频重新领取同一任务。
		job.State = domain.JobRetry
		job.ScheduledAt = now.Add(retryDelay(job.Attempts))
	} else {
		job.State = domain.JobDone
		job.CompletedAt = &now
	}
	return s.store.SaveJob(job)
}

// maxRetries 是放弃重试前允许的最大执行次数（含首次执行）。
// 超过后任务转为 JobFailed 终态，不再被重新领取。
const maxRetries = 8

// retryDelay 返回第 attempts 次失败后应等待的重试时长。
// attempts 为已经发生的执行次数（>=1）。返回值始终为正，
// 且在 [1s, 5m] 区间内；高次数下也不会溢出。
func retryDelay(attempts int) time.Duration {
	const (
		baseDelay = time.Second
		maxDelay  = 5 * time.Minute
	)
	if attempts < 1 {
		return baseDelay
	}
	shift := attempts
	// time.Duration 为 int64 纳秒；位移过大时 1<<shift 会溢出为 0 或负数。
	// 先把位移钳制在安全范围内，再做溢出兜底。
	if shift > 30 {
		return maxDelay
	}
	backoff := baseDelay << uint(shift)
	if backoff <= 0 || backoff > maxDelay {
		return maxDelay
	}
	return backoff
}
