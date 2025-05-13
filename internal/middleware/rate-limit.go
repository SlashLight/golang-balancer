package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/SlashLight/golang-balancer/internal/api"
	resp "github.com/SlashLight/golang-balancer/internal/api/response"
	"github.com/SlashLight/golang-balancer/internal/logger"
)

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

func RateLimitMiddleware(limiter RateLimiter, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userIP, err := api.GetIpFromRequest(r)
			if err != nil {
				log.Error("Error trying to get IP addr", logger.Err(err))
				resp.RespondError(w, http.StatusInternalServerError, "Internal error", log)
				return
			}

			allowed, err := limiter.Allow(r.Context(), userIP)
			if err != nil {
				log.Error("Error trying to get rate limits for user", logger.Err(err))
				resp.RespondError(w, http.StatusInternalServerError, "Internal error", log)
				return
			}

			if !allowed {
				log.Info("User reached the limit", userIP)
				resp.RespondError(w, http.StatusTooManyRequests, "Rate limit exceeded", log)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
