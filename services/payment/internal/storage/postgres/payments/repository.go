package payments

import (
	"context"
	"uuid"

	"github.com/samber/oops"
	"lab2/payment/ent"
	"lab2/payment/ent/payment"
	"lab2/payment/internal/domain"
	model "lab2/payment/internal/domain/payments"
	"lab2/platform/observability/metrics"
)

type Repository struct{ client *ent.Client }

func New(client *ent.Client) *Repository { return &Repository{client: client} }
func modelOf(row *ent.Payment) model.Payment {
	return model.Payment{PaymentUID: row.PaymentUID, Status: model.Status(row.Status), Price: row.Price}
}

func (r *Repository) Create(ctx context.Context, id uuid.UUID, price int) (_ model.Payment, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Create), &err)
	row, err := r.client.Payment.Create().SetPaymentUID(id).SetStatus(payment.StatusPAID).SetPrice(price).Save(ctx)
	if err != nil {
		return model.Payment{}, oops.Wrapf(err, "create payment")
	}
	return modelOf(row), nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (_ model.Payment, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Get), &err)
	row, err := r.client.Payment.Query().Where(payment.PaymentUIDEQ(id)).Only(ctx)
	if ent.IsNotFound(err) {
		return model.Payment{}, oops.Wrap(domain.ErrNotFound)
	}
	if err != nil {
		return model.Payment{}, oops.Wrapf(err, "get payment")
	}
	return modelOf(row), nil
}

func (r *Repository) Cancel(ctx context.Context, id uuid.UUID) (err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Cancel), &err)
	count, err := r.client.Payment.Update().Where(payment.PaymentUIDEQ(id)).SetStatus(payment.StatusCANCELED).Save(ctx)
	if err != nil {
		return oops.Wrapf(err, "cancel payment")
	}
	if count == 0 {
		return oops.Wrap(domain.ErrNotFound)
	}
	return nil
}
