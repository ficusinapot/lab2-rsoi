package main

import (
	"errors"

	"github.com/samber/oops"
	"lab2/gateway/internal/httpclient"
	"lab2/platform/logging"
	"lab2/platform/observability/metrics"
	"lab2/platform/rest"
)

type Config struct {
	EnrichmentConcurrency int                `mapstructure:"enrichment_concurrency"`
	HTTP                  rest.Config        `mapstructure:"http"`
	Logging               logging.Config     `mapstructure:"logging"`
	Metrics               metrics.Config     `mapstructure:"metrics"`
	OpenAPI               rest.OpenAPIConfig `mapstructure:"openapi"`
	Clients               httpclient.Config  `mapstructure:"clients"`
}

func (c Config) Validate() error {
	if c.EnrichmentConcurrency <= 0 {
		return oops.Errorf("enrichment_concurrency must be positive")
	}
	return oops.Wrapf(errors.Join(c.HTTP.Validate(), c.Logging.Validate(),
		c.Metrics.Validate(), c.OpenAPI.Validate(), c.Clients.Validate()), "validate gateway configuration")
}
