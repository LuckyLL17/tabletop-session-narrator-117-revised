package httpapi

import (
	"net/http"
	"strings"
)

func (
	a *App,
) searchRoute(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, _ := currentUser(r)
	writeOK(w, a.search.Search(user.ID, strings.TrimSpace(r.URL.Query().Get("q"))))
}
