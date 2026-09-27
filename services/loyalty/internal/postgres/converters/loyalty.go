package converters

import (
	"lab2/loyalty/ent"
	"lab2/loyalty/internal/models/entities"
)

func Loyalty(row *ent.Loyalty) entities.Loyalty {
	return entities.Loyalty{
		Status: entities.Status(row.Status), Discount: row.Discount, ReservationCount: row.ReservationCount,
	}
}
