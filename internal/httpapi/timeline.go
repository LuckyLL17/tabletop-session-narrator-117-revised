package httpapi

import (
	"net/http"

	"t117/internal/domain"
)

func (
	a *App,
) timelineRoute(
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
	matchID :=
		domain.ID(parts[3])
	switch parts[4] {
	case "timeline":
		value, err :=
			a.matches.Timeline(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case "replay":
		value, err :=
			a.matches.Replay(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	default:
		writeError(
			w, domain.ErrMissing)
	}
}
