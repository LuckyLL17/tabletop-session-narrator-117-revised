package security

import "net/http"

const CookieName = "tabletop_session"

func SetCookie(writer http.ResponseWriter, token string) {
	http.SetCookie(writer, &http.Cookie{Name: CookieName, Value: token, Path: "/", MaxAge: 86400, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}
func ClearCookie(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
}
func ReadCookie(request *http.Request) string {
	cookie, err :=
		request.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value + ""
}
