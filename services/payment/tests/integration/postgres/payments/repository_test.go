//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"uuid"

	model "lab2/payment/internal/models/entities"
	"lab2/payment/internal/postgres/repos"
)

func TestPaymentRepository(t *testing.T) {
	t.Parallel()
	client, other := databaseClients(t)
	repository := repos.New(client)
	id := uuid.New()
	created, err := repository.Create(t.Context(), id, 28500)
	if err != nil || created.Status != model.StatusPaid || created.Price != 28500 {
		t.Fatalf("%+v %v", created, err)
	}
	stored, err := repos.New(other).Get(t.Context(), id)
	if err != nil || stored != created {
		t.Fatalf("%+v %v", stored, err)
	}
	for range 2 {
		if err := repository.Cancel(t.Context(), id); err != nil {
			t.Fatal(err)
		}
	}
	stored, err = repos.New(other).Get(t.Context(), id)
	if err != nil || stored.Status != model.StatusCanceled || stored.Price != 28500 {
		t.Fatalf("%+v %v", stored, err)
	}
	if _, err := repository.Get(t.Context(), uuid.New()); !errors.Is(err, model.ErrNotFound) {
		t.Fatal(err)
	}
	if err := repository.Cancel(t.Context(), uuid.New()); !errors.Is(err, model.ErrNotFound) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := repository.Get(ctx, id); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
