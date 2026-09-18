package bookings

import (
	"context"
	"uuid"

	"lab2/reservation/internal/domain/bookings"
)

type Repository interface {
	List(context.Context, string) ([]bookings.Reservation, error)
	Get(context.Context, uuid.UUID, string) (bookings.Reservation, error)
	Create(context.Context, bookings.Booking) (bookings.Reservation, error)
	Cancel(context.Context, uuid.UUID, string) error
}

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=repository_mock_test.go -package=bookings Repository
