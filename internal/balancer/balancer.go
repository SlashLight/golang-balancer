package balancer

import (
	"net/http"
	"net/url"
	"sync"
)

type Balancer interface {
	Next(r *http.Request) (*Backend, error)
	getAliveBackends() []*Backend
}

type Backend struct {
	URL   *url.URL
	Alive bool
	mu    sync.RWMutex
}
