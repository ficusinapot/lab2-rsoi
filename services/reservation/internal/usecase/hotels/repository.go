package hotels

import (
	"context"
	"uuid"

	"lab2/reservation/internal/domain/hotels"
)

type Repository interface {
	List(context.Context, int, int) (hotels.Page, error)
	Get(context.Context, uuid.UUID) (hotels.Hotel, error)
}

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=repository_mock_test.go -package=hotels Repository
