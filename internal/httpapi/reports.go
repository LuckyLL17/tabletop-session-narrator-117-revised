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
	if len(parts) != 5 {
		writeError(
			w, domain.ErrMissing)
		return
	}
	report, err := a.reports.Get(user.ID, domain.ID(parts[3]))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, report)
}
