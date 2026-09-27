package bookings

import (
	"testing"
	"time"
	"uuid"

	"lab2/reservation/ent"
	entreservation "lab2/reservation/ent/reservation"
	"lab2/reservation/internal/models/entities"
)

func TestReservation(t *testing.T) {
	t.Parallel()
	hotelID := uuid.New()
	paymentID := uuid.New()
	start := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 2)
	row := &ent.Reservation{
		ReservationUID: uuid.New(), PaymentUID: paymentID, Status: entreservation.StatusPAID,
		StartDate: &start, EndDate: &end, Edges: ent.ReservationEdges{Hotel: &ent.Hotel{HotelUID: hotelID}},
	}
	got, err := Reservation(t.Context(), row)
	if err != nil || got.HotelUID != hotelID || got.PaymentUID != paymentID ||
		got.Status != entities.StatusPaid || got.StartDate != "2030-01-01" || got.EndDate != "2030-01-03" {
		t.Fatalf("reservation: %+v %v", got, err)
	}
	row.EndDate = nil
	if _, err := Reservation(t.Context(), row); err == nil {
		t.Fatal("missing date accepted")
	}
}
