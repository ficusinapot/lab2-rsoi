package bookings

import (
	"math"
	"time"

	"lab2/gateway/internal/domain"
	"lab2/gateway/internal/domain/bookings"

	"github.com/samber/oops"
)

func dates(input bookings.CreateRequest) (int64, error) {
	start, e1 := time.Parse(time.DateOnly, input.StartDate)
	end, e2 := time.Parse(time.DateOnly, input.EndDate)
	if e1 != nil || e2 != nil || !end.After(start) {
		return 0, oops.Wrap(domain.ErrInvalidInput)
	}
	return (end.Unix() - start.Unix()) / 86400, nil
}

func cost(nights int64, price, discount int) (int, error) {
	if nights <= 0 || price < 0 || discount < 0 || discount > 100 {
		return 0, oops.Wrap(domain.ErrInvalidInput)
	}
	factor := int64(100 - discount)
	if price > 0 && nights > math.MaxInt64/int64(price) {
		return 0, oops.Wrap(domain.ErrInvalidInput)
	}
	base := nights * int64(price)
	if factor > 0 && base > math.MaxInt64/factor {
		return 0, oops.Wrap(domain.ErrInvalidInput)
	}
	total := base * factor / 100
	if total > math.MaxInt32 {
		return 0, oops.Wrap(domain.ErrInvalidInput)
	}
	return int(total), nil
}
