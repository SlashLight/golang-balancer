package balancer

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type HashBalancer struct {
	backends []*Backend
	hasher   func(string) int // TODO: заменить интерфейсом
	mu       sync.RWMutex
}

func NewHashBalancer(backendsURLs []string) (*HashBalancer, error) {
	backends, err := GetBackendsFromURLS(backendsURLs)
	if err != nil {
		return nil, fmt.Errorf("Error at creating Hash balancer: %w", err)
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

	index := hash % len(hb.backends)

	return hb.backends[index], nil
}

func (hb *HashBalancer) AddNewBackend(back *Backend) {
	hb.backends = append(hb.backends, back)
}

func (hb *HashBalancer) RemoveBackend(idx int) {
	hb.backends = append(hb.backends[:idx], hb.backends[idx+1:]...)
}
