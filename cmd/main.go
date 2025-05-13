package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	bl "github.com/SlashLight/golang-balancer/internal/balancer"
	"github.com/SlashLight/golang-balancer/internal/config"
	health_check "github.com/SlashLight/golang-balancer/internal/health-check"
	"github.com/SlashLight/golang-balancer/internal/logger"
	"github.com/SlashLight/golang-balancer/internal/middleware"
	"github.com/SlashLight/golang-balancer/internal/rate-limiter/controller"
	"github.com/SlashLight/golang-balancer/internal/rate-limiter/storage"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

//TODO: [] add graceful shutdown
//TODO: [] add Readme

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting golang balancer",
		slog.String("env", cfg.Env),
	)

	balancer, err := bl.NewBalancer(cfg.Balancer.Algorithm, cfg.Balancer.Backends)
	if err != nil {
		log.Error("failed to init balancer", logger.Err(err))
		os.Exit(1)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	maxRetries := cfg.Retries
	limiter := storage.NewRedisRateLimiter(cfg)
	if err = limiter.Client.Ping(context.Background()).Err(); err != nil {
		log.Error("Error connecting to Redis", logger.Err(err))
		os.Exit(1)
	}

	chain := middleware.RateLimitMiddleware(limiter, log)(
		middleware.AccessLog(log)(
			middleware.RetryMiddleware(balancer, log, maxRetries)(
				handler,
			),
		))
	clientHandler := middleware.AccessLog(log)(controller.NewRateLimitController(limiter, log))

	mux := http.NewServeMux()
	mux.Handle("/", chain)
	mux.Handle("/clients", clientHandler)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	checker, err := health_check.NewHealthChecker(cfg.HealthChecker.Interval,
		cfg.Backends,
		cfg.HealthChecker.CheckURL,
		log)
	if err != nil {
		log.Error("failed to init health checker", logger.Err(err))
		os.Exit(1)
	}

	go checker.Start(balancer)

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", logger.Err(err))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	//TODO: [] настроить вывод логгера
	switch env {
	case envLocal:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
