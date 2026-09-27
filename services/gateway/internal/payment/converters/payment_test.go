package converters

import (
	"testing"
	"uuid"

	"lab2/gateway/internal/models/entities"
)

func TestPaymentConversion(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	got := (Payment{PaymentUID: id, Status: "PAID", Price: 27000}).InternalEntity()
	if got.PaymentUID != id || got.Status != entities.PaymentPaid || got.Price != 27000 {
		t.Fatalf("payment: %+v", got)
	}
}
