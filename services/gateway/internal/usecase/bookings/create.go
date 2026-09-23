package bookings

import (
	"context"

	"github.com/samber/oops"
	"lab2/gateway/internal/domain/bookings"
	"lab2/platform/observability/metrics"
)

func (s *Service) Create(
	ctx context.Context, user string, input bookings.CreateRequest,
) (_ bookings.Created, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, s.Create), &err)
	if err = validUser(user); err != nil {
		return bookings.Created{}, err
	}
	if err = validID(input.HotelUID); err != nil {
		return bookings.Created{}, err
	}
	nights, err := dates(input)
	if err != nil {
		return bookings.Created{}, err
	}
	hotel, err := s.reservations.Hotel(ctx, input.HotelUID)
	if err != nil {
		return bookings.Created{}, oops.Wrapf(err, "get hotel")
	}
	loyalty, err := s.loyalty.Loyalty(ctx, user)
	if err != nil {
		return bookings.Created{}, oops.Wrapf(err, "get loyalty")
	}
	price, err := cost(nights, hotel.Price, loyalty.Discount)
	if err != nil {
		return bookings.Created{}, err
	}
	payment, err := s.payments.CreatePayment(ctx, price)
	if err != nil {
		return bookings.Created{}, oops.Wrapf(err, "create payment")
	}
	row, err := s.reservations.CreateReservation(ctx, user, bookings.InternalCreate{
		CreateRequest: input, PaymentUID: payment.PaymentUID.String(),
	})
	if err != nil {
		return bookings.Created{}, oops.Wrapf(err, "create reservation")
	}
	if err = s.loyalty.ChangeLoyalty(ctx, user, 1); err != nil {
		return bookings.Created{}, oops.Wrapf(err, "increase loyalty")
	}
	return bookings.Created{
		ReservationUID: row.ReservationUID, HotelUID: row.HotelUID, StartDate: row.StartDate, EndDate: row.EndDate,
		Status: row.Status, Discount: loyalty.Discount, Payment: payment.Payment,
	}, nil
}
