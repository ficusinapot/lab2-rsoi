package hotels

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
)

const hotelsTag = "Hotels"

func Register(api huma.API, service Service, logger *slog.Logger) {
	registerList(api, service, logger)
	registerGet(api, service, logger)
}
