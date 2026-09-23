package clients

import (
	"context"
	"net/http"
	"net/url"

	"lab2/gateway/internal/domain/bookings"
)

func (c *Client) Payment(ctx context.Context, id string) (bookings.Payment, error) {
	var out bookings.Payment
	err := c.call(ctx, http.MethodGet, c.cfg.PaymentURL+"/payments/"+url.PathEscape(id), "", nil, &out)
	return out, err
}

func (c *Client) CreatePayment(ctx context.Context, price int) (bookings.InternalPayment, error) {
	var out bookings.InternalPayment
	input := struct {
		Price int `json:"price"`
	}{Price: price}
	err := c.call(ctx, http.MethodPost, c.cfg.PaymentURL+"/payments", "", input, &out)
	return out, err
}

func (c *Client) CancelPayment(ctx context.Context, id string) error {
	return c.call(ctx, http.MethodDelete, c.cfg.PaymentURL+"/payments/"+url.PathEscape(id), "", nil, nil)
}
