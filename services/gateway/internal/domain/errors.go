package domain

import "errors"

var ErrInvalidInput = errors.New("invalid input")

type UpstreamError struct{ Status int }

func (e *UpstreamError) Error() string { return "upstream request failed" }
