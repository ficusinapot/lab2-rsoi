package payments

import (
	"context"
	"errors"
	"math"
	"uuid"

	"github.com/samber/oops"
	"lab2/payment/internal/domain"
	"lab2/payment/internal/domain/payments"
)

type Repository interface {
	Create(context.Context, uuid.UUID, int) (payments.Payment, error)
	Get(context.Context, uuid.UUID) (payments.Payment, error)
	Cancel(context.Context, uuid.UUID) error
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Create(ctx context.Context, price int) (payments.Payment, error) {
	if price < 0 || price > math.MaxInt32 {
		return payments.Payment{}, oops.Wrap(domain.ErrInvalidInput)
	}
	item, err := s.repo.Create(ctx, uuid.New(), price)
	return item, oops.Wrapf(err, "create payment")
}

func parse(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil(), oops.Wrap(errors.Join(domain.ErrInvalidInput, err))
	}
	return uid, nil
}

func (s *Service) Get(ctx context.Context, id string) (payments.Payment, error) {
	uid, err := parse(id)
	if err != nil {
		return payments.Payment{}, err
	}
	item, err := s.repo.Get(ctx, uid)
	return item, oops.Wrapf(err, "get payment")
}

func (s *Service) Cancel(ctx context.Context, id string) error {
	uid, err := parse(id)
	if err != nil {
		return err
	}
	return oops.Wrapf(s.repo.Cancel(ctx, uid), "cancel payment")
}
