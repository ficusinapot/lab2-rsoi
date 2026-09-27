//go:build integration

package hotels_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"lab2/reservation/ent/hotel"
	"lab2/reservation/internal/models/entities"
	hotelrepo "lab2/reservation/internal/postgres/repos/hotels"
	"lab2/reservation/tests/integration/infrastructure"
)

func TestHotelRepository(t *testing.T) {
	db := infrastructure.NewDatabase(t)
	id := uuid.New()
	_, err := db.Client.Hotel.Create().
		SetHotelUID(id).SetName("Repository hotel").SetCountry("Russia").
		SetCity("Moscow").SetAddress("Street").SetPrice(1000).Save(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Client.Hotel.Delete().Where(hotel.HotelUIDEQ(id)).Exec(context.WithoutCancel(t.Context())); err != nil {
			t.Error(err)
		}
	})
	repo := hotelrepo.New(db.Client)
	got, err := repo.Get(t.Context(), id)
	if err != nil || got.HotelUID != id || got.Price != 1000 {
		t.Fatalf("get: %+v %v", got, err)
	}
	page, err := repo.List(t.Context(), 1, 100)
	if err != nil || page.TotalElements < 1 || page.Page != 1 {
		t.Fatalf("list: %+v %v", page, err)
	}
	found := false
	for _, item := range page.Items {
		found = found || item.HotelUID == id
	}
	if !found {
		t.Fatalf("hotel not found in page: %+v", page)
	}
	if _, err := repo.Get(t.Context(), uuid.New()); !errors.Is(err, entities.ErrNotFound) {
		t.Fatalf("missing hotel: %v", err)
	}
}
