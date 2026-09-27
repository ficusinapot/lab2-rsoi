//go:build integration

package bookings

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/samber/oops"
	"lab2/gateway/internal/httpclient"
	loyaltyrepo "lab2/gateway/internal/loyalty/repos"
	"lab2/gateway/internal/models/entities"
)

func TestHTTPClient(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-Name") != "user" {
			t.Error("missing user header")
		}
		if r.URL.Path != "/loyalty" {
			t.Error(r.URL.Path)
		}
		if _, err := w.Write([]byte(`{"status":"BRONZE","discount":5,"reservationCount":0}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	c := newLoyalty(t, server.URL, time.Second)
	item, err := c.Loyalty(t.Context(), "user")
	if err != nil || item.Discount != 5 || item.ReservationCount != 0 {
		t.Fatalf("%+v %v", item, err)
	}
}

func TestUpstreamErrors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		code int
		want error
		body string
	}{
		{http.StatusBadRequest, entities.ErrInvalidInput, ""},
		{http.StatusUnprocessableEntity, entities.ErrInvalidInput, ""},
		{http.StatusNotFound, entities.ErrNotFound, ""},
		{http.StatusOK, entities.ErrUpstreamUnavailable, `{ "status":`},
		{http.StatusInternalServerError, entities.ErrUpstreamUnavailable, "secret database error"},
		{http.StatusOK, entities.ErrUpstreamUnavailable, "invalid JSON"},
		{http.StatusOK, entities.ErrUpstreamUnavailable, `{"status":"BRONZE"}{"status":"GOLD"}`},
		{http.StatusOK, entities.ErrUpstreamUnavailable, `{"status":"BRONZE"} trailing`},
	} {
		t.Run(http.StatusText(tc.code)+tc.body, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.code)
				if _, err := w.Write([]byte(tc.body)); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			_, err := newLoyalty(t, server.URL, time.Second).Loyalty(t.Context(), "user")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestTimeoutAndCancellation(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	c := newLoyalty(t, server.URL, 20*time.Millisecond)
	_, err := c.Loyalty(t.Context(), "user")
	if !errors.Is(err, entities.ErrUpstreamTimeout) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = c.Loyalty(ctx, "user")
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestNoRedirect(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/other", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	_, err := newLoyalty(t, server.URL, time.Second).Loyalty(t.Context(), "user")
	if !errors.Is(err, entities.ErrUpstreamUnavailable) {
		t.Fatal(oops.Wrap(err))
	}
}

func newLoyalty(t *testing.T, address string, timeout time.Duration) *loyaltyrepo.Repository {
	t.Helper()
	client := httpclient.New(httpclient.Config{
		ReservationURL: address, PaymentURL: address, LoyaltyURL: address, Timeout: timeout,
		MaxInFlight: 2, MaxIdleConnections: 4, MaxIdleConnectionsPerHost: 2, IdleConnectionTimeout: time.Second,
	})
	t.Cleanup(client.CloseIdleConnections)
	return loyaltyrepo.New(client, address)
}

func TestGlobalLimitAndWaitingCancellation(t *testing.T) {
	t.Parallel()
	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		select {
		case <-release:
			if _, err := w.Write([]byte(`{"status":"BRONZE","discount":5,"reservationCount":0}`)); err != nil {
				t.Error(err)
			}
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)
	client := newLoyalty(t, server.URL+"/", time.Second)
	finished := make(chan error, 2)
	for range 2 {
		go func() { _, err := client.Loyalty(t.Context(), "user"); finished <- err }()
	}
	for range 2 {
		<-entered
	}
	ctx, cancel := context.WithCancel(t.Context())
	waiting := make(chan error, 1)
	go func() { _, err := client.Loyalty(ctx, "user"); waiting <- err }()
	cancel()
	if err := <-waiting; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	select {
	case <-entered:
		t.Fatal("request exceeded global limit")
	default:
	}
	release <- struct{}{}
	release <- struct{}{}
	for range 2 {
		if err := <-finished; err != nil {
			t.Fatal(err)
		}
	}
}

func TestTruncatedResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "256")
		if _, err := w.Write([]byte(`{"discount":5}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	client := newLoyalty(t, server.URL, time.Second)
	_, err := client.Loyalty(t.Context(), "user")
	if !errors.Is(err, entities.ErrUpstreamUnavailable) {
		t.Fatal(err)
	}
}

func TestGlobalLimitAcrossRequests(t *testing.T) {
	t.Parallel()
	var active, peak, calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		running := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); running > old; old = peak.Load() {
			if peak.CompareAndSwap(old, running) {
				break
			}
		}
		calls.Add(1)
		time.Sleep(time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := newLoyalty(t, server.URL, time.Second)
	var workers sync.WaitGroup
	for range 24 {
		workers.Go(func() {
			if err := client.ChangeLoyalty(t.Context(), "user", -1); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
	if calls.Load() != 24 || peak.Load() > 2 || active.Load() != 0 {
		t.Fatalf("calls %d peak %d active %d", calls.Load(), peak.Load(), active.Load())
	}
}
