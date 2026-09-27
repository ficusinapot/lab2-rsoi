package loyalties

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/loyalty/internal/models/coreifc"
	"lab2/loyalty/internal/rest/response"
	"lab2/platform/rest"
)

type (
	Input struct {
		Username string `header:"X-User-Name"`
	}
	Output struct{ Body LoyaltyResponse }
)

func Register(api huma.API, service coreifc.Loyalties, logger *slog.Logger, _ rest.Config) {
	huma.Register(api, huma.Operation{OperationID: "get-loyalty", Method: http.MethodGet, Path: "/loyalty"},
		func(ctx context.Context, in *Input) (*Output, error) {
			item, err := service.Get(ctx, in.Username)
			if err != nil {
				return nil, response.Error(ctx, logger, err)
			}
			return &Output{Body: loyaltyResponse(item)}, nil
		})
	for _, op := range []struct {
		method, id string
		delta      int
	}{{http.MethodPost, "increase-loyalty", 1}, {http.MethodDelete, "decrease-loyalty", -1}} {
		huma.Register(api, huma.Operation{OperationID: op.id, Method: op.method, Path: "/loyalty/reservations"},
			func(ctx context.Context, in *Input) (*Output, error) {
				item, err := service.Change(ctx, in.Username, op.delta)
				if err != nil {
					return nil, response.Error(ctx, logger, err)
				}
				return &Output{Body: loyaltyResponse(item)}, nil
			})
	}
}
