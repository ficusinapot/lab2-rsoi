package bookings

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"lab2/gateway/internal/domain"
	"lab2/gateway/internal/domain/bookings"

	"github.com/samber/oops"
)

func (s *Service) Hotels(ctx context.Context, page, size string) (bookings.Page, error) {
	query := url.Values{}
	zeroPage := false
	if page != "" {
		n, err := strconv.Atoi(page)
		if err != nil || n < 0 {
			return bookings.Page{}, oops.Wrap(domain.ErrInvalidInput)
		}
		zeroPage = n == 0
		if zeroPage {
			n = 1
		}
		query.Set("page", strconv.Itoa(n))
	}
	if size != "" {
		n, err := strconv.Atoi(size)
		if err != nil || n < 1 || n > 100 {
			return bookings.Page{}, oops.Wrap(domain.ErrInvalidInput)
		}
		query.Set("size", strconv.Itoa(n))
	}
	suffix := ""
	if len(query) > 0 {
		suffix = "?" + query.Encode()
	}
	item, err := s.reservations.Hotels(ctx, suffix)
	if err != nil {
		return bookings.Page{}, oops.Wrapf(err, "list hotels")
	}
	if zeroPage {
		item.Page = 0
	}
	return item, nil
}

func (s *Service) Loyalty(ctx context.Context, user string) (bookings.Loyalty, error) {
	if err := validUser(user); err != nil {
		return bookings.Loyalty{}, err
	}
	item, err := s.loyalty.Loyalty(ctx, user)
	return item, oops.Wrapf(err, "get loyalty")
}

func (s *Service) enrich(ctx context.Context, item bookings.InternalReservation) (bookings.Reservation, error) {
	hotel, err := s.reservations.Hotel(ctx, item.HotelUID.String())
	if err != nil {
		return bookings.Reservation{}, oops.Wrapf(err, "get hotel")
	}
	payment, err := s.payments.Payment(ctx, item.PaymentUID.String())
	if err != nil {
		return bookings.Reservation{}, oops.Wrapf(err, "get payment")
	}
	return bookings.Reservation{
		ReservationUID: item.ReservationUID, StartDate: item.StartDate, EndDate: item.EndDate, Status: item.Status,
		Hotel: bookings.HotelInfo{
			HotelUID: hotel.HotelUID, Name: hotel.Name, Stars: hotel.Stars,
			FullAddress: strings.Join([]string{hotel.Country, hotel.City, hotel.Address}, ", "),
		}, Payment: payment,
	}, nil
}

func (s *Service) List(ctx context.Context, user string) ([]bookings.Reservation, error) {
	if err := validUser(user); err != nil {
		return nil, err
	}
	rows, err := s.reservations.Reservations(ctx, user)
	if err != nil {
		return nil, oops.Wrapf(err, "list reservations")
	}
	return s.enrichList(ctx, rows)
}

func (s *Service) Get(ctx context.Context, id, user string) (bookings.Reservation, error) {
	if err := validUser(user); err != nil {
		return bookings.Reservation{}, err
	}
	if err := validID(id); err != nil {
		return bookings.Reservation{}, err
	}
	row, err := s.reservations.Reservation(ctx, id, user)
	if err != nil {
		return bookings.Reservation{}, oops.Wrapf(err, "get reservation")
	}
	return s.enrich(ctx, row)
}

func (s *Service) Me(ctx context.Context, user string) (bookings.UserInfo, error) {
	items, err := s.List(ctx, user)
	if err != nil {
		return bookings.UserInfo{}, err
	}
	loyalty, err := s.Loyalty(ctx, user)
	if err != nil {
		return bookings.UserInfo{}, err
	}
	return bookings.UserInfo{Reservations: items, Loyalty: loyalty}, nil
}
