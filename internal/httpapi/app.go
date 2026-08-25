package httpapi

import (
	"net/http"
	"time"

	"t117/internal/security"
	"t117/internal/service"
	"t117/internal/telemetry"
)

type App struct {
	auth    *service.AuthService
	games   *service.GameService
	matches *service.MatchService
	reports *service.ReportService
	exports *service.ExportService
	search  *service.SearchService
	reflect *service.ReflectionService
	jobs    *service.JobService
	tokens  security.TokenCodec
	logger  *telemetry.Logger
	metrics *telemetry.Metrics
}

func NewApp(auth *service.AuthService, games *service.GameService, matches *service.MatchService, reports *service.ReportService, exports *service.ExportService, search *service.SearchService, reflect *service.ReflectionService, jobs *service.JobService, tokens security.TokenCodec, logger *telemetry.Logger, metrics *telemetry.Metrics) *App {
	return &App{auth: auth, games: games, matches: matches, reports: reports, exports: exports, search: search, reflect: reflect, jobs: jobs, tokens: tokens, logger: logger, metrics: metrics}
}
func (a *App) Handler() http.Handler { return a.metrics.Middleware(a.logMiddleware(a.routes())) }
func (
	a *App,
) logMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		a.logger.Request(r.Method, r.URL.Path, started)
	})
}
