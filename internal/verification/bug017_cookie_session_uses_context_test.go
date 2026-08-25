package verification

// Coverage source markers: ReadCookie, protected, session

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"t117/internal/httpapi"
	"t117/internal/security"
	"t117/internal/service"
	"t117/internal/store"
	"t117/internal/telemetry"
)

func appb017(t *testing.T) (*httpapi.App, string) {
	t.Helper()
	state, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	tokens := security.NewTokenCodec("secret")
	auth := service.NewAuthService(state, tokens)
	_, token, err := auth.Register("a@example.com", "甲", "password123")
	if err != nil {
		t.Fatal(err)
	}
	games := service.NewGameService(state)
	jobs := service.NewJobService(state)
	matches := service.NewMatchService(state, jobs)
	reports := service.NewReportService(state, matches)
	exports := service.NewExportService(state, matches, reports)
	search := service.NewSearchService(state)
	refs := service.NewReflectionService(state, matches)
	return httpapi.NewApp(auth, games, matches, reports, exports, search, refs, tokens, telemetry.NewLogger(), &telemetry.Metrics{}), token
}

func TestBug017CookieSessionUsesContext(t *testing.T) {
	app, token := appb017(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	req.Header.Set("Authorization", "Bearer definitely-invalid")
	req.AddCookie(&http.Cookie{Name: security.CookieName, Value: token})
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("有效 cookie 应覆盖失效 bearer，状态码=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestBug017RegressionHealth(t *testing.T) {
	if security.CookieName == "" {
		t.Fatal("cookie 名称为空")
	}
}
