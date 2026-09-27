package loyalties

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lab2/loyalty/internal/models/entities"
	"lab2/platform/rest"
)

type testService struct {
	count int
	user  string
}

func (s *testService) Get(_ context.Context, user string) (entities.Loyalty, error) {
	s.user = user
	return entities.ForCount(s.count), nil
}

func (s *testService) Change(_ context.Context, user string, delta int) (entities.Loyalty, error) {
	s.user = user
	s.count = max(0, s.count+delta)
	return entities.ForCount(s.count), nil
}

func TestLoyaltyHTTPContract(t *testing.T) {
	t.Parallel()
	cfg := rest.Config{APIPrefix: "/api/v1"}
	router, api := rest.New(cfg, rest.OpenAPIConfig{Title: "Loyalty", Version: "1"})
	service := &testService{}
	Register(api, service, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg)

	request := func(method, path string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-User-Name", "Test Max")
		out := httptest.NewRecorder()
		router.ServeHTTP(out, req)
		return out
	}
	for _, step := range []struct {
		method, path, body string
	}{
		{http.MethodGet, "/api/v1/loyalty", `{"status":"BRONZE","discount":5,"reservationCount":0}`},
		{http.MethodPost, "/api/v1/loyalty/reservations", `{"status":"BRONZE","discount":5,"reservationCount":1}`},
		{http.MethodDelete, "/api/v1/loyalty/reservations", `{"status":"BRONZE","discount":5,"reservationCount":0}`},
	} {
		got := request(step.method, step.path)
		if got.Code != http.StatusOK || strings.TrimSpace(got.Body.String()) != step.body ||
			service.user != "Test Max" {
			t.Fatalf("%s: %d %s user=%s", step.method, got.Code, got.Body, service.user)
		}
	}
}
