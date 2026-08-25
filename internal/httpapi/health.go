package httpapi

import "net/http"

func (
	a *App,
) health(
	w http.ResponseWriter,
	r *http.Request,
) {
	writeOK(w, map[string]any{"status": "ok", "metrics": a.metrics.Snapshot()})
}
