package balancer

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type RoundRobinBalancer struct {
	backends []*Backend
	current  uint64
	mu       sync.RWMutex
}

func NewRoundRobinBalancer(backendsURLs []string) (*RoundRobinBalancer, error) {
	backends, err := GetBackendsFromURLS(backendsURLs)
	if err != nil {
		return nil, fmt.Errorf("Error at creating round-robin balancer: %w", err)
	}

	return &RoundRobinBalancer{
		backends: backends,
		current:  0,
		mu:       sync.RWMutex{},
	}, nil
}

func (rr *RoundRobinBalancer) Next(r *http.Request) (*Backend, error) {
	rr.mu.RLock()
	defer rr.mu.RUnlock()
	if len(rr.backends) == 0 {
		return nil, my_err.ErrNoAliveBackends
	}
	index := atomic.AddUint64(&rr.current, 1) % uint64(len(rr.backends))

	return rr.backends[index], nil
}

func (rr *RoundRobinBalancer) getAliveBackends() []*Backend {
	var aliveBackends []*Backend
	for _, back := range rr.backends {
		rr.mu.RLock()
		if back.Alive {
			aliveBackends = append(aliveBackends, back)
		}
		rr.mu.RUnlock()
	}

	return aliveBackends
}

func (rr *RoundRobinBalancer) AddNewBackend(back *Backend) {
	rr.backends = append(rr.backends, back)
}

func (rr *RoundRobinBalancer) RemoveBackend(idx int) {
	rr.backends = append(rr.backends[:idx], rr.backends[idx+1:]...)
}
