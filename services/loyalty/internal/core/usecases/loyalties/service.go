package loyalties

import (
	"context"
	"strings"
	"unicode/utf8"

	"lab2/loyalty/internal/models/dbifc"
	loyalties "lab2/loyalty/internal/models/entities"

	"github.com/samber/oops"
)

type Service struct{ repo dbifc.Loyalties }

func New(repo dbifc.Loyalties) *Service { return &Service{repo: repo} }
func validate(name string) error {
	if strings.TrimSpace(name) == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 80 {
		return oops.Wrap(loyalties.ErrInvalidInput)
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
		return loyalties.Loyalty{}, oops.Wrap(loyalties.ErrInvalidInput)
	}
	item, err := s.repo.Change(ctx, name, delta)
	return item, oops.Wrapf(err, "change loyalty")
}
