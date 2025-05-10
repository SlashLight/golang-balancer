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
