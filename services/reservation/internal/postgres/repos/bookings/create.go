package bookings

import (
	"context"

	"github.com/samber/oops"

	"lab2/platform/observability/metrics"

	"lab2/reservation/ent"
	"lab2/reservation/ent/hotel"
	"lab2/reservation/ent/reservation"
	bookings "lab2/reservation/internal/models/entities"
	bookingconverter "lab2/reservation/internal/postgres/converters/bookings"
)

func (r *Repository) Create(ctx context.Context, booking bookings.Booking) (_ bookings.Reservation, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Create), &err)
	h, err := r.client.Hotel.Query().Where(hotel.HotelUIDEQ(booking.HotelUID)).Only(ctx)
	if ent.IsNotFound(err) {
		return bookings.Reservation{}, oops.In("repository").Code("not_found").
			With("entity", "hotel").
			Public("not found").
			Wrap(bookings.ErrNotFound)
	}
	if err != nil {
		return bookings.Reservation{}, oops.FromContext(ctx).In("bookings.postgres").
			With("operation", "get_hotel").
			Wrapf(err, "operation failed")
	}
	row, err := r.client.Reservation.Create().SetReservationUID(booking.ReservationUID).SetUsername(booking.Username).
		SetPaymentUID(booking.PaymentUID).SetHotelID(h.ID).SetStatus(reservation.Status(bookings.StatusPaid)).
		SetStartDate(booking.StartDate).SetEndDate(booking.EndDate).Save(ctx)
	if err != nil {
		return bookings.Reservation{}, oops.FromContext(ctx).In("bookings.postgres").
			With("operation", "create").
			Wrapf(err, "operation failed")
	}
	row.Edges.Hotel = h
	item, err := bookingconverter.Reservation(ctx, row)
	return item, oops.Wrapf(err, "map reservation")
}
