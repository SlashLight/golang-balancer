package balancer

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	mu    sync.RWMutex
}

func (b *Backend) SetAlive(status bool) {
	b.mu.Lock()
	b.Alive = status
	b.mu.Unlock()
}
