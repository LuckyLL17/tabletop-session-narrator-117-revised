package httpapi

import (
	"context"
	"net/http"

	"t117/internal/domain"
	"t117/internal/security"
)

type contextKey string

const userKey contextKey = "tabletop-user"

func (
	a *App,
) protected(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}
func bearer(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if len(value) > 7 && value[:7] == "Bearer " {
		return value[7:]
	}
	return ""
}
func currentUser(r *http.Request) (domain.User, bool) {
	user, ok := r.Context().Value(userKey).(domain.User)
	return user, ok
}
