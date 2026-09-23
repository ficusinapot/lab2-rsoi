package clients

import (
	"context"
	"net/http"

	"lab2/gateway/internal/domain/bookings"
)

func (c *Client) Loyalty(ctx context.Context, user string) (bookings.Loyalty, error) {
	var out bookings.Loyalty
	err := c.call(ctx, http.MethodGet, c.cfg.LoyaltyURL+"/loyalty", user, nil, &out)
	return out, err
}

func (c *Client) ChangeLoyalty(ctx context.Context, user string, delta int) error {
	method := http.MethodPost
	if delta < 0 {
		method = http.MethodDelete
	}
	return c.call(ctx, method, c.cfg.LoyaltyURL+"/loyalty/reservations", user, nil, nil)
}
