package hotels

import (
	"context"
	"errors"
	"strconv"
	"uuid"

	"github.com/samber/oops"
	"lab2/reservation/internal/domain"
	"lab2/reservation/internal/domain/hotels"
)

type PaginationConfig struct {
	DefaultPage int `mapstructure:"default_page"`
	DefaultSize int `mapstructure:"default_size"`
	MaxSize     int `mapstructure:"max_size"`
}

func (c PaginationConfig) Validate() error {
	invalidDefaults := c.DefaultPage < 1 || c.DefaultSize < 1
	if invalidDefaults {
		return oops.Errorf("pagination defaults must be positive")
	}
	invalidMax := c.MaxSize < c.DefaultSize
	offsetOverflow := int64(c.DefaultPage-1) > (1<<31-1)/int64(c.DefaultSize)
	if invalidMax || offsetOverflow {
		return oops.In("configuration").Code("invalid_pagination").Errorf("invalid pagination configuration")
	}
	return nil
}

type Service struct {
	repo       Repository
	pagination PaginationConfig
}

func New(repo Repository, cfg PaginationConfig) *Service {
	return &Service{repo: repo, pagination: cfg}
}

func (s *Service) List(ctx context.Context, pageText, sizeText string) (hotels.Page, error) {
	page, size := s.pagination.DefaultPage, s.pagination.DefaultSize
	var err error
	if pageText != "" {
		page, err = strconv.Atoi(pageText)
		if err != nil {
			return hotels.Page{}, oops.In("validation").Code("invalid_input").Public("invalid page").
				Wrap(errors.Join(domain.ErrInvalidInput, err))
		}
	}
	if sizeText != "" {
		size, err = strconv.Atoi(sizeText)
		if err != nil {
			return hotels.Page{}, oops.In("validation").Code("invalid_input").Public("invalid size").
				Wrap(errors.Join(domain.ErrInvalidInput, err))
		}
	}
	invalidBounds := page < 1 || size < 1
	if invalidBounds {
		return hotels.Page{}, oops.In("validation").Code("invalid_input").
			Public("invalid page or size").
			Wrap(domain.ErrInvalidInput)
	}
	tooLarge := size > s.pagination.MaxSize
	offsetOverflow := int64(page-1) > (1<<31-1)/int64(size)
	if tooLarge || offsetOverflow {
		return hotels.Page{}, oops.In("validation").Code("invalid_input").
			Public("invalid page or size").
			Wrap(domain.ErrInvalidInput)
	}
	result, err := s.repo.List(ctx, page, size)
	return result, oops.FromContext(ctx).In("hotels.usecase").With("operation", "list").Wrapf(err, "operation failed")
}

func (s *Service) Get(ctx context.Context, id string) (hotels.Hotel, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return hotels.Hotel{}, oops.In("validation").Code("invalid_input").
			Public("invalid hotel UUID").
			Wrap(errors.Join(domain.ErrInvalidInput, err))
	}
	result, err := s.repo.Get(ctx, uid)
	return result, oops.FromContext(ctx).In("hotels.usecase").With("operation", "get").Wrapf(err, "operation failed")
}
