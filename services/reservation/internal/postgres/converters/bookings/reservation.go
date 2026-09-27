package bookings

import (
	"context"
	"time"

	"github.com/samber/oops"
	"lab2/reservation/ent"
	bookings "lab2/reservation/internal/models/entities"
)

func Reservation(ctx context.Context, r *ent.Reservation) (bookings.Reservation, error) {
	h, err := r.Edges.HotelOrErr()
	if err != nil {
		return bookings.Reservation{}, oops.FromContext(ctx).In("bookings.postgres").
			With("operation", "load_hotel").
			Wrapf(err, "operation failed")
	}
	if r.StartDate == nil || r.EndDate == nil {
		return bookings.Reservation{}, oops.In("bookings.postgres").Code("invalid_stored_booking").
			Errorf("reservation dates are missing")
	}
	return bookings.Reservation{
		ReservationUID: r.ReservationUID, HotelUID: h.HotelUID,
		PaymentUID: r.PaymentUID, Status: bookings.Status(r.Status),
		StartDate: r.StartDate.UTC().Format(time.DateOnly), EndDate: r.EndDate.UTC().Format(time.DateOnly),
	}, nil
}
