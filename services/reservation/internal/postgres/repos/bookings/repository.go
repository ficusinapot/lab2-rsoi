package bookings

import (
	"context"
	"uuid"

	"github.com/samber/oops"

	"lab2/platform/observability/metrics"

	"lab2/reservation/ent"
	"lab2/reservation/ent/reservation"
	bookings "lab2/reservation/internal/models/entities"
	bookingconverter "lab2/reservation/internal/postgres/converters/bookings"
)

type Repository struct{ client *ent.Client }

func New(client *ent.Client) *Repository { return &Repository{client: client} }

func (r *Repository) List(ctx context.Context, name string) (_ []bookings.Reservation, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.List), &err)
	rows, err := r.client.Reservation.Query().
		Where(reservation.UsernameEQ(name)).
		WithHotel().Order(ent.Asc(reservation.FieldID)).All(ctx)
	if err != nil {
		return nil, oops.FromContext(ctx).In("bookings.postgres").With("operation", "list").Wrapf(err, "operation failed")
	}
	items := make([]bookings.Reservation, 0, len(rows))
	for _, row := range rows {
		item, err := bookingconverter.Reservation(ctx, row)
		if err != nil {
			return nil, oops.Wrapf(err, "map reservation")
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID, name string) (_ bookings.Reservation, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Get), &err)
	row, err := r.client.Reservation.Query().
		Where(reservation.ReservationUIDEQ(id), reservation.UsernameEQ(name)).
		WithHotel().Only(ctx)
	if ent.IsNotFound(err) {
		return bookings.Reservation{}, oops.In("repository").Code("not_found").
			With("entity", "reservation").
			Public("not found").
			Wrap(bookings.ErrNotFound)
	}
	if err != nil {
		return bookings.Reservation{}, oops.FromContext(ctx).In("bookings.postgres").
			With("operation", "get").
			Wrapf(err, "operation failed")
	}
	item, err := bookingconverter.Reservation(ctx, row)
	return item, oops.Wrapf(err, "map reservation")
}

func (r *Repository) Cancel(ctx context.Context, id uuid.UUID, name string) (err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Cancel), &err)
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
			Wrap(bookings.ErrNotFound)
	}
	return nil
}
