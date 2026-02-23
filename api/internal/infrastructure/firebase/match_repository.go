package firebase

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/l3co/traceo-api/internal/domain/matching"
)

const matchCollection = "matches"

type MatchRepository struct {
	client *firestore.Client
}

func NewMatchRepository(client *firestore.Client) *MatchRepository {
	return &MatchRepository{client: client}
}

type matchDoc struct {
	ID             string    `firestore:"id"`
	HomelessID     string    `firestore:"homeless_id"`
	MissingID      string    `firestore:"missing_id"`
	Score          float64   `firestore:"score"`
	Status         string    `firestore:"status"`
	GeminiAnalysis string    `firestore:"gemini_analysis"`
	CreatedAt      time.Time `firestore:"created_at"`
	ReviewedAt     time.Time `firestore:"reviewed_at,omitempty"`
}

func toMatchDoc(m *matching.Match) matchDoc {
	return matchDoc{
		ID:             m.ID,
		HomelessID:     m.HomelessID,
		MissingID:      m.MissingID,
		Score:          m.Score,
		Status:         string(m.Status),
		GeminiAnalysis: m.GeminiAnalysis,
		CreatedAt:      m.CreatedAt,
		ReviewedAt:     m.ReviewedAt,
	}
}

func toMatchEntity(d matchDoc) *matching.Match {
	return &matching.Match{
		ID:             d.ID,
		HomelessID:     d.HomelessID,
		MissingID:      d.MissingID,
		Score:          d.Score,
		Status:         matching.MatchStatus(d.Status),
		GeminiAnalysis: d.GeminiAnalysis,
		CreatedAt:      d.CreatedAt,
		ReviewedAt:     d.ReviewedAt,
	}
}

func (r *MatchRepository) Create(ctx context.Context, m *matching.Match) error {
	_, err := r.client.Collection(matchCollection).Doc(m.ID).Set(ctx, toMatchDoc(m))
	if err != nil {
		return fmt.Errorf("firestore: creating match: %w", err)
	}
	return nil
}

func (r *MatchRepository) FindByID(ctx context.Context, id string) (*matching.Match, error) {
	doc, err := r.client.Collection(matchCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, matching.ErrMatchNotFound
		}
		return nil, fmt.Errorf("firestore: finding match: %w", err)
	}

	var d matchDoc
	if err := doc.DataTo(&d); err != nil {
		return nil, fmt.Errorf("firestore: decoding match: %w", err)
	}

	return toMatchEntity(d), nil
}

func (r *MatchRepository) FindByHomelessID(ctx context.Context, homelessID string) ([]*matching.Match, error) {
	docs, err := r.client.Collection(matchCollection).
		Where("homeless_id", "==", homelessID).
		OrderBy("score", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: finding matches by homeless: %w", err)
	}

	result := make([]*matching.Match, 0, len(docs))
	for _, doc := range docs {
		var d matchDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toMatchEntity(d))
	}
	return result, nil
}

func (r *MatchRepository) FindByMissingID(ctx context.Context, missingID string) ([]*matching.Match, error) {
	docs, err := r.client.Collection(matchCollection).
		Where("missing_id", "==", missingID).
		OrderBy("score", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: finding matches by missing: %w", err)
	}

	result := make([]*matching.Match, 0, len(docs))
	for _, doc := range docs {
		var d matchDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toMatchEntity(d))
	}
	return result, nil
}

func (r *MatchRepository) UpdateStatus(ctx context.Context, id string, status matching.MatchStatus) error {
	_, err := r.client.Collection(matchCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "status", Value: string(status)},
		{Path: "reviewed_at", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("firestore: updating match status: %w", err)
	}
	return nil
}
