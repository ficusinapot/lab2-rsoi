package hotels

import (
	"context"

	"lab2/reservation/internal/domain/hotels"
)

type Service interface {
	List(context.Context, string, string) (hotels.Page, error)
	Get(context.Context, string) (hotels.Hotel, error)
}
