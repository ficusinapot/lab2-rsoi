package dbifc

import (
	"context"

	"lab2/loyalty/internal/models/entities"
)

type Loyalties interface {
	Get(context.Context, string) (entities.Loyalty, error)
	Change(context.Context, string, int) (entities.Loyalty, error)
}
