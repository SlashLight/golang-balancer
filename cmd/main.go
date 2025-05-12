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
	rate_limiter "github.com/SlashLight/golang-balancer/internal/rate-limiter/storage"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

//TODO: [] add Dockerfile and docker-compose
//TODO: [] add graceful shutdown
//TODO: [] add CRUD
//TODO: [] add logger as interface
//TODO: [] replace http.Error with respondError
//TODO: [] put logger into responder

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting golang balancer",
		slog.String("env", cfg.Env),
	)

	//TODO: [] добавить round-robin balancer
	balancer, err := bl.NewBalancer(cfg.Balancer.Algorithm, cfg.Balancer.Backends)
	if err != nil {
		log.Error("failed to init balancer", logger.Err(err))
		os.Exit(1)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	maxRetries := cfg.Retries
	limiter := rate_limiter.NewRedisRateLimiter(cfg)
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

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: chain,
	}

	checker := health_check.NewHealthChecker(cfg.HealthChecker.Interval,
		balancer,
		cfg.HealthChecker.CheckURL,
		log)

	go checker.Start()

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", logger.Err(err))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	//TODO: настроить вывод логгера
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
