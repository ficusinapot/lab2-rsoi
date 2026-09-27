package converters

import (
	"uuid"

	"lab2/gateway/internal/models/entities"
)

type Payment struct {
	PaymentUID uuid.UUID `json:"paymentUid"`
	Status     string    `json:"status"`
	Price      int       `json:"price"`
}

func (p Payment) Entity() entities.Payment {
	return entities.Payment{Status: entities.PaymentStatus(p.Status), Price: p.Price}
}

func (p Payment) InternalEntity() entities.InternalPayment {
	return entities.InternalPayment{Payment: p.Entity(), PaymentUID: p.PaymentUID}
}

type CreateRequest struct {
	Price int `json:"price"`
}
