package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/SlashLight/golang-balancer/internal/api/response"
	bl "github.com/SlashLight/golang-balancer/internal/balancer"
	"github.com/SlashLight/golang-balancer/internal/logger"
)

type Balancer interface {
	Next(r *http.Request) (*bl.Backend, error)
	RemoveBackend(int)
}

type ConnectionTracker interface {
	Release(*bl.Backend)
}

var AllowedMethods = map[string]bool{
	http.MethodGet:  true,
	http.MethodHead: true,
}

func RetryMiddleware(balancer Balancer, log *slog.Logger, maxRetries int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if !AllowedMethods[r.Method] {
				next.ServeHTTP(w, r)
				return
			}

			var body []byte
			var err error
			if r.Body != nil {
				body, err = io.ReadAll(r.Body)
				if err != nil {
					log.Error("error at reading request body", logger.Err(err))
					return
				}
				defer r.Body.Close()
			}

			for attempt := 0; attempt < maxRetries; attempt++ {
				if body != nil {
					r.Body = io.NopCloser(bytes.NewReader(body))
				}

				backend, err := balancer.Next(r)
				if err != nil {
					log.Error("error at getting next alive backend server", logger.Err(err))
					response.RespondError(w, http.StatusServiceUnavailable, "Service unavailable. Try again later", log)
					return
				}

				proxy := httputil.NewSingleHostReverseProxy(backend.URL)
				recorder := NewResponseRecorder(w)
				log.Info("Trying to connect to backend server: ", backend.URL) //TODO подумать над уровнями логирования
				proxy.ServeHTTP(recorder, r)

				if recorder.StatusCode < 500 && !isConnectionError(recorder) {
					if tracker, ok := balancer.(ConnectionTracker); ok {
						defer tracker.Release(backend)
					}
					return
				}

				log.Error("Failed to connect to backend server: ", backend.URL)
				backend.SetAlive(false)
				balancer.RemoveBackend(backend.Index)
			}

			log.Error("Couldn't connect to any server after retries")
			response.RespondError(w, http.StatusTooManyRequests, "Service unavailable. Try again later", log)
		}
		return http.HandlerFunc(fn)
	}
}

func isConnectionError(recorder *ResponseRecorder) bool {
	return strings.Contains(recorder.Body.String(), "connection refused")
}
