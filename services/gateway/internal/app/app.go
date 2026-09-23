package app

import (
	"context"
	"net/http"

	"github.com/samber/oops"
	"lab2/gateway/internal/clients"
	transport "lab2/gateway/internal/transport/http/bookings"
	"lab2/gateway/internal/usecase/bookings"
	"lab2/platform/logging"
	"lab2/platform/observability/metrics"
	"lab2/platform/rest"
)

func Run(ctx context.Context, cfg Config) error {
	logger, closeLogs, err := logging.New(cfg.Logging)
	if err != nil {
		return oops.Wrapf(err, "initialize logging")
	}
	defer closeLogs()
	telemetry, err := metrics.New(cfg.Metrics)
	if err != nil {
		return oops.Wrap(err)
	}
	stop, err := telemetry.StartAutometrics(cfg.Metrics)
	if err != nil {
		return oops.Wrap(err)
	}
	defer stop(nil)
	transport.ConfigureErrors()
	router, api := rest.New(cfg.HTTP, cfg.OpenAPI, rest.Observability(logger, telemetry))
	client := clients.New(cfg.Clients)
	defer client.CloseIdleConnections()
	transport.Register(api, bookings.New(client, client, client, cfg.EnrichmentConcurrency), logger, cfg.HTTP)
	router.Handle(cfg.Metrics.Path, telemetry.Handler())
	router.Get("/manage/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	logger.Info("gateway listening", "address", cfg.HTTP.Address)
	return oops.Wrapf(rest.Serve(ctx, cfg.HTTP, router), "run gateway")
}
