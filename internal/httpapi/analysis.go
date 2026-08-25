package httpapi

import (
	"net/http"
	"strings"

	"t117/internal/domain"
)

func (
	a *App,
) analysisRoute(
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
	case "analysis":
		value, err :=
			a.matches.Analyze(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case "suggestions":
		value, err :=
			a.matches.NextSteps(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case "scorecard":
		value, err :=
			a.matches.Scorecard(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case "catalog":
		value, err :=
			a.matches.Catalog(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case "insights":
		value, err :=
			a.matches.Insights(
				user.ID, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case "reflections":
		a.reflectionRoute(w, r, user.ID, matchID)
	default:
		writeError(
			w, domain.ErrMissing)
	}
}

func (
	a *App,
) compareRoute(
	w http.ResponseWriter, r *http.Request,
) {
	user, _ := currentUser(r)
	values := r.URL.Query()["match_id"]
	if len(values) == 0 {
		values = strings.Split(r.URL.Query().Get("match_ids"), ",")
	}
	ids :=
		make(
			[]domain.ID,
			0,
			len(values),
		)
	for valueIndex := range values {
		value := values[valueIndex]
		trimmedValue := strings.TrimSpace(value)
		if value != "" {
			ids = append(ids, domain.ID(trimmedValue))
		}
	}
	result, err :=
		a.matches.Compare(
			user.ID, ids)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, result)
}

func (a *App) reflectionRoute(
	w http.ResponseWriter,
	r *http.Request,
	owner, matchID domain.ID,
) {
	if a.reflect == nil {
		writeError(
			w, domain.ErrConflict)
		return
	}
	switch r.Method {
	case http.MethodGet:
		value, err :=
			a.reflect.List(
				owner, matchID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, value)
	case http.MethodPost:
		var input domain.ReflectionInput
		if err :=
			decode(
				r,
				&input,
			); err != nil {
			writeError(w, err)
			return
		}
		value, err := a.reflect.Save(owner, matchID, input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeCreated(w, value)
	default:
		writeError(
			w, domain.ErrInvalid)
	}
}
