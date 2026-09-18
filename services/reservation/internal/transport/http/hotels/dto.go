package hotels

import "lab2/reservation/internal/domain/hotels"

type ListInput struct {
	Page string `query:"page"`
	Size string `query:"size"`
}

type ListOutput struct {
	Body hotels.Page
}

type GetInput struct {
	UID string `path:"hotelUid"`
}

type GetOutput struct {
	Body hotels.Hotel
}
