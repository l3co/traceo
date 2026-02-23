package homeless

import (
	"context"

	"github.com/l3co/traceo-api/internal/domain/shared"
)

type GenderStat struct {
	Gender shared.Gender
	Count  int64
}

type Repository interface {
	Create(ctx context.Context, h *Homeless) error
	FindByID(ctx context.Context, id string) (*Homeless, error)
	FindAll(ctx context.Context) ([]*Homeless, error)
	Count(ctx context.Context) (int64, error)
	CountByGender(ctx context.Context) ([]GenderStat, error)
}
