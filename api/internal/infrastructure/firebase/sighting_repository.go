package firebase

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/l3co/traceo-api/internal/domain/sighting"
)

const sightingCollection = "sightings"

type SightingRepository struct {
	client *firestore.Client
}

func NewSightingRepository(client *firestore.Client) *SightingRepository {
	return &SightingRepository{client: client}
}

type sightingDoc struct {
	ID          string    `firestore:"id"`
	MissingID   string    `firestore:"missing_id"`
	Lat         float64   `firestore:"lat"`
	Lng         float64   `firestore:"lng"`
	Observation string    `firestore:"observation"`
	CreatedAt   time.Time `firestore:"created_at"`
}

func toSightingDoc(s *sighting.Sighting) sightingDoc {
	return sightingDoc{
		ID:          s.ID,
		MissingID:   s.MissingID,
		Lat:         s.Location.Lat,
		Lng:         s.Location.Lng,
		Observation: s.Observation,
		CreatedAt:   s.CreatedAt,
	}
}

func toSightingEntity(d sightingDoc) *sighting.Sighting {
	return &sighting.Sighting{
		ID:        d.ID,
		MissingID: d.MissingID,
		Location: sighting.GeoPoint{
			Lat: d.Lat,
			Lng: d.Lng,
		},
		Observation: d.Observation,
		CreatedAt:   d.CreatedAt,
	}
}

func (r *SightingRepository) Create(ctx context.Context, s *sighting.Sighting) error {
	_, err := r.client.Collection(sightingCollection).Doc(s.ID).Set(ctx, toSightingDoc(s))
	if err != nil {
		return fmt.Errorf("firestore: creating sighting: %w", err)
	}
	return nil
}

func (r *SightingRepository) FindByID(ctx context.Context, id string) (*sighting.Sighting, error) {
	doc, err := r.client.Collection(sightingCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, sighting.ErrSightingNotFound
		}
		return nil, fmt.Errorf("firestore: finding sighting: %w", err)
	}

	var d sightingDoc
	if err := doc.DataTo(&d); err != nil {
		return nil, fmt.Errorf("firestore: decoding sighting: %w", err)
	}

	return toSightingEntity(d), nil
}

func (r *SightingRepository) FindByMissingID(ctx context.Context, missingID string) ([]*sighting.Sighting, error) {
	docs, err := r.client.Collection(sightingCollection).
		Where("missing_id", "==", missingID).
		OrderBy("created_at", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("firestore: finding sightings by missing_id: %w", err)
	}

	result := make([]*sighting.Sighting, 0, len(docs))
	for _, doc := range docs {
		var d sightingDoc
		if err := doc.DataTo(&d); err != nil {
			continue
		}
		result = append(result, toSightingEntity(d))
	}

	return result, nil
}
