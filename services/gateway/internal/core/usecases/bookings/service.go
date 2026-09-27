package bookings

import (
	"strings"
	"unicode/utf8"
	"uuid"

	"github.com/samber/oops"
	bookings "lab2/gateway/internal/models/entities"
	"lab2/gateway/internal/models/loyaltyifc"
	"lab2/gateway/internal/models/paymentifc"
	"lab2/gateway/internal/models/reservationifc"
)

type Service struct {
	reservations reservationifc.Bookings
	payments     paymentifc.Payments
	loyalty      loyaltyifc.Client
	concurrency  int
}

func New(
	reservations reservationifc.Bookings, payments paymentifc.Payments, loyalty loyaltyifc.Client, concurrency int,
) *Service {
	return &Service{reservations: reservations, payments: payments, loyalty: loyalty, concurrency: concurrency}
}

func validUser(name string) error {
	if strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return oops.Wrap(bookings.ErrInvalidInput)
	}
	return nil
}

func validID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, oops.Wrap(bookings.ErrInvalidInput)
	}
	return parsed, nil
}
