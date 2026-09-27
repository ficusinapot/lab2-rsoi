package coreifc

import (
	"context"

	"lab2/reservation/internal/models/entities"
)

type Hotels interface {
	List(context.Context, string, string) (entities.Page, error)
	Get(context.Context, string) (entities.Hotel, error)
}
