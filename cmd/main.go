package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	bl "github.com/SlashLight/golang-balancer/internal/balancer"
	"github.com/SlashLight/golang-balancer/internal/config"
	health_check "github.com/SlashLight/golang-balancer/internal/health-check"
	"github.com/SlashLight/golang-balancer/internal/logger"
	"github.com/SlashLight/golang-balancer/internal/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"

	RoundRobinAlgorithm = "round-robin"
	HashAlgorithm       = "hash"
)

// TODO: перенести в другое место или избавиться(?)
type Balancer interface {
	Next(r *http.Request) (*bl.Backend, error)
	GetAllBackends() []*bl.Backend
}

//TODO: [x] add healthcheck
//TODO: [] add rate-limits
//TODO: [x] fix concurrency
//TODO: [] add Dockerfile and docker-compose
//TODO: [] add graceful shutdown
//TODO: [] add CRUD
//TODO: [] add database (SQLite)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting golang balancer",
		slog.String("env", cfg.Env),
	)

	var balancer Balancer
	var err error

	//TODO: [] добавить round-robin balancer
	switch cfg.Algorithm {
	case HashAlgorithm:
		balancer, err = bl.NewHashBalancer(cfg.Balancer.Backends)
	default:
		log.Error("unknown balancer's algorithm", logger.Err(err))
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	maxRetries := cfg.Retries
	chain := middleware.AccessLog(log)(
		middleware.RetryMiddleware(balancer, log, maxRetries)(
			handler,
		),
	)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: chain,
	}

	checker := health_check.NewHealthChecker(cfg.HealthChecker.Interval,
		balancer,
		cfg.HealthChecker.CheckURL,
		log)
	go checker.Start()

	server.ListenAndServe()
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
