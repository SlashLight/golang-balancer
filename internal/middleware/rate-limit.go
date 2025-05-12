package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	resp "github.com/SlashLight/golang-balancer/internal/api/response"
	"github.com/SlashLight/golang-balancer/internal/logger"
)

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

func RateLimitMiddleware(limiter RateLimiter, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			userIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				log.Error("Error trying to get IP addr", logger.Err(err))
				err = resp.RespondError(w, http.StatusInternalServerError, "Internal error")
				if err != nil {
					log.Error("Error at sending message", logger.Err(err))
				}
				return
			}
			userIP = strings.Replace(userIP, ":", "_", -1)

			allowed, err := limiter.Allow(r.Context(), userIP)
			if err != nil {
				log.Error("Error trying to get rate limits for user", logger.Err(err))
				err = resp.RespondError(w, http.StatusInternalServerError, "Internal error")
				if err != nil {
					log.Error("Error at sending message", logger.Err(err))
				}
				return
			}

			if !allowed {
				log.Info("User reached the limit", userIP)
				err = resp.RespondError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				if err != nil {
					log.Error("Error at sending message", logger.Err(err))
				}
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
