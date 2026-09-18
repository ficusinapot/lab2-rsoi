package bookings

import (
	"context"
	"uuid"

	"github.com/samber/oops"

	"github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

	"lab2/reservation/ent"
	"lab2/reservation/ent/reservation"
	"lab2/reservation/internal/domain"
	"lab2/reservation/internal/domain/bookings"
)

type Repository struct{ client *ent.Client }

func New(client *ent.Client) *Repository { return &Repository{client: client} }

func (r *Repository) List(ctx context.Context, name string) (_ []bookings.Reservation, err error) {
	defer autometrics.Instrument(autometrics.PreInstrument(ctx), &err)
	rows, err := r.client.Reservation.Query().
		Where(reservation.UsernameEQ(name)).
		WithHotel().Order(ent.Asc(reservation.FieldID)).All(ctx)
	if err != nil {
		return nil, oops.FromContext(ctx).In("bookings.postgres").With("operation", "list").Wrapf(err, "operation failed")
	}
	items := make([]bookings.Reservation, 0, len(rows))
	for _, row := range rows {
		item, err := model(ctx, row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID, name string) (_ bookings.Reservation, err error) {
	defer autometrics.Instrument(autometrics.PreInstrument(ctx), &err)
	row, err := r.client.Reservation.Query().
		Where(reservation.ReservationUIDEQ(id), reservation.UsernameEQ(name)).
		WithHotel().Only(ctx)
	if ent.IsNotFound(err) {
		return bookings.Reservation{}, oops.In("repository").Code("not_found").
			With("entity", "reservation").
			Public("not found").
			Wrap(domain.ErrNotFound)
	}
	if err != nil {
		return bookings.Reservation{}, oops.FromContext(ctx).In("bookings.postgres").
			With("operation", "get").
			Wrapf(err, "operation failed")
	}
	return model(ctx, row)
}

func (r *Repository) Cancel(ctx context.Context, id uuid.UUID, name string) (err error) {
	defer autometrics.Instrument(autometrics.PreInstrument(ctx), &err)
	count, err := r.client.Reservation.Update().
		Where(reservation.ReservationUIDEQ(id), reservation.UsernameEQ(name)).
		SetStatus(reservation.Status(bookings.StatusCanceled)).Save(ctx)
	if err != nil {
		return oops.FromContext(ctx).In("bookings.postgres").With("operation", "cancel").Wrapf(err, "operation failed")
	}
	if count == 0 {
		return oops.In("repository").Code("not_found").
			With("entity", "reservation").
			Public("not found").
			Wrap(domain.ErrNotFound)
	}
	return nil
}
