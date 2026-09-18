package bookings

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/reservation/internal/transport/http/response"
)

func registerCancel(api huma.API, service Service, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID:   "cancel-reservation",
		Method:        http.MethodDelete,
		Path:          "/reservations/{reservationUid}",
		Tags:          []string{reservationsTag},
		DefaultStatus: http.StatusNoContent,
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, func(ctx context.Context, input *ItemInput) (*struct{}, error) {
		if err := service.Cancel(ctx, input.UID, input.Username); err != nil {
			return nil, response.Error(ctx, logger, err)
		}
		return nil, nil
	})
}
