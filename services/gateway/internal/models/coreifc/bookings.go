package coreifc

import (
	"context"

	"lab2/gateway/internal/models/entities"
)

type Bookings interface {
	Hotels(context.Context, string, string) (entities.Page, error)
	List(context.Context, string) ([]entities.Reservation, error)
	Get(context.Context, string, string) (entities.Reservation, error)
	Me(context.Context, string) (entities.UserInfo, error)
	Loyalty(context.Context, string) (entities.Loyalty, error)
	Create(context.Context, string, entities.CreateRequest) (entities.Created, error)
	Cancel(context.Context, string, string) error
}
