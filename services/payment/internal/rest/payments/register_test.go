package payments

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"lab2/payment/internal/models/entities"
	"lab2/platform/rest"
)

type testService struct {
	id       uuid.UUID
	canceled bool
}

func (s *testService) Create(_ context.Context, price int) (entities.Payment, error) {
	return entities.Payment{PaymentUID: s.id, Status: entities.StatusPaid, Price: price}, nil
}

func (s *testService) Get(context.Context, string) (entities.Payment, error) {
	return entities.Payment{PaymentUID: s.id, Status: entities.StatusPaid, Price: 27000}, nil
}

func (s *testService) Cancel(context.Context, string) error {
	s.canceled = true
	return nil
}

func TestPaymentHTTPContract(t *testing.T) {
	t.Parallel()
	cfg := rest.Config{APIPrefix: "/api/v1", MaxBodyBytes: 4096}
	router, api := rest.New(cfg, rest.OpenAPIConfig{Title: "Payment", Version: "1"})
	service := &testService{id: uuid.New()}
	Register(api, service, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		return out
	}
	created := request(http.MethodPost, "/api/v1/payments", `{"price":27000}`)
	want := `{"paymentUid":"` + service.id.String() + `","status":"PAID","price":27000}`
	if created.Code != http.StatusCreated || strings.TrimSpace(created.Body.String()) != want {
		t.Fatalf("create: %d %s", created.Code, created.Body)
	}
	got := request(http.MethodGet, "/api/v1/payments/"+service.id.String(), "")
	if got.Code != http.StatusOK || strings.TrimSpace(got.Body.String()) != want {
		t.Fatalf("get: %d %s", got.Code, got.Body)
	}
	canceled := request(http.MethodDelete, "/api/v1/payments/"+service.id.String(), "")
	if canceled.Code != http.StatusNoContent || !service.canceled {
		t.Fatalf("cancel: %d %s", canceled.Code, canceled.Body)
	}
}
