package bookings

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"

	"go.uber.org/mock/gomock"
	bookings "lab2/reservation/internal/models/entities"
)

const (
	bookingID    = "049161bb-badd-4fa8-9d90-87c9a82b0668"
	bookingStart = "2021-10-08"
	bookingEnd   = "2021-10-11"
	testUser     = "user"
)

func TestServiceOwnership(t *testing.T) {
	t.Parallel()
	const username = " Test Max "
	cause := bookings.ErrNotFound
	ctrl := gomock.NewController(t)
	repo := NewMockBookings(ctrl)
	repo.EXPECT().List(gomock.Any(), username).
		Return([]bookings.Reservation{{ReservationUID: uuid.MustParse(bookingID)}}, nil)
	repo.EXPECT().Get(gomock.Any(), uuid.MustParse(bookingID), username).
		Return(bookings.Reservation{}, cause)
	repo.EXPECT().Cancel(gomock.Any(), uuid.MustParse(bookingID), username).
		Return(cause)
	s := New(repo)
	if result, err := s.List(t.Context(), username); err != nil || len(result) != 1 {
		t.Fatalf("list=%v err=%v", result, err)
	}
	if _, err := s.Get(t.Context(), bookingID, username); !errors.Is(err, cause) {
		t.Fatalf("get error=%v", err)
	}
	if err := s.Cancel(t.Context(), bookingID, username); !errors.Is(err, cause) {
		t.Fatalf("cancel error=%v", err)
	}
}

func TestServiceCreate(t *testing.T) {
	t.Parallel()
	cause := errors.New("write failed")
	ctrl := gomock.NewController(t)
	repo := NewMockBookings(ctrl)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, input bookings.Booking) (bookings.Reservation, error) {
			if input.Username != "Test Max" || input.HotelUID.String() != bookingID || input.PaymentUID.String() != bookingID {
				t.Fatalf("booking=%+v", input)
			}
			if input.StartDate.Format(time.DateOnly) != bookingStart || input.EndDate.Format(time.DateOnly) != bookingEnd {
				t.Fatalf("dates=%+v", input)
			}
			return bookings.Reservation{}, cause
		},
	)
	if _, err := New(repo).Create(t.Context(), "Test Max", bookings.CreateInput{
		HotelUID:   bookingID,
		PaymentUID: bookingID,
		StartDate:  bookingStart,
		EndDate:    bookingEnd,
	}); !errors.Is(err, cause) {
		t.Fatalf("create error=%v", err)
	}
}

func TestServiceValidation(t *testing.T) {
	t.Parallel()
	const (
		invalidUUID = "bad"
		invalidDate = "not a date"
	)
	for _, tc := range []struct {
		name, username string
		input          bookings.CreateInput
		parseDate      bool
	}{
		{name: "empty user", username: " "},
		{name: "long user", username: strings.Repeat("a", 81)},
		{name: "invalid hotel", username: testUser, input: bookings.CreateInput{HotelUID: invalidUUID}},
		{name: "invalid payment", username: testUser, input: bookings.CreateInput{
			HotelUID: bookingID, PaymentUID: invalidUUID,
		}},
		{name: "same dates", username: testUser, input: bookings.CreateInput{
			HotelUID: bookingID, PaymentUID: bookingID, StartDate: bookingStart, EndDate: bookingStart,
		}},
		{name: "invalid dates", username: testUser, parseDate: true, input: bookings.CreateInput{
			HotelUID: bookingID, PaymentUID: bookingID, StartDate: invalidDate, EndDate: bookingEnd,
		}},
		{name: "invalid end date", username: testUser, parseDate: true, input: bookings.CreateInput{
			HotelUID: bookingID, PaymentUID: bookingID, StartDate: bookingStart, EndDate: invalidDate,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := New(NewMockBookings(ctrl))
			_, err := s.Create(t.Context(), tc.username, tc.input)
			if !errors.Is(err, bookings.ErrInvalidInput) {
				t.Fatalf("validation error=%v", err)
			}
			if tc.input.HotelUID == invalidUUID || tc.input.PaymentUID == invalidUUID {
				_, cause := uuid.Parse(invalidUUID)
				if !errors.Is(err, cause) {
					t.Fatalf("uuid parse cause lost: %v", err)
				}
			}
			if tc.parseDate {
				var cause *time.ParseError
				if !errors.As(err, &cause) || cause.Value != invalidDate {
					t.Fatalf("date parse cause lost: %v", err)
				}
			}
		})
	}
	ctrl := gomock.NewController(t)
	s := New(NewMockBookings(ctrl))
	if _, err := s.List(t.Context(), ""); !errors.Is(err, bookings.ErrInvalidUsername) {
		t.Fatalf("list validation=%v", err)
	}
	if _, err := s.Get(t.Context(), invalidUUID, testUser); !errors.Is(err, bookings.ErrInvalidInput) {
		t.Fatalf("get validation=%v", err)
	}
	if err := s.Cancel(t.Context(), invalidUUID, testUser); !errors.Is(err, bookings.ErrInvalidInput) {
		t.Fatalf("cancel validation=%v", err)
	}
}
