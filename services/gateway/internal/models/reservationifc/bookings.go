package reservationifc

import (
	"context"

	"lab2/gateway/internal/models/entities"
)

type Bookings interface {
	Hotels(context.Context, entities.PageQuery) (entities.Page, error)
	Hotel(context.Context, string) (entities.Hotel, error)
	Reservations(context.Context, string) ([]entities.InternalReservation, error)
	Reservation(context.Context, string, string) (entities.InternalReservation, error)
	CreateReservation(context.Context, string, entities.InternalCreate) (entities.InternalReservation, error)
	CancelReservation(context.Context, string, string) error
}
