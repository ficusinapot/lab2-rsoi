package repos

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/samber/oops"
	"lab2/gateway/internal/httpclient"
	bookings "lab2/gateway/internal/models/entities"
	"lab2/gateway/internal/reservation/converters"
)

type Repository struct {
	client  *httpclient.Client
	baseURL string
}

func New(client *httpclient.Client, baseURL string) *Repository {
	return &Repository{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (r *Repository) Hotels(ctx context.Context, query bookings.PageQuery) (bookings.Page, error) {
	values := url.Values{}
	if query.Page != nil {
		values.Set("page", strconv.Itoa(*query.Page))
	}
	if query.Size != nil {
		values.Set("size", strconv.Itoa(*query.Size))
	}
	address := r.baseURL + "/hotels"
	if len(values) > 0 {
		address += "?" + values.Encode()
	}
	var out converters.Page
	if err := r.client.Call(ctx, http.MethodGet, address, "", nil, &out); err != nil {
		return bookings.Page{}, oops.Wrapf(err, "list hotels")
	}
	return out.Entity(), nil
}

func (r *Repository) Hotel(ctx context.Context, id string) (bookings.Hotel, error) {
	var out converters.Hotel
	err := r.client.Call(ctx, http.MethodGet, r.baseURL+"/hotels/"+url.PathEscape(id), "", nil, &out)
	return out.Entity(), err
}

func (r *Repository) Reservations(ctx context.Context, user string) ([]bookings.InternalReservation, error) {
	var out []converters.Reservation
	err := r.client.Call(ctx, http.MethodGet, r.baseURL+"/reservations", user, nil, &out)
	if err != nil {
		return nil, oops.Wrapf(err, "list reservations")
	}
	items := make([]bookings.InternalReservation, len(out))
	for i, item := range out {
		items[i] = item.Entity()
	}
	return items, nil
}

func (r *Repository) Reservation(ctx context.Context, id, user string) (bookings.InternalReservation, error) {
	var out converters.Reservation
	err := r.client.Call(ctx, http.MethodGet, r.baseURL+"/reservations/"+url.PathEscape(id), user, nil, &out)
	return out.Entity(), err
}

func (r *Repository) CreateReservation(
	ctx context.Context, user string, input bookings.InternalCreate,
) (bookings.InternalReservation, error) {
	var out converters.Reservation
	err := r.client.Call(ctx, http.MethodPost, r.baseURL+"/reservations", user,
		converters.NewCreateRequest(input), &out)
	return out.Entity(), err
}

func (r *Repository) CancelReservation(ctx context.Context, id, user string) error {
	return oops.Wrapf(r.client.Call(
		ctx, http.MethodDelete, r.baseURL+"/reservations/"+url.PathEscape(id), user, nil, nil,
	), "cancel reservation")
}
