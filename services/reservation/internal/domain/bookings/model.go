package bookings

import (
	"time"
	"uuid"
)

type Status string

const (
	StatusPaid     Status = "PAID"
	StatusCanceled Status = "CANCELED"
)

type Reservation struct {
	ReservationUID uuid.UUID `json:"reservationUid" format:"uuid"`
	HotelUID       uuid.UUID `json:"hotelUid" format:"uuid"`
	PaymentUID     uuid.UUID `json:"paymentUid" format:"uuid"`
	Status         Status    `json:"status" enum:"PAID,CANCELED"`
	StartDate      string    `json:"startDate" format:"date"`
	EndDate        string    `json:"endDate" format:"date"`
}

type Booking struct {
	ReservationUID uuid.UUID
	HotelUID       uuid.UUID
	PaymentUID     uuid.UUID
	Username       string
	StartDate      time.Time
	EndDate        time.Time
}
