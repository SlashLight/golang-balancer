package balancer

import (
	"net/http"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type Balancer interface {
	Next(r *http.Request) (*Backend, error)
	AddNewBackend(*Backend)
	RemoveBackend(int)
}

const (
	RoundRobinAlgorithm = "round-robin"
	HashAlgorithm       = "hash"
	LeastConnections    = "least-connections"
)

func NewBalancer(algorithm string, backendURLs []string) (Balancer, error) {
	var err error
	var balancer Balancer
	switch algorithm {
	case HashAlgorithm:
		balancer, err = NewHashBalancer(backendURLs)
	case RoundRobinAlgorithm:
		balancer, err = NewRoundRobinBalancer(backendURLs)
	case LeastConnections:
		balancer, err = NewLeastConnectionBalancer(backendURLs)
	default:
		err = my_err.ErrUnknownAlgorithm
	}

	if err != nil {
		return nil, err
	}

	return balancer, nil
}
