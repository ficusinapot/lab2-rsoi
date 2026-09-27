package converters

import (
	"testing"

	"lab2/gateway/internal/models/entities"
)

func TestLoyaltyConversion(t *testing.T) {
	t.Parallel()
	got := (Loyalty{Status: "GOLD", Discount: 10, ReservationCount: 25}).Entity()
	if got.Status != entities.LoyaltyGold || got.Discount != 10 || got.ReservationCount != 25 {
		t.Fatalf("loyalty: %+v", got)
	}
}
