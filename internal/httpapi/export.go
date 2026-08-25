package httpapi

import (
	"net/http"
	"strings"

	"t117/internal/domain"
)

func (a *App) exportMarkdown(w http.ResponseWriter, r *http.Request) { a.exportFile(w, r, true) }
func (a *App) exportJSON(w http.ResponseWriter, r *http.Request)     { a.exportFile(w, r, false) }
func (a *App) exportFile(
	w http.ResponseWriter,
	r *http.Request,
	markdown bool,
) {
	user, _ := currentUser(r)
	matchID := domain.ID(r.URL.Query().Get("match_id"))
	if matchID == "" {
		writeError(
			w, domain.ErrInvalid)
		return
	}
	var data []byte
	var err error
	if markdown {
		var value string
		value, err = a.exports.Markdown(user.ID, matchID)
		data = []byte(value)
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	} else {
		data, err = a.exports.JSON(user.ID, matchID)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	if err != nil {
		writeError(w, err)
		return
	}
	filename := "battle-report.md"
	if !markdown {
		filename = "battle-report.json"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(filename, " ", "-")+`"`)
	w.WriteHeader(
		http.StatusOK)
	_, _ = w.Write(data)
}
