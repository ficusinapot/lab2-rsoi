package clients

import (
	"context"
	"net/http"
	"net/url"

	"lab2/gateway/internal/domain/bookings"
)

func (c *Client) Hotels(ctx context.Context, query string) (bookings.Page, error) {
	var out bookings.Page
	err := c.call(ctx, http.MethodGet, c.cfg.ReservationURL+"/hotels"+query, "", nil, &out)
	return out, err
}

func (c *Client) Hotel(ctx context.Context, id string) (bookings.Hotel, error) {
	var out bookings.Hotel
	err := c.call(ctx, http.MethodGet, c.cfg.ReservationURL+"/hotels/"+url.PathEscape(id), "", nil, &out)
	return out, err
}

func (c *Client) Reservations(ctx context.Context, user string) ([]bookings.InternalReservation, error) {
	var out []bookings.InternalReservation
	err := c.call(ctx, http.MethodGet, c.cfg.ReservationURL+"/reservations", user, nil, &out)
	return out, err
}

func (c *Client) Reservation(ctx context.Context, id, user string) (bookings.InternalReservation, error) {
	var out bookings.InternalReservation
	err := c.call(ctx, http.MethodGet, c.cfg.ReservationURL+"/reservations/"+url.PathEscape(id), user, nil, &out)
	return out, err
}

func (c *Client) CreateReservation(
	ctx context.Context, user string, input bookings.InternalCreate,
) (bookings.InternalReservation, error) {
	var out bookings.InternalReservation
	err := c.call(ctx, http.MethodPost, c.cfg.ReservationURL+"/reservations", user, input, &out)
	return out, err
}

func (c *Client) CancelReservation(ctx context.Context, id, user string) error {
	return c.call(ctx, http.MethodDelete, c.cfg.ReservationURL+"/reservations/"+url.PathEscape(id), user, nil, nil)
}
