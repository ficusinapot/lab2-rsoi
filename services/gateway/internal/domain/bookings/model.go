package bookings

import "uuid"

type (
	LoyaltyStatus     string
	PaymentStatus     string
	ReservationStatus string
)

const (
	LoyaltyBronze       LoyaltyStatus     = "BRONZE"
	LoyaltySilver       LoyaltyStatus     = "SILVER"
	LoyaltyGold         LoyaltyStatus     = "GOLD"
	PaymentPaid         PaymentStatus     = "PAID"
	PaymentCanceled     PaymentStatus     = "CANCELED"
	ReservationPaid     ReservationStatus = "PAID"
	ReservationCanceled ReservationStatus = "CANCELED"
)

type Hotel struct {
	HotelUID uuid.UUID `json:"hotelUid" format:"uuid"`
	Name     string    `json:"name"`
	Country  string    `json:"country"`
	City     string    `json:"city"`
	Address  string    `json:"address"`
	Stars    *int      `json:"stars"`
	Price    int       `json:"price"`
}
type Page struct {
	Page          int     `json:"page"`
	PageSize      int     `json:"pageSize"`
	TotalElements int     `json:"totalElements"`
	Items         []Hotel `json:"items"`
}
type Payment struct {
	Status PaymentStatus `json:"status" enum:"PAID,CANCELED"`
	Price  int           `json:"price"`
}
type InternalPayment struct {
	Payment
	PaymentUID uuid.UUID `json:"paymentUid"`
}
type Loyalty struct {
	Status           LoyaltyStatus `json:"status" enum:"BRONZE,SILVER,GOLD"`
	Discount         int           `json:"discount"`
	ReservationCount int           `json:"reservationCount"`
}
type InternalReservation struct {
	ReservationUID uuid.UUID         `json:"reservationUid"`
	HotelUID       uuid.UUID         `json:"hotelUid"`
	PaymentUID     uuid.UUID         `json:"paymentUid"`
	StartDate      string            `json:"startDate"`
	EndDate        string            `json:"endDate"`
	Status         ReservationStatus `json:"status"`
}
type HotelInfo struct {
	HotelUID    uuid.UUID `json:"hotelUid" format:"uuid"`
	Name        string    `json:"name"`
	FullAddress string    `json:"fullAddress"`
	Stars       *int      `json:"stars"`
}
type Reservation struct {
	ReservationUID uuid.UUID         `json:"reservationUid" format:"uuid"`
	Hotel          HotelInfo         `json:"hotel"`
	StartDate      string            `json:"startDate" format:"date"`
	EndDate        string            `json:"endDate" format:"date"`
	Status         ReservationStatus `json:"status" enum:"PAID,CANCELED"`
	Payment        Payment           `json:"payment"`
}
type CreateRequest struct {
	HotelUID  string `json:"hotelUid"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}
type InternalCreate struct {
	CreateRequest
	PaymentUID string `json:"paymentUid"`
}
type Created struct {
	ReservationUID uuid.UUID         `json:"reservationUid" format:"uuid"`
	HotelUID       uuid.UUID         `json:"hotelUid" format:"uuid"`
	StartDate      string            `json:"startDate" format:"date"`
	EndDate        string            `json:"endDate" format:"date"`
	Discount       int               `json:"discount"`
	Status         ReservationStatus `json:"status" enum:"PAID,CANCELED"`
	Payment        Payment           `json:"payment"`
}
type UserInfo struct {
	Reservations []Reservation `json:"reservations"`
	Loyalty      Loyalty       `json:"loyalty"`
}
