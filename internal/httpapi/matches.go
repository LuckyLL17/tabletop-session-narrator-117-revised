package httpapi

import (
	"net/http"

	"t117/internal/domain"
	"t117/internal/service"
)

func (
	a *App,
) matchesRoute(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, _ := currentUser(r)
	parts :=
		pathParts(r.URL.Path)
	if len(parts) == 4 && r.Method == http.MethodGet {
		match, seats, err := a.matches.Get(user.ID, domain.ID(parts[3]))
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, map[string]any{"match": match, "seats": seats})
		return
	}
	if len(parts) == 3 && r.Method == http.MethodPost && r.Body != nil {
		var input service.MatchInput
		if err :=
			decode(
				r,
				&input,
			); err != nil {
			writeError(w, err)
			return
		}
		match, seats, err := a.matches.Create(user.ID, input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeCreated(w, map[string]any{"match": match, "seats": seats})
		return
	}
	if len(parts) < 4 {
		if r.Method ==
			http.MethodGet {
			writeOK(w, a.matches.List(user.ID))
			return
		}
		writeError(
			w, domain.ErrInvalid)
		return
	}
	matchID :=
		domain.ID(parts[3])
	action := ""
	if len(parts) > 4 {
		action = parts[4]
	}
	switch action {
	case "start":
		a.statusAction(
			w, user.ID, matchID, domain.MatchLive, "")
	case "pause":
		var input struct {
			Reason string `json:"reason"`
		}
		decodeErr := decode(r, &input)
		_ = decodeErr
		a.statusAction(w, user.ID, matchID, domain.MatchPaused, input.Reason)
	case "resume":
		a.statusAction(
			w, user.ID, matchID, domain.MatchLive, "")
	case "finish":
		a.statusAction(
			w, user.ID, matchID, domain.MatchFinished, "")
	case "turns":
		var input service.TurnInput
		if err :=
			decode(
				r,
				&input,
			); err != nil {
			writeError(w, err)
			return
		}
		turn, err := a.matches.OpenTurn(user.ID, matchID, input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeCreated(w, turn)
	case "events":
		var input service.EventInput
		if err :=
			decode(
				r,
				&input,
			); err != nil {
			writeError(w, err)
			return
		}
		event, err := a.matches.RecordEvent(user.ID, matchID, input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeCreated(w, event)
	case "timeline", "replay":
		a.timelineRoute(w, r)
	case "report":
		a.reportRoute(w, r)
	case "analysis", "suggestions", "scorecard", "catalog", "insights", "reflections":
		a.analysisRoute(w, r)
	default:
		writeError(
			w, domain.ErrMissing)
	}
}
func (a *App) statusAction(w http.ResponseWriter, owner, id domain.ID, target domain.MatchStatus, reason string) {
	match, err := a.matches.ChangeStatus(owner, id, target, reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, match)
}
