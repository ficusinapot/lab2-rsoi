package bookings

import (
	"context"
	"sync"

	"github.com/samber/oops"
	"lab2/gateway/internal/domain/bookings"
)

func (s *Service) enrichList(ctx context.Context, rows []bookings.InternalReservation) ([]bookings.Reservation, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	items := make([]bookings.Reservation, len(rows))
	jobs := make(chan int)
	var workers sync.WaitGroup
	var firstError error
	var once sync.Once
	for range min(s.concurrency, len(rows)) {
		workers.Go(func() {
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				item, err := s.enrich(ctx, rows[index])
				if err != nil {
					once.Do(func() {
						firstError = err
						cancel()
					})
					return
				}
				items[index] = item
			}
		})
	}
dispatch:
	for index := range rows {
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- index:
		}
	}
	close(jobs)
	workers.Wait()
	if firstError != nil {
		return nil, firstError
	}
	if err := ctx.Err(); err != nil {
		return nil, oops.Wrapf(err, "enrich reservations")
	}
	return items, nil
}
