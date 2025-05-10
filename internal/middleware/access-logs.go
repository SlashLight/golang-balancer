package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func AccessLog(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/logger"),
		)

		log.Info("logger middleware enabled")

		requestID := uuid.New().String()

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "requestID", requestID)
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", requestID),
			)
			responseWithCode := NewResponseRecorder(w)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("code", responseWithCode.StatusCode),
					slog.String("duration", time.Since(t1).String()),
				)

			}()

			next.ServeHTTP(responseWithCode, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
