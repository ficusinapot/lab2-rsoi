package bookings

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"lab2/gateway/internal/domain"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/oops"
)

type (
	FieldError struct {
		Field string `json:"field"`
		Error string `json:"error"`
	}
	APIError struct {
		Status  int          `json:"-"`
		Message string       `json:"message"`
		Errors  []FieldError `json:"errors,omitempty"`
	}
)

func (e *APIError) Error() string  { return e.Message }
func (e *APIError) GetStatus() int { return e.Status }

// ConfigureErrors is called once before serving requests. Huma's error factory is process-wide.
func ConfigureErrors() {
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		if status == http.StatusUnprocessableEntity {
			status = http.StatusBadRequest
		}
		result := &APIError{Status: status, Message: message}
		for _, err := range errs {
			if detail, ok := errors.AsType[*huma.ErrorDetail](err); ok {
				result.Errors = append(result.Errors, FieldError{Field: detail.Location, Error: detail.Message})
			}
		}
		return result
	}
}

func responseError(ctx context.Context, logger *slog.Logger, err error) error {
	status := http.StatusInternalServerError
	message := "internal server error"
	var upstream *domain.UpstreamError
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		message = "invalid input"
	case errors.As(err, &upstream):
		status = upstream.Status
		message = http.StatusText(status)
	}
	if status >= 500 {
		logger.ErrorContext(ctx, "request failed", "error", err)
	}
	return oops.Wrap(&APIError{Status: status, Message: message})
}
