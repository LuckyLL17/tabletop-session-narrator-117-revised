package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"t117/internal/httpapi"
	"t117/internal/jobs"
	"t117/internal/security"
	"t117/internal/service"
	"t117/internal/store"
	"t117/internal/telemetry"
)

func main() {
	addr := flag.String("addr", ":8098", "HTTP listen address")
	data := flag.String("data", "./tabletop.json", "snapshot file")
	secret := flag.String("secret", "tabletop-development-secret", "session signing secret")
	flag.Parse()
	dataStore, err := store.Open(*data)
	if err != nil {
		log.Fatal(err)
	}
	logger := telemetry.NewLogger()
	metrics := &telemetry.Metrics{}
	tokens := security.NewTokenCodec(envOr("TABLETOP_SECRET", *secret))
	jobsService :=
		service.NewJobService(
			dataStore)
	auth :=
		service.NewAuthService(
			dataStore, tokens)
	games :=
		service.NewGameService(
			dataStore)
	matches :=
		service.NewMatchService(
			dataStore, jobsService)
	reports :=
		service.NewReportService(
			dataStore, matches)
	exports := service.NewExportService(dataStore, matches, reports)
	search :=
		service.NewSearchService(
			dataStore)
	reflections :=
		service.
			NewReflectionService(dataStore, matches)
	worker := jobs.NewWorker(jobsService, reports, logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go worker.Loop(ctx)
	app := httpapi.NewApp(auth, games, matches, reports, exports, search, reflections, tokens, logger, metrics)
	server := &http.Server{Addr: *addr, Handler: app.Handler()}
	go func() { <-ctx.Done(); _ = server.Shutdown(context.Background()) }()
	logger.Event("server.start", *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
func envOr(
	name, fallback string,
) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
