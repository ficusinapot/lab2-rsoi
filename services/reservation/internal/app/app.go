package app

import (
	"context"
	"errors"

	"lab2/platform/database"
	"lab2/platform/logging"
	"lab2/platform/observability/metrics"
	"lab2/platform/rest"
	"lab2/reservation/internal/storage/postgres"
	bookingstorage "lab2/reservation/internal/storage/postgres/bookings"
	hotelstorage "lab2/reservation/internal/storage/postgres/hotels"
	httptransport "lab2/reservation/internal/transport/http"
	"lab2/reservation/internal/usecase/bookings"
	"lab2/reservation/internal/usecase/hotels"

	"github.com/samber/oops"
)

func Run(ctx context.Context, cfg Config) (runErr error) {
	logger, closeLogs, err := logging.New(cfg.Logging)
	if err != nil {
		return oops.Wrapf(err, "initialize logging")
	}
	defer closeLogs()

	pool, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return oops.Wrapf(err, "initialize database")
	}
	defer func() { runErr = oops.Wrapf(errors.Join(runErr, pool.Close()), "close database") }()

	telemetry, err := metrics.New(cfg.Metrics)
	if err != nil {
		return oops.Wrapf(err, "initialize metrics")
	}
	if err := telemetry.RegisterDatabase(pool); err != nil {
		return oops.Wrap(err)
	}
	stopMetrics, err := telemetry.StartAutometrics(cfg.Metrics)
	if err != nil {
		return oops.Wrap(err)
	}
	defer stopMetrics(nil)

	client := postgres.NewClient(pool)
	hotelService := hotels.New(hotelstorage.New(client), cfg.Pagination)
	bookingService := bookings.New(bookingstorage.New(client))
	handler := httptransport.New(httptransport.Config{HTTP: cfg.HTTP, OpenAPI: cfg.OpenAPI, MetricsPath: cfg.Metrics.Path},
		httptransport.Dependencies{
			Database: pool,
			Hotels:   hotelService,
			Bookings: bookingService,
			Logger:   logger,
			Metrics:  telemetry,
		})
	logger.Info("reservation service listening", "address", cfg.HTTP.Address)
	return oops.Wrapf(rest.Serve(ctx, cfg.HTTP, handler), "run reservation service")
}
