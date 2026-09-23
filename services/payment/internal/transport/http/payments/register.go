package payments

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/payment/internal/domain/payments"
	"lab2/payment/internal/transport/http/response"
	"lab2/platform/rest"
)

type Service interface {
	Create(context.Context, int) (payments.Payment, error)
	Get(context.Context, string) (payments.Payment, error)
	Cancel(context.Context, string) error
}
type (
	CreateInput struct {
		Body struct {
			Price int `json:"price"`
		}
	}
	ItemInput struct {
		UID string `path:"paymentUid"`
	}
	Output struct{ Body payments.Payment }
)

func Register(api huma.API, service Service, logger *slog.Logger, cfg rest.Config) {
	huma.Register(api, huma.Operation{
		OperationID: "create-payment", Method: http.MethodPost, Path: "/payments",
		DefaultStatus: http.StatusCreated, MaxBodyBytes: cfg.MaxBodyBytes, BodyReadTimeout: cfg.ReadTimeout,
	},
		func(ctx context.Context, in *CreateInput) (*Output, error) {
			item, err := service.Create(ctx, in.Body.Price)
			if err != nil {
				return nil, response.Error(ctx, logger, err)
			}
			return &Output{Body: item}, nil
		})
	huma.Register(api, huma.Operation{OperationID: "get-payment", Method: http.MethodGet, Path: "/payments/{paymentUid}"},
		func(ctx context.Context, in *ItemInput) (*Output, error) {
			item, err := service.Get(ctx, in.UID)
			if err != nil {
				return nil, response.Error(ctx, logger, err)
			}
			return &Output{Body: item}, nil
		})
	huma.Register(api, huma.Operation{
		OperationID: "cancel-payment", Method: http.MethodDelete, Path: "/payments/{paymentUid}",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, in *ItemInput) (*struct{}, error) {
		if err := service.Cancel(ctx, in.UID); err != nil {
			return nil, response.Error(ctx, logger, err)
		}
		return nil, nil
	})
}
