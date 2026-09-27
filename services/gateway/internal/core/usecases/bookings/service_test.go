package bookings

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strconv"
	"testing"
	"uuid"

	"github.com/samber/oops"
	model "lab2/gateway/internal/models/entities"
	"lab2/gateway/internal/models/loyaltyifc"
	"lab2/gateway/internal/models/paymentifc"
	"lab2/gateway/internal/models/reservationifc"
)

const (
	testPaid = "PAID"
	testEnd  = "2026-01-04"
	hotelID  = "049161bb-badd-4fa8-9d90-87c9a82b0668"
)

type servicesStub struct {
	reservationifc.Bookings
	paymentifc.Payments
	loyaltyifc.Client
	calls  []string
	fail   string
	status model.ReservationStatus
	price  int
}

func (s *servicesStub) step(name string) error {
	s.calls = append(s.calls, name)
	if s.fail == name {
		return oops.Wrap(errors.New("dependency failure"))
	}
	return nil
}

func (s *servicesStub) Hotel(context.Context, string) (model.Hotel, error) {
	return model.Hotel{Price: 101}, s.step("hotel")
}

func (s *servicesStub) Loyalty(context.Context, string) (model.Loyalty, error) {
	return model.Loyalty{Discount: 7}, s.step("loyalty")
}

func (s *servicesStub) CreatePayment(_ context.Context, price int) (model.InternalPayment, error) {
	s.price = price
	return model.InternalPayment{
		PaymentUID: uuid.New(), Payment: model.Payment{Price: price, Status: testPaid},
	}, s.step("payment")
}

func (s *servicesStub) CreateReservation(
	_ context.Context, _ string, input model.InternalCreate,
) (model.InternalReservation, error) {
	if input.PaymentUID == (uuid.UUID{}) || input.HotelUID != uuid.MustParse(hotelID) {
		return model.InternalReservation{}, oops.Errorf("invalid internal UUID")
	}
	return model.InternalReservation{
		ReservationUID: uuid.New(), HotelUID: input.HotelUID, Status: testPaid,
	}, s.step("reservation")
}

func (s *servicesStub) ChangeLoyalty(_ context.Context, _ string, delta int) error {
	if delta > 0 {
		return s.step("increase")
	}
	return s.step("decrease")
}

func (s *servicesStub) Reservation(context.Context, string, string) (model.InternalReservation, error) {
	return model.InternalReservation{Status: s.status, PaymentUID: uuid.New()}, s.step("get")
}

func (s *servicesStub) CancelReservation(context.Context, string, string) error {
	return s.step("cancel reservation")
}

func (s *servicesStub) CancelPayment(context.Context, string) error { return s.step("cancel payment") }

func TestCreateStopsAtFailure(t *testing.T) {
	t.Parallel()
	steps := []string{"hotel", "loyalty", "payment", "reservation", "increase"}
	for i := 0; i <= len(steps); i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			stub := &servicesStub{}
			want := steps
			if i < len(steps) {
				stub.fail = steps[i]
				want = steps[:i+1]
			}
			out, err := New(stub, stub, stub, 4).Create(t.Context(), "user", model.CreateRequest{
				HotelUID: hotelID, StartDate: "2026-01-01", EndDate: testEnd,
			})
			if (err != nil) != (i < len(steps)) {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(stub.calls, want) {
				t.Fatalf("calls %v, want %v", stub.calls, want)
			}
			if i == len(steps) && (out.Discount != 7 || stub.price != 281) {
				t.Fatalf("wrong discount/cost: %+v price %d", out, stub.price)
			}
		})
	}
}

func TestCancelStopsAtFailure(t *testing.T) {
	t.Parallel()
	steps := []string{"get", "cancel reservation", "cancel payment", "decrease"}
	for i := 0; i <= len(steps); i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			stub := &servicesStub{status: testPaid}
			want := steps
			if i < len(steps) {
				stub.fail = steps[i]
				want = steps[:i+1]
			}
			err := New(stub, stub, stub, 4).Cancel(t.Context(), hotelID, "user")
			if (err != nil) != (i < len(steps)) {
				t.Fatalf("unexpected error %v", err)
			}
			if !reflect.DeepEqual(stub.calls, want) {
				t.Fatalf("calls %v, want %v", stub.calls, want)
			}
		})
	}
}

func TestRepeatCancel(t *testing.T) {
	t.Parallel()
	stub := &servicesStub{status: "CANCELED"}
	if err := New(stub, stub, stub, 4).Cancel(t.Context(), hotelID, "user"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stub.calls, []string{"get"}) {
		t.Fatal(stub.calls)
	}
}

func TestInvalidCreationHasNoSideEffects(t *testing.T) {
	t.Parallel()
	for _, input := range []model.CreateRequest{
		{HotelUID: "invalid", StartDate: "2026-01-01", EndDate: testEnd},
		{HotelUID: hotelID, StartDate: "bad", EndDate: testEnd},
		{HotelUID: hotelID, StartDate: testEnd, EndDate: testEnd},
	} {
		stub := &servicesStub{}
		_, err := New(stub, stub, stub, 4).Create(t.Context(), "user", input)
		if !errors.Is(err, model.ErrInvalidInput) || len(stub.calls) != 0 {
			t.Fatalf("err %v calls %v", err, stub.calls)
		}
	}
	stub := &servicesStub{}
	_, err := New(stub, stub, stub, 4).Create(t.Context(), " ", model.CreateRequest{HotelUID: hotelID})
	if !errors.Is(err, model.ErrInvalidInput) || len(stub.calls) != 0 {
		t.Fatalf("err %v calls %v", err, stub.calls)
	}
}

func TestCost(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		n          int64
		p, d, want int
		bad        bool
	}{
		{3, 10000, 5, 28500, false},
		{3, 10000, 10, 27000, false},
		{3, 101, 7, 281, false},
		{math.MaxInt64, 2, 5, 0, true},
		{2, math.MaxInt32, 5, 0, true},
		{1, -1, 5, 0, true},
		{1, 1, 101, 0, true},
	} {
		got, err := cost(tc.n, tc.p, tc.d)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("%+v: got %d err %v", tc, got, err)
		}
	}
}
