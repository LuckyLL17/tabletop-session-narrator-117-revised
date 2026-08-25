package httpapi

import (
	"net/http"

	"t117/internal/domain"
)

func (
	a *App,
) reportRoute(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, _ := currentUser(r)
	parts :=
		pathParts(r.URL.Path)
	if len(parts) < 5 {
		writeError(
			w, domain.ErrMissing)
		return
	}
	matchID := domain.ID(parts[3])
	// POST .../matches/{id}/report/retry re-queues a failed/stale report job so
	// the user can retry generation; the failure state and LastError are
	// preserved up to the retry.
	if parts[4] == "report" && len(parts) > 5 && parts[5] == "retry" && r.Method == http.MethodPost {
		job, err := a.jobs.RetryFailed(user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, job)
		return
	}
	if len(parts) != 5 {
		writeError(
			w, domain.ErrMissing)
		return
	}
	report, err := a.reports.Get(user.ID, matchID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, report)
}
