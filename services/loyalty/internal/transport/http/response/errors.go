package response

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/oops"
	"lab2/loyalty/internal/domain"
)

func Error(ctx context.Context, logger *slog.Logger, err error) error {
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, message = http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrInvalidInput):
		status, message = http.StatusBadRequest, oops.GetPublic(err, "invalid input")
		if errors.Is(err, domain.ErrInvalidUsername) {
			message = "valid X-User-Name is required"
		}
	default:
		logger.ErrorContext(ctx, "request failed", "error", err)
	}
	return oops.FromContext(ctx).Wrap(huma.NewError(status, message))
}
