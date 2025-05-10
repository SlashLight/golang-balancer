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
	}, nil
}

func (hb *HashBalancer) Next(r *http.Request) (*Backend, error) {
	clientIP := r.RemoteAddr
	if clientIP == "" {
		return nil, my_err.ErrNoClientAddr
	}

	hash := hb.hasher(clientIP)

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
		if back.Alive {
			aliveBackends = append(aliveBackends, back)
		}
	}

	return aliveBackends
}
