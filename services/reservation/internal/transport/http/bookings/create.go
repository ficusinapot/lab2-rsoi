package bookings

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/platform/rest"
	"lab2/reservation/internal/transport/http/response"
	usecase "lab2/reservation/internal/usecase/bookings"
)

func registerCreate(api huma.API, service Service, logger *slog.Logger, cfg rest.Config) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-reservation",
		Method:        http.MethodPost,
		Path:          "/reservations",
		Tags:          []string{reservationsTag},
		DefaultStatus: http.StatusCreated,
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		},
		MaxBodyBytes:    cfg.MaxBodyBytes,
		BodyReadTimeout: cfg.ReadTimeout,
	}, func(ctx context.Context, input *CreateInput) (*CreateOutput, error) {
		item, err := service.Create(ctx, input.Username, usecase.CreateInput{
			HotelUID:   input.Body.HotelUID,
			PaymentUID: input.Body.PaymentUID,
			StartDate:  input.Body.StartDate,
			EndDate:    input.Body.EndDate,
		})
		if err != nil {
			return nil, response.Error(ctx, logger, err)
		}
		return &CreateOutput{
			Location: cfg.APIPrefix + "/reservations/" + item.ReservationUID.String(),
			Body:     item,
		}, nil
	})
}
