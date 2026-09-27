package paymentifc

import (
	"context"

	"lab2/gateway/internal/models/entities"
)

type Payments interface {
	Payment(context.Context, string) (entities.Payment, error)
	CreatePayment(context.Context, int) (entities.InternalPayment, error)
	CancelPayment(context.Context, string) error
}
