package entities

import "testing"

func TestTierBoundaries(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		count, discount int
		status          Status
	}{
		{0, 5, "BRONZE"}, {9, 5, "BRONZE"}, {10, 7, "SILVER"}, {19, 7, "SILVER"}, {20, 10, "GOLD"},
	} {
		got := ForCount(tc.count)
		if got.Status != tc.status || got.Discount != tc.discount || got.ReservationCount != tc.count {
			t.Fatalf("%+v: %+v", tc, got)
		}
	}
}
