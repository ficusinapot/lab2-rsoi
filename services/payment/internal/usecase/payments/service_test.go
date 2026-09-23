package payments

import (
	"context"
	"errors"
	"math"
	"testing"
	"uuid"

	"github.com/samber/oops"
	"lab2/payment/internal/domain"
	model "lab2/payment/internal/domain/payments"
)

type repositoryStub struct {
	Repository
	created bool
}

func (r *repositoryStub) Create(_ context.Context, id uuid.UUID, price int) (model.Payment, error) {
	r.created = true
	return model.Payment{PaymentUID: id, Price: price, Status: "PAID"}, nil
}

func TestCreateAndValidation(t *testing.T) {
	t.Parallel()
	repo := &repositoryStub{}
	s := New(repo)
	item, err := s.Create(t.Context(), 27000)
	if err != nil || item.PaymentUID == uuid.Nil() || item.Price != 27000 || item.Status != "PAID" {
		t.Fatalf("%+v %v", item, err)
	}
	repo.created = false
	for _, price := range []int{-1, math.MaxInt32 + 1} {
		_, err = s.Create(t.Context(), price)
		if !errors.Is(err, domain.ErrInvalidInput) || repo.created {
			t.Fatalf("price %d: err %v", price, err)
		}
	}
	if _, err = s.Get(t.Context(), "bad"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal(err)
	}
	if err = s.Cancel(t.Context(), "bad"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestRepositoryErrorPreserved(t *testing.T) {
	t.Parallel()
	cause := oops.Errorf("storage failed")
	s := New(&failingRepository{cause: cause})
	_, err := s.Create(t.Context(), 1)
	if !errors.Is(err, cause) {
		t.Fatal(err)
	}
}

type failingRepository struct {
	Repository
	cause error
}

func (r *failingRepository) Create(context.Context, uuid.UUID, int) (model.Payment, error) {
	return model.Payment{}, r.cause
}
