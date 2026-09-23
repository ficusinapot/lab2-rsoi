//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"lab2/gateway/internal/domain/bookings"
)

const hotelUID = "049161bb-badd-4fa8-9d90-87c9a82b0668"

func request(t *testing.T, method, address, user string, input, output any, want int) {
	t.Helper()
	var data bytes.Buffer
	if input != nil {
		if err := json.NewEncoder(&data).Encode(input); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequestWithContext(t.Context(), method, address, &data)
	if err != nil {
		t.Fatal(err)
	}
	if user != "" {
		req.Header.Set("X-User-Name", user)
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Error(err)
		}
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d want %d body %s", method, address, res.StatusCode, want, body)
	}
	if strings.Contains(address, "/api/v1/reservations") && strings.Contains(string(body), "paymentUid") {
		t.Fatal("internal payment UID leaked")
	}
	if output != nil {
		if err := json.Unmarshal(body, output); err != nil {
			t.Fatal(err)
		}
	}
}

func loyalty(t *testing.T, s *stack, user string, count, discount int, status bookings.LoyaltyStatus) {
	t.Helper()
	var out bookings.Loyalty
	request(t, http.MethodGet, s.gateway+"/api/v1/loyalty", user, nil, &out, http.StatusOK)
	if out.ReservationCount != count || out.Discount != discount || out.Status != status {
		t.Fatalf("unexpected loyalty %+v", out)
	}
}

func TestBookingLifecycle(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	for _, address := range s.urls {
		for _, path := range []string{"/manage/health", "/metrics", "/api/v1/docs", "/api/v1/openapi.json"} {
			request(t, http.MethodGet, address+path, "", nil, nil, http.StatusOK)
		}
	}
	var hotels bookings.Page
	request(t, http.MethodGet, s.gateway+"/api/v1/hotels?page=1&size=1", "", nil, &hotels, http.StatusOK)
	if len(hotels.Items) != 1 || hotels.Items[0].HotelUID.String() != hotelUID {
		t.Fatalf("wrong hotels %+v", hotels)
	}
	loyalty(t, s, "new user", 0, 5, "BRONZE")
	input := bookings.CreateRequest{HotelUID: hotelUID, StartDate: "2026-01-01", EndDate: "2026-01-04"}
	ids := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		var out bookings.Created
		request(t, http.MethodPost, s.gateway+"/api/v1/reservations", "new user", input, &out, http.StatusOK)
		discount := 5
		if i >= 10 {
			discount = 7
		}
		if out.Discount != discount || out.Payment.Price != 30000*(100-discount)/100 || out.Status != "PAID" {
			t.Fatalf("booking %d: %+v", i, out)
		}
		ids = append(ids, out.ReservationUID.String())
		if i == 8 {
			loyalty(t, s, "new user", 9, 5, "BRONZE")
		}
		if i == 9 {
			loyalty(t, s, "new user", 10, 7, "SILVER")
		}
		if i == 18 {
			loyalty(t, s, "new user", 19, 7, "SILVER")
		}
	}
	loyalty(t, s, "new user", 20, 10, "GOLD")
	path := s.gateway + "/api/v1/reservations/" + ids[19]
	request(t, http.MethodGet, path, "stranger", nil, nil, http.StatusNotFound)
	request(t, http.MethodDelete, path, "stranger", nil, nil, http.StatusNotFound)
	loyalty(t, s, "new user", 20, 10, "GOLD")
	request(t, http.MethodDelete, path, "new user", nil, nil, http.StatusNoContent)
	loyalty(t, s, "new user", 19, 7, "SILVER")
	var item bookings.Reservation
	request(t, http.MethodGet, path, "new user", nil, &item, http.StatusOK)
	if item.Status != "CANCELED" || item.Payment.Status != "CANCELED" {
		t.Fatalf("cancel %+v", item)
	}
	if item.Hotel.FullAddress != "Россия, Москва, Неглинная ул., 4" {
		t.Fatal(item.Hotel.FullAddress)
	}
	request(t, http.MethodDelete, path, "new user", nil, nil, http.StatusNoContent)
	loyalty(t, s, "new user", 19, 7, "SILVER")
	for i := 18; i >= 9; i-- {
		request(t, http.MethodDelete, s.gateway+"/api/v1/reservations/"+ids[i], "new user", nil, nil, http.StatusNoContent)
	}
	loyalty(t, s, "new user", 9, 5, "BRONZE")
	var me bookings.UserInfo
	request(t, http.MethodGet, s.gateway+"/api/v1/me", "new user", nil, &me, http.StatusOK)
	if len(me.Reservations) != 20 || me.Loyalty.ReservationCount != 9 {
		t.Fatalf("me %+v", me)
	}
	var list []bookings.Reservation
	request(t, http.MethodGet, s.gateway+"/api/v1/reservations", "new user", nil, &list, http.StatusOK)
	if len(list) != 20 {
		t.Fatal(len(list))
	}
	input.EndDate = input.StartDate
	request(t, http.MethodPost, s.gateway+"/api/v1/reservations", "new user", input, nil, http.StatusBadRequest)
	loyalty(t, s, "new user", 9, 5, "BRONZE")
	request(t, http.MethodGet, s.gateway+"/api/v1/loyalty", "", nil, nil, http.StatusBadRequest)
	// Exercise concurrent first access and changes at the owning service boundary.
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() {
			requestConcurrent(t, s.loyalty+"/api/v1/loyalty/reservations")
		})
	}
	wg.Wait()
	loyalty(t, s, "concurrent user", 20, 10, "GOLD")
	request(t, http.MethodDelete, s.loyalty+"/api/v1/loyalty/reservations", "zero user", nil, nil, http.StatusOK)
	loyalty(t, s, "zero user", 0, 5, "BRONZE")
	partialFailures(t, s)
}

func requestConcurrent(t *testing.T, address string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, address, http.NoBody)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("X-User-Name", "concurrent user")
	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err := io.Copy(io.Discard, res.Body); err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("concurrent change: %d", res.StatusCode)
	}
}
