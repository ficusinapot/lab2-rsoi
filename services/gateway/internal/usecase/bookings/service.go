package bookings

import (
	"context"
	"strings"
	"unicode/utf8"
	"uuid"

	"github.com/samber/oops"
	"lab2/gateway/internal/domain"
	"lab2/gateway/internal/domain/bookings"
)

type ReservationService interface {
	Hotels(context.Context, string) (bookings.Page, error)
	Hotel(context.Context, string) (bookings.Hotel, error)
	Reservations(context.Context, string) ([]bookings.InternalReservation, error)
	Reservation(context.Context, string, string) (bookings.InternalReservation, error)
	CreateReservation(context.Context, string, bookings.InternalCreate) (bookings.InternalReservation, error)
	CancelReservation(context.Context, string, string) error
}
type PaymentService interface {
	Payment(context.Context, string) (bookings.Payment, error)
	CreatePayment(context.Context, int) (bookings.InternalPayment, error)
	CancelPayment(context.Context, string) error
}
type LoyaltyService interface {
	Loyalty(context.Context, string) (bookings.Loyalty, error)
	ChangeLoyalty(context.Context, string, int) error
}
type Service struct {
	reservations ReservationService
	payments     PaymentService
	loyalty      LoyaltyService
	concurrency  int
}

func New(reservations ReservationService, payments PaymentService, loyalty LoyaltyService, concurrency int) *Service {
	return &Service{reservations: reservations, payments: payments, loyalty: loyalty, concurrency: concurrency}
}

func validUser(name string) error {
	if strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return oops.Wrap(domain.ErrInvalidInput)
	}
	return nil
}

func validID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return oops.Wrap(domain.ErrInvalidInput)
	}
	return nil
}
