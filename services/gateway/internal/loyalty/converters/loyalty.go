package converters

import "lab2/gateway/internal/models/entities"

type Loyalty struct {
	Status           string `json:"status"`
	Discount         int    `json:"discount"`
	ReservationCount int    `json:"reservationCount"`
}

func (l Loyalty) Entity() entities.Loyalty {
	return entities.Loyalty{
		Status: entities.LoyaltyStatus(l.Status), Discount: l.Discount, ReservationCount: l.ReservationCount,
	}
}
