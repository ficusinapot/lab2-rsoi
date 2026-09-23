package loyalties

import (
	"context"
	"errors"
	"math"

	"lab2/loyalty/ent"
	"lab2/loyalty/ent/loyalty"
	"lab2/loyalty/internal/domain"
	model "lab2/loyalty/internal/domain/loyalties"
	"lab2/platform/observability/metrics"

	"github.com/samber/oops"
)

type Repository struct{ client *ent.Client }

func New(client *ent.Client) *Repository { return &Repository{client: client} }
func (r *Repository) ensure(ctx context.Context, name string) error {
	return oops.Wrapf(r.client.Loyalty.Create().SetUsername(name).SetDiscount(5).
		OnConflictColumns(loyalty.FieldUsername).Ignore().Exec(ctx), "initialize loyalty")
}

func modelOf(row *ent.Loyalty) model.Loyalty {
	return model.Loyalty{Status: model.Status(row.Status), Discount: row.Discount, ReservationCount: row.ReservationCount}
}

func (r *Repository) Get(ctx context.Context, name string) (_ model.Loyalty, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Get), &err)
	if err = r.ensure(ctx, name); err != nil {
		return model.Loyalty{}, err
	}
	row, err := r.client.Loyalty.Query().Where(loyalty.UsernameEQ(name)).Only(ctx)
	if err != nil {
		return model.Loyalty{}, oops.Wrapf(err, "get loyalty")
	}
	return modelOf(row), nil
}

func (r *Repository) Change(ctx context.Context, name string, delta int) (_ model.Loyalty, err error) {
	defer metrics.Instrument(metrics.PreInstrument(ctx, r.Change), &err)
	if err = r.ensure(ctx, name); err != nil {
		return model.Loyalty{}, err
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return model.Loyalty{}, oops.Wrapf(err, "begin loyalty change")
	}
	defer func() {
		if err != nil {
			err = oops.Wrap(errors.Join(err, tx.Rollback()))
		}
	}()
	row, err := tx.Loyalty.Query().Where(loyalty.UsernameEQ(name)).ForUpdate().Only(ctx)
	if err != nil {
		return model.Loyalty{}, oops.Wrapf(err, "lock loyalty")
	}
	if row.ReservationCount == math.MaxInt32 && delta > 0 {
		return model.Loyalty{}, oops.Wrap(domain.ErrInvalidInput)
	}
	count := max(0, row.ReservationCount+delta)
	item := model.ForCount(count)
	_, err = tx.Loyalty.UpdateOneID(row.ID).SetReservationCount(count).
		SetStatus(loyalty.Status(item.Status)).SetDiscount(item.Discount).Save(ctx)
	if err != nil {
		return model.Loyalty{}, oops.Wrapf(err, "update loyalty")
	}
	if err = tx.Commit(); err != nil {
		return model.Loyalty{}, oops.Wrapf(err, "commit loyalty change")
	}
	return item, nil
}
