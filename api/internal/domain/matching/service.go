package matching

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/domain/notification"
)

type FaceComparer interface {
	CompareFaces(ctx context.Context, photo1URL, photo2URL string) (*FaceComparisonResult, error)
}

type FaceComparisonResult struct {
	SimilarityScore   float64
	Analysis          string
	MatchingFeatures  []string
	DifferentFeatures []string
	Confidence        string
}

type Service struct {
	missingRepo  missing.Repository
	homelessRepo homeless.Repository
	matchRepo    Repository
	comparer     FaceComparer
	notifier     notification.Notifier
}

func NewService(
	missingRepo missing.Repository,
	homelessRepo homeless.Repository,
	matchRepo Repository,
	comparer FaceComparer,
	notifier notification.Notifier,
) *Service {
	return &Service{
		missingRepo:  missingRepo,
		homelessRepo: homelessRepo,
		matchRepo:    matchRepo,
		comparer:     comparer,
		notifier:     notifier,
	}
}

func (s *Service) ProcessFaceMatching(ctx context.Context, homelessID string) error {
	h, err := s.homelessRepo.FindByID(ctx, homelessID)
	if err != nil {
		return fmt.Errorf("finding homeless %s: %w", homelessID, err)
	}

	if h.PhotoURL == "" {
		slog.Warn("homeless has no photo, skipping face matching", "id", homelessID)
		return nil
	}

	candidates, err := s.missingRepo.FindCandidates(ctx, missing.CandidateFilter{
		Gender: h.Gender,
		Skin:   h.Skin,
		MinAge: h.Age() - 15,
		MaxAge: h.Age() + 15,
		Status: missing.StatusDisappeared,
		Limit:  20,
	})
	if err != nil {
		return fmt.Errorf("finding candidates: %w", err)
	}

	slog.Info("face matching started",
		"homeless_id", homelessID,
		"candidates", len(candidates),
	)

	for _, candidate := range candidates {
		if candidate.PhotoURL == "" {
			continue
		}

		comparison, err := s.comparer.CompareFaces(ctx, h.PhotoURL, candidate.PhotoURL)
		if err != nil {
			slog.Error("face comparison failed",
				"homeless_id", homelessID,
				"missing_id", candidate.ID,
				"error", err.Error(),
			)
			continue
		}

		slog.Info("face comparison result",
			"homeless_id", homelessID,
			"missing_id", candidate.ID,
			"score", comparison.SimilarityScore,
		)

		if comparison.SimilarityScore >= 0.6 {
			match := &Match{
				ID:             uuid.NewString(),
				HomelessID:     homelessID,
				MissingID:      candidate.ID,
				Score:          comparison.SimilarityScore,
				Status:         MatchStatusPending,
				GeminiAnalysis: comparison.Analysis,
				CreatedAt:      time.Now(),
			}

			if err := s.matchRepo.Create(ctx, match); err != nil {
				slog.Error("saving match failed", "error", err.Error())
				continue
			}

			if comparison.SimilarityScore >= 0.8 && s.notifier != nil {
				go func(missingName string, score float64, analysis string) {
					bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					if err := s.notifier.NotifyPotentialMatch(bgCtx, missingName, score, analysis); err != nil {
						slog.Error("match notification failed", "error", err.Error())
					}
				}(candidate.Name, comparison.SimilarityScore, comparison.Analysis)
			}
		}
	}

	return nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*Match, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidMatch)
	}
	return s.matchRepo.FindByID(ctx, id)
}

func (s *Service) FindByHomelessID(ctx context.Context, homelessID string) ([]*Match, error) {
	return s.matchRepo.FindByHomelessID(ctx, homelessID)
}

func (s *Service) FindByMissingID(ctx context.Context, missingID string) ([]*Match, error) {
	return s.matchRepo.FindByMissingID(ctx, missingID)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status MatchStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidMatch, status)
	}
	return s.matchRepo.UpdateStatus(ctx, id, status)
}

func (s *Service) ProcessAgeProgression(ctx context.Context, missingID, photoURL string, birthDate time.Time) error {
	slog.Info("age progression not yet implemented",
		"missing_id", missingID,
	)
	return nil
}
