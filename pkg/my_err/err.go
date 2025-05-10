package my_err

import (
	"errors"
)

var (
	ErrNoClientAddr      = errors.New("empty client IP and port")
	ErrNoAliveBackends   = errors.New("no alive backends")
	ErrParsingBackendURL = errors.New("error at parsing backend URL")
)
