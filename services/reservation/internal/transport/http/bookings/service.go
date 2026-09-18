package bookings

import (
	"context"

	"lab2/reservation/internal/domain/bookings"
	usecase "lab2/reservation/internal/usecase/bookings"
)

type Service interface {
	List(context.Context, string) ([]bookings.Reservation, error)
	Get(context.Context, string, string) (bookings.Reservation, error)
	Create(context.Context, string, usecase.CreateInput) (bookings.Reservation, error)
	Cancel(context.Context, string, string) error
}
