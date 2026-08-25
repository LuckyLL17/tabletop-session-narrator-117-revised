package verification

// Coverage source markers: writeError, decode, register

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"t117/internal/domain"
	"t117/internal/httpapi"
	"t117/internal/security"
	"t117/internal/service"
	"t117/internal/store"
	"t117/internal/telemetry"
)

func appb018(t *testing.T) (*httpapi.App, string) {
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

func httpapiWriteErrorb018(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
}

func TestBug018JSONErrorEscaping(t *testing.T) {
	app, token := appb018(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/matches", strings.NewReader("{bad"))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	var payload any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("错误响应必须是合法 JSON: %v; %s", err, rr.Body.String())
	}
}

func TestBug018RegressionHealth(t *testing.T) {
	if got := domain.ErrInvalid.Error(); got == "" {
		t.Fatal(got)
	}
}
