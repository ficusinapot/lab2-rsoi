package loyaltyifc

import (
	"context"

	"lab2/gateway/internal/models/entities"
)

type Client interface {
	Loyalty(context.Context, string) (entities.Loyalty, error)
	ChangeLoyalty(context.Context, string, int) error
}
