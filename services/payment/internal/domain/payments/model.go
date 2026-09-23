package payments

import "uuid"

type Status string

const (
	StatusPaid     Status = "PAID"
	StatusCanceled Status = "CANCELED"
)

type Payment struct {
	PaymentUID uuid.UUID `json:"paymentUid" format:"uuid"`
	Status     Status    `json:"status" enum:"PAID,CANCELED"`
	Price      int       `json:"price"`
}
