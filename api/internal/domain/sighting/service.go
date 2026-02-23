package sighting

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"

	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/domain/notification"
)

type Service struct {
	repo        Repository
	missingRepo missing.Repository
	notifier    notification.Notifier
	sanitizer   *bluemonday.Policy
}

func NewService(repo Repository, missingRepo missing.Repository, notifier notification.Notifier) *Service {
	return &Service{
		repo:        repo,
		missingRepo: missingRepo,
		notifier:    notifier,
		sanitizer:   bluemonday.StrictPolicy(),
	}
}

type CreateInput struct {
	MissingID   string
	Lat         float64
	Lng         float64
	Observation string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Sighting, error) {
	m, err := s.missingRepo.FindByID(ctx, input.MissingID)
	if err != nil {
		return nil, fmt.Errorf("%w: missing person not found", ErrInvalidSighting)
	}

	observation := s.sanitizer.Sanitize(input.Observation)

	sighting := &Sighting{
		ID:        uuid.NewString(),
		MissingID: input.MissingID,
		Location: GeoPoint{
			Lat: input.Lat,
			Lng: input.Lng,
		},
		Observation: observation,
		CreatedAt:   time.Now(),
	}

	if err := sighting.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, sighting); err != nil {
		return nil, fmt.Errorf("creating sighting: %w", err)
	}

	if s.notifier != nil {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			slog.Info("sending sighting notification",
				"missing_id", m.ID,
				"missing_name", m.Name,
				"sighting_id", sighting.ID,
			)

			if err := s.notifier.NotifySighting(bgCtx, m.UserID, observation); err != nil {
				slog.Error("failed to send sighting notification",
					"missing_id", m.ID,
					"error", err,
				)
			} else {
				slog.Info("sighting notification sent",
					"missing_id", m.ID,
					"sighting_id", sighting.ID,
				)
			}
		}()
	}

	return sighting, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*Sighting, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidSighting)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) FindByMissingID(ctx context.Context, missingID string) ([]*Sighting, error) {
	if missingID == "" {
		return nil, fmt.Errorf("%w: missing_id is required", ErrInvalidSighting)
	}
	return s.repo.FindByMissingID(ctx, missingID)
}
