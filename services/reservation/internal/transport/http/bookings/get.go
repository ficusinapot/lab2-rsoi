package bookings

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/reservation/internal/transport/http/response"
)

func registerGet(api huma.API, service Service, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "get-reservation",
		Method:      http.MethodGet,
		Path:        "/reservations/{reservationUid}",
		Tags:        []string{reservationsTag},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, func(ctx context.Context, input *ItemInput) (*ItemOutput, error) {
		item, err := service.Get(ctx, input.UID, input.Username)
		if err != nil {
			return nil, response.Error(ctx, logger, err)
		}
		return &ItemOutput{Body: item}, nil
	})
}
