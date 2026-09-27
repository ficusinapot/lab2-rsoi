package bookings

import (
	"encoding/json"
	"testing"
	"uuid"

	"lab2/reservation/internal/models/entities"
)

func TestReservationResponse(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	body, err := json.Marshal(reservationResponse(entities.Reservation{
		ReservationUID: id, Status: entities.StatusPaid, StartDate: "2030-01-01", EndDate: "2030-01-03",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["reservationUid"] != id.String() || got["status"] != "PAID" ||
		got["startDate"] != "2030-01-01" || got["endDate"] != "2030-01-03" {
		t.Fatalf("response: %s", body)
	}
}
