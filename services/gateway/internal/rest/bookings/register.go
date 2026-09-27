package bookings

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/gateway/internal/models/coreifc"
	"lab2/platform/rest"
)

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
		Body CreateRequest
	}
	HotelsOutput  struct{ Body PageResponse }
	ListOutput    struct{ Body []ReservationResponse }
	ItemOutput    struct{ Body ReservationResponse }
	MeOutput      struct{ Body UserInfoResponse }
	LoyaltyOutput struct{ Body LoyaltyResponse }
	CreateOutput  struct{ Body CreatedResponse }
)

func Register(api huma.API, service coreifc.Bookings, logger *slog.Logger, cfg rest.Config) {
	api = configureAPI(api)
	huma.Register(api, withErrorResponse(api,
		huma.Operation{OperationID: "list-hotels", Method: http.MethodGet, Path: "/hotels"}),
		func(ctx context.Context, in *HotelsInput) (*HotelsOutput, error) {
			item, err := service.Hotels(ctx, in.Page, in.Size)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &HotelsOutput{Body: pageResponse(item)}, nil
		})
	huma.Register(api, withErrorResponse(api,
		huma.Operation{OperationID: "list-reservations", Method: http.MethodGet, Path: "/reservations"}),
		func(ctx context.Context, in *UserInput) (*ListOutput, error) {
			items, err := service.List(ctx, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &ListOutput{Body: reservationResponses(items)}, nil
		})
	huma.Register(api, withErrorResponse(api, huma.Operation{
		OperationID: "get-reservation", Method: http.MethodGet, Path: "/reservations/{reservationUid}",
	}),
		func(ctx context.Context, in *ItemInput) (*ItemOutput, error) {
			item, err := service.Get(ctx, in.UID, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &ItemOutput{Body: reservationResponse(item)}, nil
		})
	huma.Register(api, withErrorResponse(api,
		huma.Operation{OperationID: "get-me", Method: http.MethodGet, Path: "/me"}),
		func(ctx context.Context, in *UserInput) (*MeOutput, error) {
			item, err := service.Me(ctx, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &MeOutput{Body: userInfoResponse(item)}, nil
		})
	huma.Register(api, withErrorResponse(api,
		huma.Operation{OperationID: "get-loyalty", Method: http.MethodGet, Path: "/loyalty"}),
		func(ctx context.Context, in *UserInput) (*LoyaltyOutput, error) {
			item, err := service.Loyalty(ctx, in.Username)
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &LoyaltyOutput{Body: loyaltyResponse(item)}, nil
		})
	huma.Register(api, withErrorResponse(api, huma.Operation{
		OperationID: "create-reservation", Method: http.MethodPost, Path: "/reservations",
		DefaultStatus: http.StatusOK, MaxBodyBytes: cfg.MaxBodyBytes, BodyReadTimeout: cfg.ReadTimeout,
	}),
		func(ctx context.Context, in *CreateInput) (*CreateOutput, error) {
			item, err := service.Create(ctx, in.Username, in.Body.domain())
			if err != nil {
				return nil, responseError(ctx, logger, err)
			}
			return &CreateOutput{Body: createdResponse(item)}, nil
		})
	huma.Register(api, withErrorResponse(api, huma.Operation{
		OperationID: "cancel-reservation", Method: http.MethodDelete, Path: "/reservations/{reservationUid}",
		DefaultStatus: http.StatusNoContent,
	}), func(ctx context.Context, in *ItemInput) (*struct{}, error) {
		if err := service.Cancel(ctx, in.UID, in.Username); err != nil {
			return nil, responseError(ctx, logger, err)
		}
		return nil, nil
	})
}
