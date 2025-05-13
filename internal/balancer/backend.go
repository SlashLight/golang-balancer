package balancer

import (
	"net/url"
	"sync"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	index int
	mu    sync.RWMutex
}

func (b *Backend) SetAlive(status bool) {
	b.mu.Lock()
	b.Alive = status
	b.mu.Unlock()
}

func GetBackendsFromURLS(backendsURLs []string) ([]*Backend, error) {
	backends := make([]*Backend, len(backendsURLs))
	for idx := range backends {
		backURL, err := url.Parse(backendsURLs[idx])
		if err != nil {
			return nil, my_err.ErrParsingBackendURL
		}

		backends[idx] = &Backend{
			URL:   backURL,
			Alive: true,
			index: idx,
			mu:    sync.RWMutex{},
		}
	}

	return backends, nil
}
