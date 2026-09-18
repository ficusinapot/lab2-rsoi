//go:build integration

package infrastructure

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"lab2/platform/config"
	"lab2/platform/database"
	"lab2/platform/observability/metrics"
	"lab2/reservation/ent"
	"lab2/reservation/internal/app"
	"lab2/reservation/internal/storage/postgres"
	bookingstorage "lab2/reservation/internal/storage/postgres/bookings"
	hotelstorage "lab2/reservation/internal/storage/postgres/hotels"
	httptransport "lab2/reservation/internal/transport/http"
	bookingusecase "lab2/reservation/internal/usecase/bookings"
	hotelusecase "lab2/reservation/internal/usecase/hotels"
)

type Database struct {
	SQL    *sql.DB
	Client *ent.Client
	Config app.Config
}

func NewDatabase(t *testing.T) *Database {
	t.Helper()
	path := testingConfig(t)
	cfg, err := config.Load[app.Config](path)
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(t.Context(), cfg.Database)
	if err != nil {
		t.Fatal(err)
	}
	client := postgres.NewClient(db)
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return &Database{SQL: db, Client: client, Config: cfg}
}

func NewAPI(t *testing.T, db *Database) http.Handler {
	t.Helper()
	telemetry, err := metrics.New(db.Config.Metrics)
	if err != nil {
		t.Fatal(err)
	}
	if err := telemetry.RegisterDatabase(db.SQL); err != nil {
		t.Fatal(err)
	}
	stop, err := telemetry.StartAutometrics(db.Config.Metrics)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(nil) })
	return httptransport.New(
		httptransport.Config{
			HTTP:        db.Config.HTTP,
			OpenAPI:     db.Config.OpenAPI,
			MetricsPath: db.Config.Metrics.Path,
		},
		httptransport.Dependencies{
			Database: db.SQL,
			Hotels:   hotelusecase.New(hotelstorage.New(db.Client), db.Config.Pagination),
			Bookings: bookingusecase.New(bookingstorage.New(db.Client)),
			Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
			Metrics:  telemetry,
		},
	)
}

func testingConfig(t *testing.T) string {
	t.Helper()
	path := testingConfigPath()
	if path == "" {
		t.Skip("TEST_CONFIG is not set")
	}
	return path
}

func testingConfigPath() string {
	return os.Getenv("TEST_CONFIG")
}
