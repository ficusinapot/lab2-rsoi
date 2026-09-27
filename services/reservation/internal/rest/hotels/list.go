package hotels

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"lab2/reservation/internal/rest/response"
)

func registerList(api huma.API, service Service, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID: "list-hotels",
		Method:      http.MethodGet,
		Path:        "/hotels",
		Tags:        []string{hotelsTag},
		Errors:      []int{http.StatusBadRequest, http.StatusInternalServerError},
	}, func(ctx context.Context, input *ListInput) (*ListOutput, error) {
		page, err := service.List(ctx, input.Page, input.Size)
		if err != nil {
			return nil, response.Error(ctx, logger, err)
		}
		return &ListOutput{Body: pageResponse(page)}, nil
	})
}
