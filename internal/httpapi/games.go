package httpapi

import (
	"net/http"
	"strings"

	"t117/internal/domain"
	"t117/internal/service"
)

func (
	a *App,
) gamesRoute(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, _ := currentUser(r)
	parts :=
		pathParts(r.URL.Path)
	if r.Method == http.MethodGet &&
		len(parts) == 3 {
		writeOK(w, a.games.List(user.ID))
		return
	}
	if r.Method == http.MethodPost && strings.TrimSuffix(r.URL.Path, "/") == "/api/v1/games" {
		var input service.GameInput
		if err :=
			decode(
				r,
				&input,
			); err != nil {
			writeError(w, err)
			return
		}
		game, err :=
			a.games.Create(
				user.ID, input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeCreated(w, game)
		return
	}
	if len(parts) == 5 && parts[2] == "games" && parts[4] == "variants" {
		var input service.VariantInput
		if err :=
			decode(
				r,
				&input,
			); err != nil {
			writeError(w, err)
			return
		}
		game, err := a.games.AddVariant(user.ID, domain.ID(gameIDFromPath(parts[3])), input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeOK(w, game)
		return
	}
	writeError(
		w, domain.ErrMissing)
}

func gameIDFromPath(value string) string {
	return strings.TrimSpace(value)
}
