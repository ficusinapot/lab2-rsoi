package entities

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrNotFound            = errors.New("not found")
	ErrUpstreamUnavailable = errors.New("upstream unavailable")
	ErrUpstreamTimeout     = errors.New("upstream timeout")
)
