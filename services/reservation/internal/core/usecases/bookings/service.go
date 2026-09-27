package bookings

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/samber/oops"

	"lab2/reservation/internal/models/dbifc"
	bookings "lab2/reservation/internal/models/entities"
)

type Service struct{ repo dbifc.Bookings }

func New(repo dbifc.Bookings) *Service { return &Service{repo: repo} }

func validateUser(name string) error {
	blank := strings.TrimSpace(name) == ""
	invalidUTF8 := !utf8.ValidString(name)
	tooLong := utf8.RuneCountInString(name) > 80
	if blank || invalidUTF8 || tooLong {
		return oops.In("validation").Code("invalid_input").
			Public("valid username is required").
			Wrap(errors.Join(bookings.ErrInvalidInput, bookings.ErrInvalidUsername))
	}
	return nil
}

func validateUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil(), oops.In("validation").Code("invalid_input").Public("invalid UUID").
			Wrap(errors.Join(bookings.ErrInvalidInput, err))
	}
	return parsed, nil
}

func (s *Service) List(ctx context.Context, name string) ([]bookings.Reservation, error) {
	if err := validateUser(name); err != nil {
		return nil, err
	}
	result, err := s.repo.List(ctx, name)
	return result, oops.FromContext(ctx).In("bookings.usecase").With("operation", "list").Wrapf(err, "operation failed")
}

func (s *Service) Get(ctx context.Context, id, name string) (bookings.Reservation, error) {
	if err := validateUser(name); err != nil {
		return bookings.Reservation{}, err
	}
	uid, err := validateUUID(id)
	if err != nil {
		return bookings.Reservation{}, err
	}
	result, err := s.repo.Get(ctx, uid, name)
	return result, oops.FromContext(ctx).In("bookings.usecase").With("operation", "get").Wrapf(err, "operation failed")
}

func (s *Service) Create(ctx context.Context, name string, input bookings.CreateInput) (bookings.Reservation, error) {
	if err := validateUser(name); err != nil {
		return bookings.Reservation{}, err
	}
	hotelUID, err := validateUUID(input.HotelUID)
	if err != nil {
		return bookings.Reservation{}, err
	}
	paymentUID, err := validateUUID(input.PaymentUID)
	if err != nil {
		return bookings.Reservation{}, err
	}
	start, startErr := time.Parse(time.DateOnly, input.StartDate)
	end, endErr := time.Parse(time.DateOnly, input.EndDate)
	invalidDates := startErr != nil || endErr != nil
	if invalidDates || !end.After(start) {
		return bookings.Reservation{}, oops.In("validation").Code("invalid_input").
			Public("valid increasing dates are required").
			Wrap(errors.Join(bookings.ErrInvalidInput, startErr, endErr))
	}
	id := uuid.New()
	result, err := s.repo.Create(ctx, bookings.Booking{
		ReservationUID: id, HotelUID: hotelUID,
		PaymentUID: paymentUID, Username: name, StartDate: start, EndDate: end,
	})
	return result, oops.FromContext(ctx).In("bookings.usecase").With("operation", "create").Wrapf(err, "operation failed")
}

func (s *Service) Cancel(ctx context.Context, id, name string) error {
	if err := validateUser(name); err != nil {
		return err
	}
	uid, err := validateUUID(id)
	if err != nil {
		return err
	}
	return oops.FromContext(ctx).In("bookings.usecase").With("operation", "cancel").
		Wrapf(s.repo.Cancel(ctx, uid, name), "operation failed")
}
