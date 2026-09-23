package bookings

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/gateway/internal/domain/bookings"
	"lab2/platform/rest"
)

type Service interface {
	Hotels(context.Context, string, string) (bookings.Page, error)
	List(context.Context, string) ([]bookings.Reservation, error)
	Get(context.Context, string, string) (bookings.Reservation, error)
	Me(context.Context, string) (bookings.UserInfo, error)
	Loyalty(context.Context, string) (bookings.Loyalty, error)
	Create(context.Context, string, bookings.CreateRequest) (bookings.Created, error)
	Cancel(context.Context, string, string) error
}
type (
	UserInput struct {
		Username string `header:"X-User-Name" required:"true"`
	}
	ItemInput struct {
		UserInput
		UID string `path:"reservationUid"`
	}
	HotelsInput struct {
		Page string `query:"page"`
		Size string `query:"size"`
	}
	CreateInput struct {
		UserInput
		Body bookings.CreateRequest
	}
	HotelsOutput  struct{ Body bookings.Page }
	ListOutput    struct{ Body []bookings.Reservation }
	ItemOutput    struct{ Body bookings.Reservation }
	MeOutput      struct{ Body bookings.UserInfo }
	LoyaltyOutput struct{ Body bookings.Loyalty }
	CreateOutput  struct{ Body bookings.Created }
)

func Register(api huma.API, service Service, logger *slog.Logger, cfg rest.Config) {
	huma.Register(api, huma.Operation{OperationID: "list-hotels", Method: http.MethodGet, Path: "/hotels"},
		func(ctx context.Context, in *HotelsInput) (*HotelsOutput, error) {
			item, err := service.Hotels(ctx, in.Page, in.Size)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &HotelsOutput{Body: item}, nil
		})
	huma.Register(api, huma.Operation{OperationID: "list-reservations", Method: http.MethodGet, Path: "/reservations"},
		func(ctx context.Context, in *UserInput) (*ListOutput, error) {
			items, err := service.List(ctx, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &ListOutput{Body: items}, nil
		})
	huma.Register(api, huma.Operation{
		OperationID: "get-reservation", Method: http.MethodGet, Path: "/reservations/{reservationUid}",
	},
		func(ctx context.Context, in *ItemInput) (*ItemOutput, error) {
			item, err := service.Get(ctx, in.UID, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &ItemOutput{Body: item}, nil
		})
	huma.Register(api, huma.Operation{OperationID: "get-me", Method: http.MethodGet, Path: "/me"},
		func(ctx context.Context, in *UserInput) (*MeOutput, error) {
			item, err := service.Me(ctx, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &MeOutput{Body: item}, nil
		})
	huma.Register(api, huma.Operation{OperationID: "get-loyalty", Method: http.MethodGet, Path: "/loyalty"},
		func(ctx context.Context, in *UserInput) (*LoyaltyOutput, error) {
			item, err := service.Loyalty(ctx, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &LoyaltyOutput{Body: item}, nil
		})
	huma.Register(api, huma.Operation{
		OperationID: "create-reservation", Method: http.MethodPost, Path: "/reservations",
		DefaultStatus: http.StatusOK, MaxBodyBytes: cfg.MaxBodyBytes, BodyReadTimeout: cfg.ReadTimeout,
	},
		func(ctx context.Context, in *CreateInput) (*CreateOutput, error) {
			item, err := service.Create(ctx, in.Username, in.Body)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &CreateOutput{Body: item}, nil
		})
	huma.Register(api, huma.Operation{
		OperationID: "cancel-reservation", Method: http.MethodDelete, Path: "/reservations/{reservationUid}",
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, in *ItemInput) (*struct{}, error) {
		if err := service.Cancel(ctx, in.UID, in.Username); err != nil {
			return nil, responseError(ctx, logger, err)
		}
		return nil, nil
	})
}
