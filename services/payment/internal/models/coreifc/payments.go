package coreifc

import (
	"context"

	"lab2/payment/internal/models/entities"
)

type Payments interface {
	Create(context.Context, int) (entities.Payment, error)
	Get(context.Context, string) (entities.Payment, error)
	Cancel(context.Context, string) error
}
