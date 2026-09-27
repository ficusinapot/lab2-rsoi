//go:build integration

package bookings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"

	"lab2/gateway/internal/httpclient"
	"lab2/gateway/internal/models/entities"
	"lab2/gateway/internal/payment/repos"
)

func TestPaymentAdapter(t *testing.T) {
	id := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/payments" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		var body struct {
			Price int `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Price != 27000 {
			t.Errorf("unexpected price: %+v, %v", body, err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"paymentUid": id, "status": "PAID", "price": 27000,
		})
	}))
	t.Cleanup(server.Close)
	client := httpclient.New(httpclient.Config{
		Timeout: time.Second, MaxInFlight: 2, MaxIdleConnections: 4,
		MaxIdleConnectionsPerHost: 2, IdleConnectionTimeout: time.Second,
	})
	t.Cleanup(client.CloseIdleConnections)
	got, err := repos.New(client, server.URL).CreatePayment(t.Context(), 27000)
	if err != nil || got.PaymentUID != id || got.Status != entities.PaymentPaid || got.Price != 27000 {
		t.Fatalf("payment: %+v, %v", got, err)
	}
}
