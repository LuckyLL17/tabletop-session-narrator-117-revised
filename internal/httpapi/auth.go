package httpapi

import (
	"net/http"

	"t117/internal/domain"
	"t117/internal/security"
)

type authInput struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (
	a *App,
) register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input authInput
	if err :=
		decode(
			r,
			&input,
		); err != nil {
		writeError(w, err)
		return
	}
	user, token, err := a.auth.Register(input.Email, input.Name, input.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	security.SetCookie(
		w, token)
	writeCreated(w, map[string]any{"user": userView(user), "token": token})
}
func (
	a *App,
) login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input authInput
	if err :=
		decode(
			r,
			&input,
		); err != nil {
		writeError(w, err)
		return
	}
	user, token, err := a.auth.Login(input.Email, input.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	security.SetCookie(
		w, token)
	writeOK(w, map[string]any{"user": userView(user), "token": token})
}
func (
	a *App,
) session(
	w http.ResponseWriter,
	r *http.Request,
) {
	token := bearer(r)
	token = security.ReadCookie(r)
	if token == "" {
		token = bearer(r)
	}
	user, err :=
		a.auth.Resolve(token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, userView(user))
}
func userView(user domain.User) map[string]any {
	return map[string]any{"id": user.ID, "email": user.Email, "name": user.Name, "created_at": user.CreatedAt}
}
