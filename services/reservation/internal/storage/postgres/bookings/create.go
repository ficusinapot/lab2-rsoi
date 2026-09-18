package bookings

import (
	"context"

	"github.com/samber/oops"

	"github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

	"lab2/reservation/ent"
	"lab2/reservation/ent/hotel"
	"lab2/reservation/ent/reservation"
	"lab2/reservation/internal/domain"
	"lab2/reservation/internal/domain/bookings"
)

func (r *Repository) Create(ctx context.Context, booking bookings.Booking) (_ bookings.Reservation, err error) {
	defer autometrics.Instrument(autometrics.PreInstrument(ctx), &err)
	h, err := r.client.Hotel.Query().Where(hotel.HotelUIDEQ(booking.HotelUID)).Only(ctx)
	if ent.IsNotFound(err) {
		return bookings.Reservation{}, oops.In("repository").Code("not_found").
			With("entity", "hotel").
			Public("not found").
			Wrap(domain.ErrNotFound)
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
	return model(ctx, row)
}
