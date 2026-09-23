package bookings

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/samber/oops"
	model "lab2/gateway/internal/domain/bookings"
)

type readStub struct {
	entered chan struct{}
	ReservationService
	PaymentService
	rows   []model.InternalReservation
	active atomic.Int32
	peak   atomic.Int32
	delay  time.Duration
	fail   bool
}

func (s *readStub) Reservations(context.Context, string) ([]model.InternalReservation, error) {
	return s.rows, nil
}

func (s *readStub) Hotel(ctx context.Context, id string) (model.Hotel, error) {
	if s.entered != nil {
		s.entered <- struct{}{}
	}
	active := s.active.Add(1)
	defer s.active.Add(-1)
	for old := s.peak.Load(); active > old; old = s.peak.Load() {
		if s.peak.CompareAndSwap(old, active) {
			break
		}
	}
	if s.fail {
		return model.Hotel{}, oops.Errorf("hotel unavailable")
	}
	timer := time.NewTimer(s.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return model.Hotel{}, oops.Wrap(ctx.Err())
	case <-timer.C:
	}
	return model.Hotel{Name: id}, nil
}

func (*readStub) Payment(context.Context, string) (model.Payment, error) {
	return model.Payment{Status: model.PaymentPaid}, nil
}

func readRows(count int) []model.InternalReservation {
	rows := make([]model.InternalReservation, count)
	for index := range rows {
		rows[index] = model.InternalReservation{ReservationUID: uuid.New(), HotelUID: uuid.New()}
	}
	return rows
}

func TestListOrderingAndLimit(t *testing.T) {
	t.Parallel()
	stub := &readStub{rows: readRows(20), delay: time.Millisecond}
	items, err := New(stub, stub, nil, 4).List(t.Context(), "user")
	if err != nil || len(items) != len(stub.rows) {
		t.Fatalf("items %d: %v", len(items), err)
	}
	for index, item := range items {
		if item.ReservationUID != stub.rows[index].ReservationUID || item.Hotel.Name != stub.rows[index].HotelUID.String() {
			t.Fatalf("wrong order at %d", index)
		}
	}
	if stub.peak.Load() > 4 || stub.active.Load() != 0 {
		t.Fatalf("peak %d active %d", stub.peak.Load(), stub.active.Load())
	}
}

func TestListCancellationAndFailure(t *testing.T) {
	t.Parallel()
	for _, fail := range []bool{false, true} {
		stub := &readStub{rows: readRows(20), delay: time.Second, fail: fail}
		ctx, cancel := context.WithCancel(t.Context())
		if !fail {
			cancel()
		}
		items, err := New(stub, stub, nil, 4).List(ctx, "user")
		cancel()
		if err == nil || items != nil || stub.active.Load() != 0 {
			t.Fatalf("items %v error %v active %d", items, err, stub.active.Load())
		}
		if !fail && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
}

type paginationStub struct {
	ReservationService
	query string
}

func (s *paginationStub) Hotels(_ context.Context, query string) (model.Page, error) {
	s.query = query
	return model.Page{Page: 1}, nil
}

func TestPaginationNormalization(t *testing.T) {
	t.Parallel()
	for _, page := range []string{"0", "00"} {
		stub := &paginationStub{}
		item, err := New(stub, nil, nil, 4).Hotels(t.Context(), page, "010")
		if err != nil || item.Page != 0 || stub.query != "?page=1&size=10" {
			t.Fatalf("%+v %q %v", item, stub.query, err)
		}
	}
}

func BenchmarkList(b *testing.B) {
	for _, count := range []int{1, 10, 50} {
		for _, limit := range []int{1, 4} {
			b.Run(fmt.Sprintf("rows=%d/workers=%d", count, limit), func(b *testing.B) {
				stub := &readStub{rows: readRows(count), delay: time.Millisecond}
				service := New(stub, stub, nil, limit)
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					if _, err := service.List(b.Context(), "user"); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(stub.peak.Load()), "peak_calls")
			})
		}
	}
}

func TestListCancellationDuringRequests(t *testing.T) {
	t.Parallel()
	stub := &readStub{rows: readRows(20), delay: time.Second, entered: make(chan struct{}, 4)}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	finished := make(chan error, 1)
	go func() { _, err := New(stub, stub, nil, 4).List(ctx, "user"); finished <- err }()
	for range 4 {
		<-stub.entered
	}
	cancel()
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if stub.active.Load() != 0 {
		t.Fatalf("active workers: %d", stub.active.Load())
	}
}
