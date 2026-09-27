package hotels

import (
	"errors"
	"strconv"
	"testing"
	"uuid"

	"go.uber.org/mock/gomock"
	hotels "lab2/reservation/internal/models/entities"
)

func TestPaginationConfigValidate(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		cfg   PaginationConfig
		valid bool
	}{
		{name: "valid", cfg: PaginationConfig{DefaultPage: 1, DefaultSize: 10, MaxSize: 100}, valid: true},
		{name: "zero page", cfg: PaginationConfig{DefaultSize: 10, MaxSize: 100}},
		{name: "zero size", cfg: PaginationConfig{DefaultPage: 1, MaxSize: 100}},
		{name: "maximum below default", cfg: PaginationConfig{DefaultPage: 1, DefaultSize: 10, MaxSize: 5}},
		{name: "offset overflow", cfg: PaginationConfig{DefaultPage: 2147483647, DefaultSize: 10, MaxSize: 100}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.cfg.Validate(); (err == nil) != tc.valid {
				t.Fatalf("validation error = %v", err)
			}
		})
	}
}

func TestServiceList(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, page, size   string
		wantPage, wantSize int
		invalid            bool
	}{
		{name: "defaults", wantPage: 2, wantSize: 10},
		{name: "explicit", page: "3", size: "5", wantPage: 3, wantSize: 5},
		{name: "zero page", page: "0", invalid: true},
		{name: "invalid size", size: "abc", invalid: true},
		{name: "oversize", size: "101", invalid: true},
		{name: "offset overflow", page: "2147483647", size: "2", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repo := NewMockHotels(ctrl)
			if !tc.invalid {
				repo.EXPECT().List(gomock.Any(), tc.wantPage, tc.wantSize).
					Return(hotels.Page{Page: tc.wantPage, PageSize: tc.wantSize}, nil)
			}
			result, err := New(repo, PaginationConfig{DefaultPage: 2, DefaultSize: 10, MaxSize: 100}).
				List(t.Context(), tc.page, tc.size)
			if tc.invalid {
				if !errors.Is(err, hotels.ErrInvalidInput) {
					t.Fatalf("invalid input: err=%v", err)
				}
				if tc.name == "invalid size" {
					var cause *strconv.NumError
					if !errors.As(err, &cause) || !errors.Is(err, strconv.ErrSyntax) {
						t.Fatalf("size parse cause lost: %v", err)
					}
				}
				return
			}
			if err != nil || result.Page != tc.wantPage || result.PageSize != tc.wantSize {
				t.Fatalf("result=%+v err=%v", result, err)
			}
		})
	}
}

func TestServiceGet(t *testing.T) {
	t.Parallel()
	const id = "049161bb-badd-4fa8-9d90-87c9a82b0668"
	cause := errors.New("database unavailable")
	ctrl := gomock.NewController(t)
	repo := NewMockHotels(ctrl)
	repo.EXPECT().Get(gomock.Any(), uuid.MustParse(id)).Return(hotels.Hotel{}, cause)
	s := New(repo, PaginationConfig{})
	_, parseErr := uuid.Parse("bad")
	_, err := s.Get(t.Context(), "bad")
	if !errors.Is(err, hotels.ErrInvalidInput) || !errors.Is(err, parseErr) {
		t.Fatalf("validation error = %v", err)
	}
	if _, err := s.Get(t.Context(), id); !errors.Is(err, cause) {
		t.Fatalf("wrapped error = %v", err)
	}
}
