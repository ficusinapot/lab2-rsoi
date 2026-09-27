package entities

import "uuid"

type CreateRequest struct {
	HotelUID  string
	StartDate string
	EndDate   string
}

type InternalCreate struct {
	HotelUID   uuid.UUID
	StartDate  string
	EndDate    string
	PaymentUID uuid.UUID
}

type Created struct {
	ReservationUID uuid.UUID
	HotelUID       uuid.UUID
	StartDate      string
	EndDate        string
	Discount       int
	Status         ReservationStatus
	Payment        Payment
}

type UserInfo struct {
	Reservations []Reservation
	Loyalty      Loyalty
}
