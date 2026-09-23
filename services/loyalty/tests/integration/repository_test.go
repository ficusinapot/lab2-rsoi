//go:build integration

package integration

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"lab2/loyalty/ent/loyalty"
	model "lab2/loyalty/internal/domain/loyalties"
	storage "lab2/loyalty/internal/storage/postgres/loyalties"
)

func TestLoyaltyRepository(t *testing.T) {
	t.Parallel()
	client, other := databaseClients(t)
	repositories := []*storage.Repository{storage.New(client), storage.New(other)}
	var workers sync.WaitGroup
	results := make(chan error, 40)
	for index := range 20 {
		workers.Go(func() { _, err := repositories[index%2].Get(t.Context(), "concurrent"); results <- err })
	}
	workers.Wait()
	for range 20 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	count, err := client.Loyalty.Query().Where(loyalty.UsernameEQ("concurrent")).Count(t.Context())
	if err != nil || count != 1 {
		t.Fatalf("rows %d: %v", count, err)
	}
	item, err := repositories[0].Get(t.Context(), "concurrent")
	if err != nil || item != model.ForCount(0) {
		t.Fatalf("%+v %v", item, err)
	}
	for expected := 1; expected <= 20; expected++ {
		item, err = repositories[expected%2].Change(t.Context(), "concurrent", 1)
		if err != nil || item != model.ForCount(expected) {
			t.Fatalf("count %d: %+v %v", expected, item, err)
		}
	}
	for expected := 19; expected >= 0; expected-- {
		item, err = repositories[expected%2].Change(t.Context(), "concurrent", -1)
		if err != nil || item != model.ForCount(expected) {
			t.Fatalf("count %d: %+v %v", expected, item, err)
		}
	}
	item, err = repositories[0].Change(t.Context(), "concurrent", -1)
	if err != nil || item != model.ForCount(0) {
		t.Fatalf("%+v %v", item, err)
	}
	for index := range 40 {
		workers.Go(func() { _, err := repositories[index%2].Change(t.Context(), "concurrent", 1); results <- err })
	}
	workers.Wait()
	for range 40 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	item, err = repositories[1].Get(t.Context(), "concurrent")
	if err != nil || item != model.ForCount(40) {
		t.Fatalf("%+v %v", item, err)
	}
	// Mixed updates stay above zero, so the exact final count must be unchanged.
	for index := range 40 {
		delta := 1
		if index%2 == 0 {
			delta = -1
		}
		workers.Go(func() { _, err := repositories[index%2].Change(t.Context(), "concurrent", delta); results <- err })
	}
	workers.Wait()
	for range 40 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	item, err = repositories[0].Get(t.Context(), "concurrent")
	if err != nil || item != model.ForCount(40) {
		t.Fatalf("%+v %v", item, err)
	}
	tx, err := client.Tx(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Loyalty.Query().Where(loyalty.UsernameEQ("concurrent")).ForUpdate().Only(t.Context()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	_, err = repositories[1].Change(ctx, "concurrent", 1)
	cancel()
	rollbackErr := tx.Rollback()
	if rollbackErr != nil {
		t.Fatal(rollbackErr)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	item, err = repositories[0].Get(t.Context(), "concurrent")
	if err != nil || item != model.ForCount(40) {
		t.Fatalf("%+v %v", item, err)
	}
}
