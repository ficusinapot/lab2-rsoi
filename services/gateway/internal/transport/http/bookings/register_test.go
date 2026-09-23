package bookings

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/oops"
	"lab2/gateway/internal/domain"
	model "lab2/gateway/internal/domain/bookings"
	"lab2/platform/rest"
)

type transportStub struct {
	Service
	user string
}

func (s *transportStub) Loyalty(_ context.Context, user string) (model.Loyalty, error) {
	s.user = user
	return model.Loyalty{Status: model.LoyaltyBronze, Discount: 5}, nil
}

func (*transportStub) Get(context.Context, string, string) (model.Reservation, error) {
	return model.Reservation{
		Status: model.ReservationPaid, Payment: model.Payment{Status: model.PaymentPaid, Price: 100},
	}, nil
}

func (*transportStub) Create(context.Context, string, model.CreateRequest) (model.Created, error) {
	return model.Created{}, oops.Wrap(domain.ErrInvalidInput)
}

//nolint:paralleltest // Configures and restores Huma's process-wide error factory.
func TestGatewayContract(t *testing.T) {
	// Huma's process-wide error factory must be configured before registering routes.
	original := huma.NewError
	defer func() { huma.NewError = original }()
	ConfigureErrors()
	stub := &transportStub{}
	cfg := rest.Config{APIPrefix: "/api/v1", MaxBodyBytes: 4096}
	docs := rest.OpenAPIConfig{Title: "Gateway", Version: "1", DocsPath: "/docs", SchemaPath: "/openapi"}
	router, api := rest.New(cfg, docs)
	Register(api, stub, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)
	request := func(method, path, body, user string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if user != "" {
			req.Header.Set("X-User-Name", user)
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		return out
	}
	out := request(http.MethodGet, "/api/v1/loyalty", "", "user")
	if out.Code != 200 || stub.user != "user" {
		t.Fatalf("%d %s user %s", out.Code, out.Body, stub.user)
	}
	out = request(http.MethodGet, "/api/v1/loyalty", "", "")
	if out.Code != 400 {
		t.Fatalf("%d %s", out.Code, out.Body)
	}
	out = request(http.MethodPost, "/api/v1/reservations", `{"hotelUid":"bad"}`, "user")
	if out.Code != 400 {
		t.Fatalf("%d %s", out.Code, out.Body)
	}
	out = request(http.MethodGet, "/api/v1/reservations/example", "", "user")
	if out.Code != 200 || strings.Contains(out.Body.String(), "paymentUid") {
		t.Fatalf("%d %s", out.Code, out.Body)
	}
	schema, err := json.Marshal(api.OpenAPI())
	if err != nil {
		t.Fatal(err)
	}
	for _, enum := range []string{`"enum":["BRONZE","SILVER","GOLD"]`, `"enum":["PAID","CANCELED"]`} {
		if !strings.Contains(string(schema), enum) {
			t.Fatalf("missing %s", enum)
		}
	}
}
