package converters

import (
	"testing"

	"lab2/loyalty/ent"
	entloyalty "lab2/loyalty/ent/loyalty"
	"lab2/loyalty/internal/models/entities"
)

func TestLoyalty(t *testing.T) {
	t.Parallel()
	got := Loyalty(&ent.Loyalty{Status: entloyalty.StatusGOLD, Discount: 10, ReservationCount: 25})
	if got.Status != entities.StatusGold || got.Discount != 10 || got.ReservationCount != 25 {
		t.Fatalf("loyalty: %+v", got)
	}
}
