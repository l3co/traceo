package homeless

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"

	"github.com/l3co/traceo-api/internal/domain/notification"
	"github.com/l3co/traceo-api/internal/domain/shared"
)

type Service struct {
	repo      Repository
	notifier  notification.Notifier
	sanitizer *bluemonday.Policy
}

func NewService(repo Repository, notifier notification.Notifier) *Service {
	return &Service{
		repo:      repo,
		notifier:  notifier,
		sanitizer: bluemonday.StrictPolicy(),
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Homeless, error) {
	now := time.Now()

	h := &Homeless{
		ID:        uuid.NewString(),
		Name:      s.sanitizer.Sanitize(input.Name),
		Nickname:  s.sanitizer.Sanitize(input.Nickname),
		BirthDate: input.BirthDate,
		Gender:    input.Gender,
		Eyes:      input.Eyes,
		Hair:      input.Hair,
		Skin:      input.Skin,
		PhotoURL:  input.PhotoURL,
		Location: shared.GeoPoint{
			Lat:     input.Lat,
			Lng:     input.Lng,
			Address: input.Address,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.GenerateSlug()

	if err := h.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, h); err != nil {
		return nil, fmt.Errorf("creating homeless: %w", err)
	}

	if s.notifier != nil {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			birthStr := ""
			if !h.BirthDate.IsZero() {
				birthStr = h.BirthDate.Format("02/01/2006")
			}

			slog.Info("notifying new homeless registration",
				"id", h.ID,
				"name", h.Name,
			)

			if err := s.notifier.NotifyNewHomeless(bgCtx, h.Name, birthStr, h.PhotoURL, h.ID); err != nil {
				slog.Error("failed to notify new homeless",
					"id", h.ID,
					"error", err,
				)
			}
		}()
	}

	return h, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*Homeless, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidHomeless)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) FindAll(ctx context.Context) ([]*Homeless, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *Service) CountByGender(ctx context.Context) ([]GenderStat, error) {
	return s.repo.CountByGender(ctx)
}
