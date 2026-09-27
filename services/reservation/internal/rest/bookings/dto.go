package bookings

import (
	"uuid"

	"lab2/reservation/internal/models/entities"
)

type UserInput struct {
	Username string `header:"X-User-Name"`
}

type ItemInput struct {
	UserInput
	UID string `path:"reservationUid"`
}

type ListOutput struct {
	Body []ReservationResponse
}

type ItemOutput struct {
	Body ReservationResponse
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
	Body     ReservationResponse
}

type ReservationResponse struct {
	ReservationUID uuid.UUID `json:"reservationUid" format:"uuid"`
	HotelUID       uuid.UUID `json:"hotelUid" format:"uuid"`
	PaymentUID     uuid.UUID `json:"paymentUid" format:"uuid"`
	Status         string    `json:"status" enum:"PAID,CANCELED"`
	StartDate      string    `json:"startDate" format:"date"`
	EndDate        string    `json:"endDate" format:"date"`
}

func reservationResponse(item entities.Reservation) ReservationResponse {
	return ReservationResponse{
		ReservationUID: item.ReservationUID, HotelUID: item.HotelUID,
		PaymentUID: item.PaymentUID, Status: string(item.Status),
		StartDate: item.StartDate, EndDate: item.EndDate,
	}
}

func reservationResponses(items []entities.Reservation) []ReservationResponse {
	if items == nil {
		return nil
	}
	out := make([]ReservationResponse, len(items))
	for i, item := range items {
		out[i] = reservationResponse(item)
	}
	return out
}
