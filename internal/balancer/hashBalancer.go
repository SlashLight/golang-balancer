package balancer

import (
	"hash/fnv"
	"net/http"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type HashBalancer struct {
	backends []*Backend
	hasher   func(string) int // TODO: заменить интерфейсом
}

func NewHashBalancer(backends []*Backend) *HashBalancer {
	return &HashBalancer{
		backends: backends,
		hasher: func(key string) int {
			h := fnv.New32a()
			h.Write([]byte(key))
			return int(h.Sum32())
		},
	}
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
