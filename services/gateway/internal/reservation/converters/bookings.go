package converters

import (
	"uuid"

	"lab2/gateway/internal/models/entities"
)

type Hotel struct {
	HotelUID uuid.UUID `json:"hotelUid"`
	Name     string    `json:"name"`
	Country  string    `json:"country"`
	City     string    `json:"city"`
	Address  string    `json:"address"`
	Stars    *int      `json:"stars"`
	Price    int       `json:"price"`
}

func (h Hotel) Entity() entities.Hotel {
	return entities.Hotel{
		HotelUID: h.HotelUID, Name: h.Name, Country: h.Country, City: h.City,
		Address: h.Address, Stars: h.Stars, Price: h.Price,
	}
}

type Page struct {
	Page          int     `json:"page"`
	PageSize      int     `json:"pageSize"`
	TotalElements int     `json:"totalElements"`
	Items         []Hotel `json:"items"`
}

func (p Page) Entity() entities.Page {
	items := make([]entities.Hotel, len(p.Items))
	for i, item := range p.Items {
		items[i] = item.Entity()
	}
	return entities.Page{Page: p.Page, PageSize: p.PageSize, TotalElements: p.TotalElements, Items: items}
}

type Reservation struct {
	ReservationUID uuid.UUID `json:"reservationUid"`
	HotelUID       uuid.UUID `json:"hotelUid"`
	PaymentUID     uuid.UUID `json:"paymentUid"`
	StartDate      string    `json:"startDate"`
	EndDate        string    `json:"endDate"`
	Status         string    `json:"status"`
}

func (r Reservation) Entity() entities.InternalReservation {
	return entities.InternalReservation{
		ReservationUID: r.ReservationUID, HotelUID: r.HotelUID, PaymentUID: r.PaymentUID,
		StartDate: r.StartDate, EndDate: r.EndDate, Status: entities.ReservationStatus(r.Status),
	}
}

type CreateRequest struct {
	HotelUID   uuid.UUID `json:"hotelUid"`
	StartDate  string    `json:"startDate"`
	EndDate    string    `json:"endDate"`
	PaymentUID uuid.UUID `json:"paymentUid"`
}

func NewCreateRequest(input entities.InternalCreate) CreateRequest {
	return CreateRequest{
		HotelUID: input.HotelUID, StartDate: input.StartDate,
		EndDate: input.EndDate, PaymentUID: input.PaymentUID,
	}
}
