package converters

import (
	"testing"
	"uuid"

	"lab2/payment/ent"
	entpayment "lab2/payment/ent/payment"
	"lab2/payment/internal/models/entities"
)

func TestPayment(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	got := Payment(&ent.Payment{PaymentUID: id, Status: entpayment.StatusPAID, Price: 27000})
	if got.PaymentUID != id || got.Status != entities.StatusPaid || got.Price != 27000 {
		t.Fatalf("payment: %+v", got)
	}
}
