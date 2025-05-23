package my_err

import (
	"errors"
)

var (
	ErrNoClientAddr      = errors.New("empty client IP and port")
	ErrNoAliveBackends   = errors.New("no alive backends")
	ErrParsingBackendURL = errors.New("error at parsing backend URL")
	ErrUnknownAlgorithm  = errors.New("unknown balancing algorithm")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
