//go:build e2e

package e2e

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"

	"lab2/gateway/internal/domain/bookings"
)

func partialFailures(t *testing.T, s *stack) {
	t.Helper()
	input := bookings.CreateRequest{HotelUID: hotelUID, StartDate: "2026-01-01", EndDate: "2026-01-04"}
	user := "partial failure"
	var created bookings.Created
	request(t, http.MethodPost, s.gateway+"/api/v1/reservations", user, input, &created, http.StatusOK)
	loyalty(t, s, user, 1, 5, bookings.LoyaltyBronze)
	// Failing Reservation cancellation must leave Payment and Loyalty untouched.
	restore := stopService(t, s, "reservation")
	request(t, http.MethodDelete, s.gateway+"/api/v1/reservations/"+created.ReservationUID.String(), user,
		nil, nil, http.StatusBadGateway)
	restore()
	var item bookings.Reservation
	request(t, http.MethodGet, s.gateway+"/api/v1/reservations/"+created.ReservationUID.String(), user,
		nil, &item, http.StatusOK)
	if item.Status != bookings.ReservationPaid || item.Payment.Status != bookings.PaymentPaid {
		t.Fatal(item)
	}
	loyalty(t, s, user, 1, 5, bookings.LoyaltyBronze)
	// Payment failure happens after Reservation cancellation; no rollback and no decrement.
	restore = stopService(t, s, "payment")
	request(t, http.MethodDelete, s.gateway+"/api/v1/reservations/"+created.ReservationUID.String(), user,
		nil, nil, http.StatusBadGateway)
	restore()
	request(t, http.MethodGet, s.gateway+"/api/v1/reservations/"+created.ReservationUID.String(), user,
		nil, &item, http.StatusOK)
	if item.Status != bookings.ReservationCanceled || item.Payment.Status != bookings.PaymentPaid {
		t.Fatal(item)
	}
	loyalty(t, s, user, 1, 5, bookings.LoyaltyBronze)
	// A failed creation stops before creating a Reservation or incrementing Loyalty.
	before := reservationCount(t, s, user)
	restore = stopService(t, s, "payment")
	request(t, http.MethodPost, s.gateway+"/api/v1/reservations", user, input, nil, http.StatusBadGateway)
	restore()
	if after := reservationCount(t, s, user); before != after {
		t.Fatalf("reservations %d -> %d", before, after)
	}
	loyalty(t, s, user, 1, 5, bookings.LoyaltyBronze)
}

func reservationCount(t *testing.T, s *stack, user string) int {
	t.Helper()
	var count int
	query := s.db.QueryRowContext(t.Context(), "SELECT count(*) FROM reservation WHERE username=$1", user)
	if err := query.Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func stopService(t *testing.T, s *stack, name string) func() {
	t.Helper()
	timeout := time.Second
	container := s.containers[name]
	if err := container.Stop(t.Context(), &timeout); err != nil {
		t.Fatal(err)
	}
	restarted := false
	restore := func() {
		if restarted {
			return
		}
		restarted = true
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := container.Start(ctx); err != nil {
			t.Fatal(err)
		}
		// Docker may allocate a new host port when restarting a container.
		host, err := container.Host(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ports := map[string]nat.Port{"reservation": "8070/tcp", "payment": "8060/tcp", loyaltyServiceName: "8050/tcp"}
		port, err := container.MappedPort(ctx, ports[name])
		if err != nil {
			t.Fatal(err)
		}
		s.addresses[name] = "http://" + net.JoinHostPort(host, port.Port())
		client := &http.Client{Timeout: time.Second}
		defer client.CloseIdleConnections()
		for ctx.Err() == nil {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.addresses[name]+"/manage/health", http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			res, err := client.Do(req)
			if err == nil {
				if err := res.Body.Close(); err != nil {
					t.Fatal(err)
				}
				if res.StatusCode == http.StatusOK {
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal(ctx.Err())
	}
	t.Cleanup(restore)
	return restore
}
