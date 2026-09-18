package bookings

type CreateInput struct {
	HotelUID   string
	PaymentUID string
	StartDate  string
	EndDate    string
}
