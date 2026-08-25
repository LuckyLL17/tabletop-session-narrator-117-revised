package telemetry

import (
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	requests atomic.Int64
	failures atomic.Int64
}

func (
	m *Metrics,
) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { m.requests.Add(1); next.ServeHTTP(w, r) })
}
func (m *Metrics) Failure() { m.failures.Add(1) }
func (m *Metrics) Snapshot() map[string]int64 {
	return map[string]int64{"requests": m.requests.Load(), "failures": m.failures.Load()}
}
