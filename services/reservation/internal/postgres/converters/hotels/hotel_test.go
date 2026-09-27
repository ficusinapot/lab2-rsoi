package hotels

import (
	"testing"
	"uuid"

	"lab2/reservation/ent"
)

func TestHotel(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	stars := 5
	got := Hotel(&ent.Hotel{HotelUID: id, Name: "Hotel", Stars: &stars, Price: 100})
	if got.HotelUID != id || got.Name != "Hotel" || got.Stars == nil || *got.Stars != stars || got.Price != 100 {
		t.Fatalf("hotel: %+v", got)
	}
}
