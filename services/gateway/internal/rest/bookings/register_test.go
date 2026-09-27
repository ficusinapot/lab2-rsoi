package bookings

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/oops"
	"lab2/gateway/internal/models/coreifc"
	model "lab2/gateway/internal/models/entities"
	"lab2/platform/rest"
)

const testUser = "user"

type transportStub struct {
	coreifc.Bookings
	user string
}

func (s *transportStub) Loyalty(_ context.Context, user string) (model.Loyalty, error) {
	s.user = user
	return model.Loyalty{Status: model.LoyaltyBronze, Discount: 5}, nil
}

func (*transportStub) Hotels(context.Context, string, string) (model.Page, error) {
	return model.Page{Page: 0, PageSize: 1, Items: []model.Hotel{{Name: "hotel"}}}, nil
}

func (*transportStub) List(context.Context, string) ([]model.Reservation, error) {
	return []model.Reservation{{Status: model.ReservationPaid}}, nil
}

func (*transportStub) Get(_ context.Context, _, user string) (model.Reservation, error) {
	if user == "stranger" {
		return model.Reservation{}, model.ErrNotFound
	}
	return model.Reservation{
		Status: model.ReservationPaid, Payment: model.Payment{Status: model.PaymentPaid, Price: 100},
	}, nil
}

func (*transportStub) Me(context.Context, string) (model.UserInfo, error) {
	return model.UserInfo{Reservations: []model.Reservation{}, Loyalty: model.Loyalty{Status: model.LoyaltyBronze}}, nil
}

func (*transportStub) Create(_ context.Context, _ string, input model.CreateRequest) (model.Created, error) {
	if input.HotelUID == "bad" {
		return model.Created{}, oops.Wrap(model.ErrInvalidInput)
	}
	return model.Created{Status: model.ReservationPaid}, nil
}

func (*transportStub) Cancel(context.Context, string, string) error { return nil }

func TestGatewayContract(t *testing.T) {
	t.Parallel()
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
	out := request(http.MethodGet, "/api/v1/loyalty", "", testUser)
	if out.Code != 200 || stub.user != testUser {
		t.Fatalf("%d %s user %s", out.Code, out.Body, stub.user)
	}
	out = request(http.MethodGet, "/api/v1/loyalty", "", "")
	if out.Code != 400 {
		t.Fatalf("%d %s", out.Code, out.Body)
	}
	if !strings.Contains(out.Body.String(), `"message":"validation failed"`) ||
		!strings.Contains(out.Body.String(), `"field":"header.X-User-Name"`) {
		t.Fatalf("unexpected validation body: %s", out.Body)
	}
	out = request(http.MethodPost, "/api/v1/reservations", `{"hotelUid":"bad"}`, testUser)
	if out.Code != 400 {
		t.Fatalf("%d %s", out.Code, out.Body)
	}
	out = request(http.MethodGet, "/api/v1/reservations/example", "", testUser)
	if out.Code != 200 || strings.Contains(out.Body.String(), "paymentUid") {
		t.Fatalf("%d %s", out.Code, out.Body)
	}
	for _, tc := range []struct {
		method, path, body, user string
		status                   int
	}{
		{http.MethodGet, "/api/v1/hotels?page=0&size=1", "", "", http.StatusOK},
		{http.MethodGet, "/api/v1/reservations", "", testUser, http.StatusOK},
		{http.MethodGet, "/api/v1/me", "", testUser, http.StatusOK},
		{http.MethodGet, "/api/v1/reservations/example", "", "stranger", http.StatusNotFound},
		{
			http.MethodPost, "/api/v1/reservations",
			`{"hotelUid":"049161bb-badd-4fa8-9d90-87c9a82b0668","startDate":"2026-01-01","endDate":"2026-01-04"}`,
			testUser, http.StatusOK,
		},
		{http.MethodDelete, "/api/v1/reservations/example", "", testUser, http.StatusNoContent},
	} {
		out = request(tc.method, tc.path, tc.body, tc.user)
		if out.Code != tc.status || strings.Contains(out.Body.String(), "paymentUid") {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, out.Code, out.Body)
		}
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

type otherInput struct {
	Value string `header:"X-Required" required:"true"`
}

type otherOutput struct{ Body string }

func TestGatewayErrorsAreLocal(t *testing.T) {
	t.Parallel()
	original := huma.NewError
	cfg := rest.Config{APIPrefix: "/api/v1", MaxBodyBytes: 4096}
	docs := rest.OpenAPIConfig{Title: "API", Version: "1", DocsPath: "/docs", SchemaPath: "/openapi"}
	_, gateway := rest.New(cfg, docs)
	Register(gateway, &transportStub{}, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)
	if reflect.ValueOf(huma.NewError).Pointer() != reflect.ValueOf(original).Pointer() {
		t.Fatal("gateway changed Huma's global error factory")
	}
	router, other := rest.New(cfg, docs)
	huma.Register(other, huma.Operation{OperationID: "other", Method: http.MethodGet, Path: "/other"},
		func(_ context.Context, _ *otherInput) (*otherOutput, error) { return &otherOutput{Body: "ok"}, nil })
	out := httptest.NewRecorder()
	router.ServeHTTP(out, httptest.NewRequest(http.MethodGet, "/api/v1/other", http.NoBody))
	if out.Code != http.StatusUnprocessableEntity {
		t.Fatalf("second Huma API returned %d: %s", out.Code, out.Body)
	}
}
