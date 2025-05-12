package balancer

import (
	"net/http"

	"github.com/SlashLight/golang-balancer/pkg/my_err"
)

type Balancer interface {
	Next(r *http.Request) (*Backend, error)
	GetAllBackends() []*Backend
}

const (
	RoundRobinAlgorithm = "round-robin"
	HashAlgorithm       = "hash"
)

func NewBalancer(algorithm string, backendURLs []string) (Balancer, error) {
	var err error
	var balancer Balancer
	switch algorithm {
	case HashAlgorithm:
		balancer, err = NewHashBalancer(backendURLs)
	default:
		err = my_err.ErrUnknownAlgotirhm
	}

	if err != nil {
		return nil, err
	}

	return balancer, nil
}
