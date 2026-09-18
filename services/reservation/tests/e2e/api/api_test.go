//go:build e2e

package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"uuid"

	bookinghttp "lab2/reservation/internal/transport/http/bookings"
	"lab2/reservation/tests/e2e/fixtures"
	"lab2/reservation/tests/e2e/infrastructure"
	"lab2/reservation/tests/e2e/support"
)

//nolint:paralleltest // The test builds and starts a dedicated container stack.
func TestReservationAPI(t *testing.T) {
	stack := infrastructure.NewReservationStack(t)
	fixtures.SeedHotel(t, stack.DatabaseURL)
	api := support.NewHTTPClient(stack.BaseURL)

	api.Request(t, http.MethodGet, "/manage/health", "", nil, http.StatusOK)
	api.Request(t, http.MethodGet, "/api/v1/hotels", "", nil, http.StatusOK)
	api.Request(t, http.MethodGet, "/api/v1/hotels?page=0", "", nil, http.StatusBadRequest)
	api.Request(t, http.MethodGet, "/api/v1/docs", "", nil, http.StatusOK)
	api.Request(t, http.MethodGet, "/api/v1/openapi.json", "", nil, http.StatusOK)

	body, err := json.Marshal(bookinghttp.CreateRequest{
		HotelUID:   fixtures.TestHotelUID,
		PaymentUID: uuid.MustParse("00000000-0000-0000-0000-000000000001").String(),
		StartDate:  "2030-01-01",
		EndDate:    "2030-01-03",
	})
	if err != nil {
		t.Fatal(err)
	}
	created := api.Request(t, http.MethodPost, "/api/v1/reservations", "e2e-user", body, http.StatusCreated)
	var reservation struct {
		ReservationUID string `json:"reservationUid"`
		Status         string `json:"status"`
	}
	if err := json.Unmarshal(created, &reservation); err != nil {
		t.Fatal(err)
	}
	if reservation.ReservationUID == "" || reservation.Status != "PAID" {
		t.Fatalf("unexpected reservation: %s", created)
	}

	path := "/api/v1/reservations/" + reservation.ReservationUID
	api.Request(t, http.MethodGet, path, "other-user", nil, http.StatusNotFound)
	api.Request(t, http.MethodGet, path, "e2e-user", nil, http.StatusOK)
	api.Request(t, http.MethodDelete, path, "e2e-user", nil, http.StatusNoContent)
	canceled := api.Request(t, http.MethodGet, path, "e2e-user", nil, http.StatusOK)
	if !containsStatus(canceled, "CANCELED") {
		t.Fatalf("reservation is not canceled: %s", canceled)
	}
}

func containsStatus(body []byte, status string) bool {
	var reservation struct {
		Status string `json:"status"`
	}
	return json.Unmarshal(body, &reservation) == nil && reservation.Status == status
}
