package httpapi

import (
	"io/fs"
	"net/http"
	"strings"

	frontend "t117/web"
)

func (
	a *App,
) frontend(
	w http.ResponseWriter,
	r *http.Request,
) {
	files, _ := fs.Sub(frontend.Files, ".")
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	if path == "index.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	http.FileServer(http.FS(files)).ServeHTTP(w, r.Clone(r.Context()))
}
