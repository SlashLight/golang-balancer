package health_check

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/SlashLight/golang-balancer/internal/balancer"
)

type BackendGetter interface {
	GetAllBackends() []*balancer.Backend
}
type HealthChecker struct {
	interval       time.Duration
	Backend        []*balancer.Backend
	HealthCheckURL string
	mu             sync.RWMutex
	log            *slog.Logger
}

func NewHealthChecker(timer time.Duration, backs BackendGetter, checkURL string, log *slog.Logger) *HealthChecker {
	return &HealthChecker{
		interval:       timer,
		Backend:        backs.GetAllBackends(),
		HealthCheckURL: checkURL,
		mu:             sync.RWMutex{},
		log:            log,
	}
}

// TODO: добавить логи
func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	for range ticker.C {
		hc.mu.RLock()
		backends := hc.Backend
		hc.mu.RUnlock()

		for _, back := range backends {
			resp, err := http.Get(back.URL.String() + hc.HealthCheckURL)

			isAlive := err == nil && resp.StatusCode < 500
			hc.mu.RLock()
			if isAlive && !back.Alive {
				hc.log.Info("backend is now alive", back.URL.String())
			} else if !isAlive && back.Alive {
				hc.log.Info("backend doesnt respond correctly", back.URL.String())
			}

			if isAlive != back.Alive {
				hc.mu.Lock()
				back.SetAlive(isAlive)
				hc.mu.Unlock()
			}

			hc.mu.RUnlock()
		}
	}
}
