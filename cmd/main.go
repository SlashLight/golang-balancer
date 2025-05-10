package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	bl "github.com/SlashLight/golang-balancer/internal/balancer"
	"github.com/SlashLight/golang-balancer/internal/config"
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

type Balancer interface {
	Next(r *http.Request) (*bl.Backend, error)
}

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
		balancer, err = bl.NewHashBalancer(cfg.Backends)
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
