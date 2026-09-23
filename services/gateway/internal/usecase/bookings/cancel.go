package bookings

import (
	"context"

	"github.com/samber/oops"
	"lab2/gateway/internal/domain/bookings"
	"lab2/platform/observability/metrics"
)

func (s *Service) Cancel(ctx context.Context, id, user string) (err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, s.Cancel), &err)
	if err = validUser(user); err != nil {
		return err
	}
	if err = validID(id); err != nil {
		return err
	}
	row, err := s.reservations.Reservation(ctx, id, user)
	if err != nil {
		return oops.Wrapf(err, "get reservation")
	}
	if row.Status == bookings.ReservationCanceled {
		return nil
	}
	if err = s.reservations.CancelReservation(ctx, id, user); err != nil {
		return oops.Wrapf(err, "cancel reservation")
	}
	if err = s.payments.CancelPayment(ctx, row.PaymentUID.String()); err != nil {
		return oops.Wrapf(err, "cancel payment")
	}
	return oops.Wrapf(s.loyalty.ChangeLoyalty(ctx, user, -1), "decrease loyalty")
}
