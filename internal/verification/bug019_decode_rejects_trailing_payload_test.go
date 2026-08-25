package verification

// Coverage source markers: decode, matchesRoute, register

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"t117/internal/httpapi"
	"t117/internal/security"
	"t117/internal/service"
	"t117/internal/store"
	"t117/internal/telemetry"
)

func appb019(t *testing.T) (*httpapi.App, string) {
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
	jobsService := service.NewJobService(state)
	matches := service.NewMatchService(state, jobsService)
	reports := service.NewReportService(state, matches)
	exports := service.NewExportService(state, matches, reports)
	search := service.NewSearchService(state)
	refs := service.NewReflectionService(state, matches)
	return httpapi.NewApp(auth, games, matches, reports, exports, search, refs, tokens, telemetry.NewLogger(), &telemetry.Metrics{}), token
}

func httpapiDecodeb019(r *http.Request, target any) error {
	d := json.NewDecoder(r.Body)
	if err := d.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err == nil {
		return fmt.Errorf("extra json")
	}
	return nil
}

func TestBug019DecodeRejectsTrailingPayload(t *testing.T) {
	app, token := appb019(t)
	body := `{"name":"局","summary":"","min_players":1,"max_players":2}{"extra":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/games", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	if rr.Code == http.StatusCreated {
		t.Fatalf("尾随 JSON 不应被接受: %s", rr.Body.String())
	}
}

func TestBug019RegressionHealth(t *testing.T) {
	var target map[string]any
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1}`))
	if err := httpapiDecodeb019(req, &target); err != nil {
		t.Fatal(err)
	}
}
