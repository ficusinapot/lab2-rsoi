package bookings

import (
	"uuid"

	bookings "lab2/gateway/internal/models/entities"
)

type HotelResponse struct {
	HotelUID uuid.UUID `json:"hotelUid" format:"uuid"`
	Name     string    `json:"name"`
	Country  string    `json:"country"`
	City     string    `json:"city"`
	Address  string    `json:"address"`
	Stars    *int      `json:"stars"`
	Price    int       `json:"price"`
}

func hotelResponse(h bookings.Hotel) HotelResponse {
	return HotelResponse{
		HotelUID: h.HotelUID, Name: h.Name, Country: h.Country, City: h.City,
		Address: h.Address, Stars: h.Stars, Price: h.Price,
	}
}

type PageResponse struct {
	Page          int             `json:"page"`
	PageSize      int             `json:"pageSize"`
	TotalElements int             `json:"totalElements"`
	Items         []HotelResponse `json:"items"`
}

func pageResponse(p bookings.Page) PageResponse {
	items := make([]HotelResponse, len(p.Items))
	for i, item := range p.Items {
		items[i] = hotelResponse(item)
	}
	return PageResponse{Page: p.Page, PageSize: p.PageSize, TotalElements: p.TotalElements, Items: items}
}

type PaymentResponse struct {
	Status string `json:"status" enum:"PAID,CANCELED"`
	Price  int    `json:"price"`
}

func paymentResponse(p bookings.Payment) PaymentResponse {
	return PaymentResponse{Status: string(p.Status), Price: p.Price}
}

type LoyaltyResponse struct {
	Status           string `json:"status" enum:"BRONZE,SILVER,GOLD"`
	Discount         int    `json:"discount"`
	ReservationCount int    `json:"reservationCount"`
}

func loyaltyResponse(l bookings.Loyalty) LoyaltyResponse {
	return LoyaltyResponse{Status: string(l.Status), Discount: l.Discount, ReservationCount: l.ReservationCount}
}

type HotelInfoResponse struct {
	HotelUID    uuid.UUID `json:"hotelUid" format:"uuid"`
	Name        string    `json:"name"`
	FullAddress string    `json:"fullAddress"`
	Stars       *int      `json:"stars"`
}

type ReservationResponse struct {
	ReservationUID uuid.UUID         `json:"reservationUid" format:"uuid"`
	Hotel          HotelInfoResponse `json:"hotel"`
	StartDate      string            `json:"startDate" format:"date"`
	EndDate        string            `json:"endDate" format:"date"`
	Status         string            `json:"status" enum:"PAID,CANCELED"`
	Payment        PaymentResponse   `json:"payment"`
}

func reservationResponse(r bookings.Reservation) ReservationResponse {
	return ReservationResponse{
		ReservationUID: r.ReservationUID,
		Hotel: HotelInfoResponse{
			HotelUID: r.Hotel.HotelUID, Name: r.Hotel.Name,
			FullAddress: r.Hotel.FullAddress, Stars: r.Hotel.Stars,
		},
		StartDate: r.StartDate, EndDate: r.EndDate, Status: string(r.Status), Payment: paymentResponse(r.Payment),
	}
}

func reservationResponses(rows []bookings.Reservation) []ReservationResponse {
	items := make([]ReservationResponse, len(rows))
	for i, row := range rows {
		items[i] = reservationResponse(row)
	}
	return items
}

type CreateRequest struct {
	HotelUID  string `json:"hotelUid"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

func (r CreateRequest) domain() bookings.CreateRequest {
	return bookings.CreateRequest{HotelUID: r.HotelUID, StartDate: r.StartDate, EndDate: r.EndDate}
}

type CreatedResponse struct {
	ReservationUID uuid.UUID       `json:"reservationUid" format:"uuid"`
	HotelUID       uuid.UUID       `json:"hotelUid" format:"uuid"`
	StartDate      string          `json:"startDate" format:"date"`
	EndDate        string          `json:"endDate" format:"date"`
	Discount       int             `json:"discount"`
	Status         string          `json:"status" enum:"PAID,CANCELED"`
	Payment        PaymentResponse `json:"payment"`
}

func createdResponse(c bookings.Created) CreatedResponse {
	return CreatedResponse{
		ReservationUID: c.ReservationUID, HotelUID: c.HotelUID, StartDate: c.StartDate,
		EndDate: c.EndDate, Discount: c.Discount, Status: string(c.Status), Payment: paymentResponse(c.Payment),
	}
}

type UserInfoResponse struct {
	Reservations []ReservationResponse `json:"reservations"`
	Loyalty      LoyaltyResponse       `json:"loyalty"`
}

func userInfoResponse(u bookings.UserInfo) UserInfoResponse {
	return UserInfoResponse{Reservations: reservationResponses(u.Reservations), Loyalty: loyaltyResponse(u.Loyalty)}
}
