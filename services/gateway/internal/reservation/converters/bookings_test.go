package converters

import (
	"testing"
	"uuid"

	"lab2/gateway/internal/models/entities"
)

func TestBookingConversions(t *testing.T) {
	t.Parallel()
	hotelID := uuid.New()
	paymentID := uuid.New()
	stars := 5
	page := Page{Page: 1, PageSize: 10, TotalElements: 1, Items: []Hotel{{
		HotelUID: hotelID, Name: "Hotel", Stars: &stars, Price: 100,
	}}}.Entity()
	if page.Page != 1 || page.TotalElements != 1 || len(page.Items) != 1 ||
		page.Items[0].HotelUID != hotelID || *page.Items[0].Stars != stars {
		t.Fatalf("page: %+v", page)
	}
	row := Reservation{HotelUID: hotelID, PaymentUID: paymentID, Status: "PAID"}.Entity()
	if row.HotelUID != hotelID || row.PaymentUID != paymentID || row.Status != entities.ReservationPaid {
		t.Fatalf("reservation: %+v", row)
	}
	request := NewCreateRequest(entities.InternalCreate{
		HotelUID: hotelID, StartDate: "2026-01-01", EndDate: "2026-01-02",
		PaymentUID: paymentID,
	})
	if request.HotelUID != hotelID || request.PaymentUID != paymentID ||
		request.StartDate != "2026-01-01" || request.EndDate != "2026-01-02" {
		t.Fatalf("request: %+v", request)
	}
}
