package missing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
)

var sanitizer = bluemonday.StrictPolicy()

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input *CreateInput) (*Missing, error) {
	m := &Missing{
		ID:                  uuid.NewString(),
		UserID:              input.UserID,
		Name:                sanitizer.Sanitize(input.Name),
		Nickname:            sanitizer.Sanitize(input.Nickname),
		BirthDate:           input.BirthDate,
		DateOfDisappearance: input.DateOfDisappearance,
		Height:              sanitizer.Sanitize(input.Height),
		Clothes:             sanitizer.Sanitize(input.Clothes),
		Gender:              input.Gender,
		Eyes:                input.Eyes,
		Hair:                input.Hair,
		Skin:                input.Skin,
		PhotoURL:            input.PhotoURL,
		Location:            input.Location,
		Status:              StatusDisappeared,
		EventReport:         sanitizer.Sanitize(input.EventReport),
		TattooDescription:   sanitizer.Sanitize(input.TattooDescription),
		ScarDescription:     sanitizer.Sanitize(input.ScarDescription),
		Timestamps: Timestamps{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	m.CalculateWasChild()
	m.GenerateSlug()

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("creating missing person: %w", err)
	}

	return m, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*Missing, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidMissing)
	}

	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, input *UpdateInput) (*Missing, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if m.UserID != userID {
		return nil, fmt.Errorf("%w: not the owner", ErrInvalidMissing)
	}

	m.Name = sanitizer.Sanitize(input.Name)
	m.Nickname = sanitizer.Sanitize(input.Nickname)
	m.BirthDate = input.BirthDate
	m.DateOfDisappearance = input.DateOfDisappearance
	m.Height = sanitizer.Sanitize(input.Height)
	m.Clothes = sanitizer.Sanitize(input.Clothes)
	m.Gender = input.Gender
	m.Eyes = input.Eyes
	m.Hair = input.Hair
	m.Skin = input.Skin
	m.Location = input.Location
	m.EventReport = sanitizer.Sanitize(input.EventReport)
	m.TattooDescription = sanitizer.Sanitize(input.TattooDescription)
	m.ScarDescription = sanitizer.Sanitize(input.ScarDescription)
	m.UpdatedAt = time.Now()

	if input.PhotoURL != "" {
		m.PhotoURL = input.PhotoURL
	}

	if input.Status != "" && input.Status.IsValid() {
		m.Status = input.Status
	}

	m.CalculateWasChild()
	m.GenerateSlug()

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("updating missing person: %w", err)
	}

	return m, nil
}

func (s *Service) UpdateEntity(ctx context.Context, m *Missing) error {
	m.UpdatedAt = time.Now()
	return s.repo.Update(ctx, m)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if m.UserID != userID {
		return fmt.Errorf("%w: not the owner", ErrInvalidMissing)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting missing person: %w", err)
	}

	return nil
}

func (s *Service) FindByUserID(ctx context.Context, userID string) ([]*Missing, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidMissing)
	}

	return s.repo.FindByUserID(ctx, userID)
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Missing, string, error) {
	if opts.PageSize <= 0 || opts.PageSize > 50 {
		opts.PageSize = 20
	}

	return s.repo.FindAll(ctx, opts)
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]*Missing, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: search query is required", ErrInvalidMissing)
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.Search(ctx, query, limit)
}

type DashboardStats struct {
	Total      int64        `json:"total"`
	ByGender   []GenderStat `json:"by_gender"`
	ChildCount int64        `json:"child_count"`
	ByYear     []YearStat   `json:"by_year"`
}

func (s *Service) GetStats(ctx context.Context) (*DashboardStats, error) {
	var (
		wg         sync.WaitGroup
		total      int64
		byGender   []GenderStat
		childCount int64
		byYear     []YearStat
		errCh      = make(chan error, 4)
	)

	wg.Add(4)

	go func() {
		defer wg.Done()
		c, err := s.repo.Count(ctx)
		if err != nil {
			errCh <- fmt.Errorf("counting total: %w", err)
			return
		}
		total = c
	}()

	go func() {
		defer wg.Done()
		gs, err := s.repo.CountByGender(ctx)
		if err != nil {
			errCh <- fmt.Errorf("counting by gender: %w", err)
			return
		}
		byGender = gs
	}()

	go func() {
		defer wg.Done()
		c, err := s.repo.CountChildren(ctx)
		if err != nil {
			errCh <- fmt.Errorf("counting children: %w", err)
			return
		}
		childCount = c
	}()

	go func() {
		defer wg.Done()
		ys, err := s.repo.CountByYear(ctx)
		if err != nil {
			errCh <- fmt.Errorf("counting by year: %w", err)
			return
		}
		byYear = ys
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return &DashboardStats{
		Total:      total,
		ByGender:   byGender,
		ChildCount: childCount,
		ByYear:     byYear,
	}, nil
}

func (s *Service) FindLocations(ctx context.Context, limit int) ([]LocationPoint, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.FindLocations(ctx, limit)
}
