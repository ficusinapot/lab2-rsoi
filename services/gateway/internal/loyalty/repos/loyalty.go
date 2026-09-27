package repos

import (
	"context"
	"net/http"
	"strings"

	"github.com/samber/oops"
	"lab2/gateway/internal/httpclient"
	"lab2/gateway/internal/loyalty/converters"
	bookings "lab2/gateway/internal/models/entities"
)

type Repository struct {
	client  *httpclient.Client
	baseURL string
}

func New(client *httpclient.Client, baseURL string) *Repository {
	return &Repository{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (r *Repository) Loyalty(ctx context.Context, user string) (bookings.Loyalty, error) {
	var out converters.Loyalty
	err := r.client.Call(ctx, http.MethodGet, r.baseURL+"/loyalty", user, nil, &out)
	return out.Entity(), err
}

func (r *Repository) ChangeLoyalty(ctx context.Context, user string, delta int) error {
	method := http.MethodPost
	if delta < 0 {
		method = http.MethodDelete
	}
	return oops.Wrapf(r.client.Call(ctx, method, r.baseURL+"/loyalty/reservations", user, nil, nil), "change loyalty")
}
