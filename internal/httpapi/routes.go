package httpapi

import "net/http"

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.frontend)
	mux.HandleFunc("/api/v1/health", a.health)
	mux.HandleFunc("/api/v1/auth/register", a.register)
	mux.HandleFunc("/api/v1/auth/login", a.login)
	mux.HandleFunc("/api/v1/session", a.session)
	mux.Handle(
		"/api/v1/games", a.protected(http.HandlerFunc(a.gamesRoute)))
	mux.Handle("/api/v1/games/", a.protected(http.HandlerFunc(a.gamesRoute)))
	mux.Handle(
		"/api/v1/matches", a.protected(http.HandlerFunc(a.matchesRoute)))
	mux.Handle("/api/v1/matches/", a.protected(http.HandlerFunc(a.matchesRoute)))
	mux.Handle(
		"/api/v1/compare", a.protected(http.HandlerFunc(a.compareRoute)))
	mux.Handle(
		"/api/v1/search", a.protected(http.HandlerFunc(a.searchRoute)))
	mux.Handle("/api/v1/export/markdown", a.protected(http.HandlerFunc(a.exportMarkdown)))
	mux.Handle("/api/v1/export/json", a.protected(http.HandlerFunc(a.exportJSON)))
	return mux
}
