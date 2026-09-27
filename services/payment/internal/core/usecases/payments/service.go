package payments

import (
	"context"
	"errors"
	"math"
	"uuid"

	"github.com/samber/oops"
	"lab2/payment/internal/models/dbifc"
	payments "lab2/payment/internal/models/entities"
)

type Service struct{ repo dbifc.Payments }

func New(repo dbifc.Payments) *Service { return &Service{repo: repo} }
func (s *Service) Create(ctx context.Context, price int) (payments.Payment, error) {
	if price < 0 || price > math.MaxInt32 {
		return payments.Payment{}, oops.Wrap(payments.ErrInvalidInput)
	}
	item, err := s.repo.Create(ctx, uuid.New(), price)
	return item, oops.Wrapf(err, "create payment")
}

func parse(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil(), oops.Wrap(errors.Join(payments.ErrInvalidInput, err))
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
