package bookings

import (
	"lab2/reservation/internal/domain/bookings"
)

type UserInput struct {
	Username string `header:"X-User-Name"`
}

type ItemInput struct {
	UserInput
	UID string `path:"reservationUid"`
}

type ListOutput struct {
	Body []bookings.Reservation
}

type ItemOutput struct {
	Body bookings.Reservation
}

type CreateInput struct {
	UserInput
	Body CreateRequest
}

type CreateRequest struct {
	HotelUID   string `json:"hotelUid"`
	PaymentUID string `json:"paymentUid"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
}

type CreateOutput struct {
	Location string `header:"Location"`
	Body     bookings.Reservation
}
