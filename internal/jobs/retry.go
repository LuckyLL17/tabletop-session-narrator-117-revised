package jobs

import (
	"time"

	"t117/internal/service"
)

// 退避参数。战报任务失败后按指数退避重新调度，超出 MaxRetries 后转为失败，
// 不再被重新领取，避免高频重试同一任务。
const (
	// baseDelay 是第一次重试前的最小等待时间。
	baseDelay = time.Second
	// maxDelay 是任意一次重试等待时间的上限，防止高次数下持续时间溢出
	// 或退避到不合理的长度。
	maxDelay = 5 * time.Minute
	// MaxRetries 是放弃重试前的最大重试次数（不含首次执行）。
	// 达到该次数后任务转为 JobFailed，稳定停在上限，不再被重新领取。
	MaxRetries = 7
)

// RetryWindow 返回第 attempts 次重试应等待的时长。
// attempts 为已经发生的重试次数（>=1 表示第一次重试）。
// 返回值始终为正，且在 [baseDelay, maxDelay] 区间内。
// 重试次数超过 MaxRetries 时调用方应改用 Failed 终态，不再调用本函数，
// 但为防御性编程，这里仍返回 maxDelay 而非溢出或负值。
func RetryWindow(attempts int) time.Duration {
	if attempts < 1 {
		return baseDelay
	}
	shift := attempts
	// 防止位移溢出：time.Duration 是 int64 纳秒，
	// shift 过大时 1<<shift 会溢出成 0 或负数。先按位移上限钳制。
	if shift > 30 {
		return maxDelay
	}
	backoff := baseDelay << uint(shift)
	// 溢出/负值兜底：超出 maxDelay 时一律取 maxDelay。
	if backoff <= 0 || backoff > maxDelay {
		return maxDelay
	}
	return backoff
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
