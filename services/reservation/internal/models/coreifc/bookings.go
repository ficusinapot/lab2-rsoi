package coreifc

import (
	"context"

	"lab2/reservation/internal/models/entities"
)

type Bookings interface {
	List(context.Context, string) ([]entities.Reservation, error)
	Get(context.Context, string, string) (entities.Reservation, error)
	Create(context.Context, string, entities.CreateInput) (entities.Reservation, error)
	Cancel(context.Context, string, string) error
}
