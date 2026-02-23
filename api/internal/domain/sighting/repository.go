package sighting

import "context"

type Repository interface {
	Create(ctx context.Context, s *Sighting) error
	FindByID(ctx context.Context, id string) (*Sighting, error)
	FindByMissingID(ctx context.Context, missingID string) ([]*Sighting, error)
}
