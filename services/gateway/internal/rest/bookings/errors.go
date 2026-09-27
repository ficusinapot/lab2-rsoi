package bookings

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"reflect"

	"lab2/gateway/internal/models/entities"

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

type gatewayAPI struct{ huma.API }

func (a gatewayAPI) Transform(ctx huma.Context, status string, value any) (any, error) {
	if source, ok := value.(*huma.ErrorModel); ok {
		result := &APIError{Status: source.Status, Message: source.Detail}
		for _, detail := range source.Errors {
			if detail != nil {
				result.Errors = append(result.Errors, FieldError{Field: detail.Location, Error: detail.Message})
			}
		}
		value = result
	}
	result, err := a.API.Transform(ctx, status, value)
	return result, oops.Wrapf(err, "transform response")
}

type humaContext interface{ huma.Context }

type gatewayContext struct{ humaContext }

func (c gatewayContext) SetStatus(status int) {
	if status == http.StatusUnprocessableEntity {
		status = http.StatusBadRequest
	}
	c.humaContext.SetStatus(status)
}

func configureAPI(api huma.API) huma.API {
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) { next(gatewayContext{ctx}) })
	return gatewayAPI{api}
}

func withErrorResponse(api huma.API, op huma.Operation) huma.Operation {
	if op.Responses == nil {
		op.Responses = make(map[string]*huma.Response)
	}
	schema := api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[APIError](), true, "GatewayError")
	op.Responses["default"] = &huma.Response{
		Description: "Error",
		Content: map[string]*huma.MediaType{
			"application/json": {Schema: schema},
		},
	}
	return op
}

func responseError(ctx context.Context, logger *slog.Logger, err error) error {
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, entities.ErrInvalidInput):
		status = http.StatusBadRequest
		message = "invalid input"
	case errors.Is(err, entities.ErrNotFound):
		status = http.StatusNotFound
		message = http.StatusText(status)
	case errors.Is(err, entities.ErrUpstreamTimeout):
		status = http.StatusGatewayTimeout
		message = http.StatusText(status)
	case errors.Is(err, entities.ErrUpstreamUnavailable):
		status = http.StatusBadGateway
		message = http.StatusText(status)
	}
	if status >= 500 {
		logger.ErrorContext(ctx, "request failed", "error", err)
	}
	return oops.Wrap(&APIError{Status: status, Message: message})
}
