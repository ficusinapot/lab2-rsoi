package entities

import "uuid"

type ReservationStatus string

const (
	ReservationPaid     ReservationStatus = "PAID"
	ReservationCanceled ReservationStatus = "CANCELED"
)

type InternalReservation struct {
	ReservationUID uuid.UUID
	HotelUID       uuid.UUID
	PaymentUID     uuid.UUID
	StartDate      string
	EndDate        string
	Status         ReservationStatus
}
type Reservation struct {
	ReservationUID uuid.UUID
	Hotel          HotelInfo
	StartDate      string
	EndDate        string
	Status         ReservationStatus
	Payment        Payment
}
