package health_check

import (
	"log/slog"
	"net/http"
	url2 "net/url"
	"sync"
	"time"

	"github.com/SlashLight/golang-balancer/internal/balancer"
)

type Balancer interface {
	AddNewBackend(*balancer.Backend)
	RemoveBackend(int)
}
type HealthChecker struct {
	interval       time.Duration
	Backend        []*balancer.Backend
	HealthCheckURL string
	mu             sync.RWMutex
	log            *slog.Logger
}

func NewHealthChecker(timer time.Duration, backs []string, checkURL string, log *slog.Logger) (*HealthChecker, error) {
	backends, err := balancer.GetBackendsFromURLS(backs)
	if err != nil {
		return nil, err
	}

	return &HealthChecker{
		interval:       timer,
		Backend:        backends,
		HealthCheckURL: checkURL,
		mu:             sync.RWMutex{},
		log:            log,
	}, nil
}

// TODO: добавить логи
func (hc *HealthChecker) Start(balancer Balancer) {
	ticker := time.NewTicker(hc.interval)
	for range ticker.C {
		hc.mu.RLock()
		backends := hc.Backend
		hc.mu.RUnlock()

		for _, back := range backends {
			doctor := http.Client{Timeout: time.Second}
			healthUrl, err := url2.Parse(back.URL.String() + hc.HealthCheckURL)
			resp, err := doctor.Do(&http.Request{Method: http.MethodGet, URL: healthUrl})

			isAlive := err == nil && resp.StatusCode < 500
			hc.mu.RLock()
			if isAlive && !back.Alive {
				hc.log.Info("backend is now alive", back.URL.String())
				balancer.AddNewBackend(back)

			} else if !isAlive && back.Alive {
				hc.log.Info("backend doesnt respond correctly", back.URL.String())
				balancer.RemoveBackend(back.Index)
			}
			hc.mu.RUnlock()

			if isAlive != back.Alive {
				hc.mu.Lock()
				back.SetAlive(isAlive)
				hc.mu.Unlock()
			}
		}
	}
}
