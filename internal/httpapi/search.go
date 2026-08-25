package httpapi

import "net/http"

func (
	a *App,
) searchRoute(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, _ := currentUser(r)
	writeOK(w, a.search.Search(user.ID, r.URL.Query().Get("q")))
}
