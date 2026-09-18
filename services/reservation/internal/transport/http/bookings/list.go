package bookings

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/reservation/internal/transport/http/response"
)

func registerList(api huma.API, service Service, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "list-reservations",
		Method:      http.MethodGet,
		Path:        "/reservations",
		Tags:        []string{reservationsTag},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, func(ctx context.Context, input *UserInput) (*ListOutput, error) {
		items, err := service.List(ctx, input.Username)
		if err != nil {
			return nil, response.Error(ctx, logger, err)
		}
		return &ListOutput{Body: items}, nil
	})
}
