package entities

type Status string

const (
	StatusBronze Status = "BRONZE"
	StatusSilver Status = "SILVER"
	StatusGold   Status = "GOLD"
)

type Loyalty struct {
	Status           Status
	Discount         int
	ReservationCount int
}

func ForCount(count int) Loyalty {
	item := Loyalty{Status: StatusBronze, Discount: 5, ReservationCount: count}
	if count >= 20 {
		item.Status = StatusGold
		item.Discount = 10
	} else if count >= 10 {
		item.Status = StatusSilver
		item.Discount = 7
	}
	return item
}
