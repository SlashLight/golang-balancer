package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	bl "github.com/SlashLight/golang-balancer/internal/balancer"
	"github.com/SlashLight/golang-balancer/internal/config"
	"github.com/SlashLight/golang-balancer/internal/logger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"

	RoundRobinAlgorithm = "round-robin"
	HashAlgorithm       = "hash"
)

func main() {
	// TODO [x] загрузить конфиг
	cfg := config.MustLoad()

	// TODO [x] создать логгер
	log := setupLogger(cfg.Env)

	log.Info(
		"starting golang balancer",
		slog.String("env", cfg.Env),
	)
	// TODO [x] создать балансировщик

	var balancer bl.Balancer
	var err error

	//TODO: добавить round-robin balancer
	switch cfg.Algorithm {
	case HashAlgorithm:
		balancer, err = bl.NewHashBalancer(cfg.Backends)
	default:
		log.Error("unknown balancer's algorithm", logger.Err(err))
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backend, err := balancer.Next(r)
		if err != nil {
			log.Error("error trying to get new server", logger.Err(err))
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(backend.URL)
		proxy.ServeHTTP(w, r)
	})

	// TODO [x] запустить сервер

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: handler,
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
