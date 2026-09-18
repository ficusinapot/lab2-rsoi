package hotels

import "uuid"

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
