package matching

import "context"

type Repository interface {
	Create(ctx context.Context, m *Match) error
	FindByID(ctx context.Context, id string) (*Match, error)
	FindByHomelessID(ctx context.Context, homelessID string) ([]*Match, error)
	FindByMissingID(ctx context.Context, missingID string) ([]*Match, error)
	UpdateStatus(ctx context.Context, id string, status MatchStatus) error
}
