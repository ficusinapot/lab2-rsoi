package repos

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/oops"
	"lab2/gateway/internal/httpclient"
	bookings "lab2/gateway/internal/models/entities"
	"lab2/gateway/internal/payment/converters"
)

type Repository struct {
	client  *httpclient.Client
	baseURL string
}

func New(client *httpclient.Client, baseURL string) *Repository {
	return &Repository{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (r *Repository) Payment(ctx context.Context, id string) (bookings.Payment, error) {
	var out converters.Payment
	err := r.client.Call(ctx, http.MethodGet, r.baseURL+"/payments/"+url.PathEscape(id), "", nil, &out)
	return out.Entity(), err
}

func (r *Repository) CreatePayment(ctx context.Context, price int) (bookings.InternalPayment, error) {
	var out converters.Payment
	err := r.client.Call(ctx, http.MethodPost, r.baseURL+"/payments", "",
		converters.CreateRequest{Price: price}, &out)
	return out.InternalEntity(), err
}

func (r *Repository) CancelPayment(ctx context.Context, id string) error {
	return oops.Wrapf(
		r.client.Call(ctx, http.MethodDelete, r.baseURL+"/payments/"+url.PathEscape(id), "", nil, nil),
		"cancel payment",
	)
}
