package bookings

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"lab2/platform/rest"
)

const reservationsTag = "Reservations"

func Register(api huma.API, service Service, logger *slog.Logger, cfg rest.Config) {
	registerList(api, service, logger)
	registerGet(api, service, logger)
	registerCreate(api, service, logger, cfg)
	registerCancel(api, service, logger)
}
