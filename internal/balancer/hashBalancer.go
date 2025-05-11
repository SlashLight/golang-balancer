package balancer

import (
	"hash/fnv"
	"net/http"
	"net/url"
	"sync"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type HashBalancer struct {
	backends []*Backend
	hasher   func(string) int // TODO: заменить интерфейсом
	mu       sync.RWMutex
}

func NewHashBalancer(backendsURLs []string) (*HashBalancer, error) {
	backends := make([]*Backend, len(backendsURLs))
	for idx := range backends {
		backURL, err := url.Parse(backendsURLs[idx])
		if err != nil {
			return nil, my_err.ErrParsingBackendURL
		}

		backends[idx] = &Backend{
			URL:   backURL,
			Alive: true,
			mu:    sync.RWMutex{},
		}
	}

	return &HashBalancer{
		backends: backends,
		hasher: func(key string) int {
			h := fnv.New32a()
			h.Write([]byte(key))
			return int(h.Sum32())
		},
		mu: sync.RWMutex{},
	}, nil
}

func (hb *HashBalancer) Next(r *http.Request) (*Backend, error) {
	clientIP := r.RemoteAddr
	if clientIP == "" {
		return nil, my_err.ErrNoClientAddr
	}

	hash := hb.hasher(clientIP)

	hb.mu.RLock()
	defer hb.mu.RUnlock()
	aliveBackends := hb.getAliveBackends()
	if len(aliveBackends) == 0 {
		return nil, my_err.ErrNoAliveBackends
	}

	index := hash % len(aliveBackends)

	return aliveBackends[index], nil
}

func (hb *HashBalancer) getAliveBackends() []*Backend {
	var aliveBackends []*Backend
	for _, back := range hb.backends {
		hb.mu.RLock()
		if back.Alive {
			aliveBackends = append(aliveBackends, back)
		}
		hb.mu.RUnlock()
	}

	return aliveBackends
}

func (hb *HashBalancer) GetAllBackends() []*Backend {
	return hb.backends
}
