package missing

import "context"

type GenderStat struct {
	Gender string
	Count  int64
}

type YearStat struct {
	Year  int
	Count int64
}

type LocationPoint struct {
	ID     string
	Name   string
	Lat    float64
	Lng    float64
	Status Status
}

type CandidateFilter struct {
	Gender Gender
	Skin   SkinColor
	MinAge int
	MaxAge int
	Status Status
	Limit  int
}

type Repository interface {
	Create(ctx context.Context, m *Missing) error
	FindByID(ctx context.Context, id string) (*Missing, error)
	Update(ctx context.Context, m *Missing) error
	Delete(ctx context.Context, id string) error
	FindByUserID(ctx context.Context, userID string) ([]*Missing, error)
	FindAll(ctx context.Context, opts ListOptions) ([]*Missing, string, error)
	Count(ctx context.Context) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*Missing, error)
	CountByGender(ctx context.Context) ([]GenderStat, error)
	CountByYear(ctx context.Context) ([]YearStat, error)
	CountChildren(ctx context.Context) (int64, error)
	FindLocations(ctx context.Context, limit int) ([]LocationPoint, error)
	FindCandidates(ctx context.Context, filter CandidateFilter) ([]*Missing, error)
	UpdateAgeProgressionURLs(ctx context.Context, id string, urls []string) error
}
