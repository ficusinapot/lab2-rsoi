package resttransport

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"lab2/platform/observability/metrics"
	"lab2/platform/rest"
	bookinghttp "lab2/reservation/internal/rest/bookings"
	hotelhttp "lab2/reservation/internal/rest/hotels"
)

type Config struct {
	HTTP        rest.Config
	OpenAPI     rest.OpenAPIConfig
	MetricsPath string
}

type Dependencies struct {
	Database *sql.DB
	Hotels   hotelhttp.Service
	Bookings bookinghttp.Service
	Logger   *slog.Logger
	Metrics  *metrics.Metrics
}

func New(cfg Config, deps Dependencies) *chi.Mux {
	router, api := rest.New(cfg.HTTP, cfg.OpenAPI, rest.Observability(deps.Logger, deps.Metrics))
	hotelhttp.Register(api, deps.Hotels, deps.Logger)
	bookinghttp.Register(api, deps.Bookings, deps.Logger, cfg.HTTP)
	router.Handle(cfg.MetricsPath, deps.Metrics.Handler())
	router.Get("/manage/health", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.Database.PingContext(r.Context()); err != nil {
			deps.Logger.ErrorContext(r.Context(), "database health check failed", "error", err)
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return router
}
