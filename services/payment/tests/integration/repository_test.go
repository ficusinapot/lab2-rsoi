//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"lab2/payment/internal/domain"
	model "lab2/payment/internal/domain/payments"
	"lab2/payment/internal/storage/postgres/payments"
)

func TestPaymentRepository(t *testing.T) {
	t.Parallel()
	client, other := databaseClients(t)
	repository := payments.New(client)
	id := uuid.New()
	created, err := repository.Create(t.Context(), id, 28500)
	if err != nil || created.Status != model.StatusPaid || created.Price != 28500 {
		t.Fatalf("%+v %v", created, err)
	}
	stored, err := payments.New(other).Get(t.Context(), id)
	if err != nil || stored != created {
		t.Fatalf("%+v %v", stored, err)
	}
	for range 2 {
		if err := repository.Cancel(t.Context(), id); err != nil {
			t.Fatal(err)
		}
	}
	stored, err = payments.New(other).Get(t.Context(), id)
	if err != nil || stored.Status != model.StatusCanceled || stored.Price != 28500 {
		t.Fatalf("%+v %v", stored, err)
	}
	if _, err := repository.Get(t.Context(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	if err := repository.Cancel(t.Context(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := repository.Get(ctx, id); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
