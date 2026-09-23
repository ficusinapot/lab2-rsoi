package loyalties

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/samber/oops"
	"lab2/loyalty/internal/domain"
	"lab2/loyalty/internal/domain/loyalties"
)

type Repository interface {
	Get(context.Context, string) (loyalties.Loyalty, error)
	Change(context.Context, string, int) (loyalties.Loyalty, error)
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func validate(name string) error {
	if strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return oops.Wrap(domain.ErrInvalidInput)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, name string) (loyalties.Loyalty, error) {
	if err := validate(name); err != nil {
		return loyalties.Loyalty{}, err
	}
	item, err := s.repo.Get(ctx, name)
	return item, oops.Wrapf(err, "get loyalty")
}

func (s *Service) Change(ctx context.Context, name string, delta int) (loyalties.Loyalty, error) {
	if err := validate(name); err != nil {
		return loyalties.Loyalty{}, err
	}
	if delta != 1 && delta != -1 {
		return loyalties.Loyalty{}, oops.Wrap(domain.ErrInvalidInput)
	}
	item, err := s.repo.Change(ctx, name, delta)
	return item, oops.Wrapf(err, "change loyalty")
}
