package loyalties

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/oops"
	"lab2/loyalty/internal/domain"
	model "lab2/loyalty/internal/domain/loyalties"
)

type repositoryStub struct {
	Repository
	count   int
	called  bool
	failure error
}

func (r *repositoryStub) Get(context.Context, string) (model.Loyalty, error) {
	r.called = true
	return model.ForCount(r.count), r.failure
}

func (r *repositoryStub) Change(_ context.Context, _ string, delta int) (model.Loyalty, error) {
	r.called = true
	r.count = max(0, r.count+delta)
	return model.ForCount(r.count), r.failure
}

func TestChanges(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		start, delta, count, discount int
		status                        model.Status
	}{
		{9, 1, 10, 7, "SILVER"},
		{19, 1, 20, 10, "GOLD"},
		{10, -1, 9, 5, "BRONZE"},
		{20, -1, 19, 7, "SILVER"},
		{0, -1, 0, 5, "BRONZE"},
	} {
		repo := &repositoryStub{count: tc.start}
		item, err := New(repo).Change(t.Context(), "user", tc.delta)
		if err != nil || item.ReservationCount != tc.count || item.Discount != tc.discount || item.Status != tc.status {
			t.Fatalf("%+v: %+v %v", tc, item, err)
		}
	}
}

func TestValidationAndErrorWrapping(t *testing.T) {
	t.Parallel()
	repo := &repositoryStub{}
	s := New(repo)
	if _, err := s.Get(t.Context(), " "); !errors.Is(err, domain.ErrInvalidInput) || repo.called {
		t.Fatal(err)
	}
	if _, err := s.Change(t.Context(), "user", 2); !errors.Is(err, domain.ErrInvalidInput) || repo.called {
		t.Fatal(err)
	}
	cause := oops.Errorf("storage failed")
	repo.failure = cause
	if _, err := s.Get(t.Context(), "user"); !errors.Is(err, cause) {
		t.Fatal(err)
	}
}
