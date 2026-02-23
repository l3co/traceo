package matching_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/matching"
	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/domain/shared"
)

// --- Mock FaceComparer ---

type mockComparer struct {
	score float64
}

func (m *mockComparer) CompareFaces(_ context.Context, _, _ string) (*matching.FaceComparisonResult, error) {
	return &matching.FaceComparisonResult{
		SimilarityScore:   m.score,
		Analysis:          "test analysis",
		MatchingFeatures:  []string{"eyes", "nose"},
		DifferentFeatures: []string{"hair"},
		Confidence:        "medium",
	}, nil
}

// --- Mock Notifier ---

type mockNotifier struct {
	mu    sync.Mutex
	calls int
}

func (m *mockNotifier) NotifySighting(_ context.Context, _, _ string) error { return nil }
func (m *mockNotifier) NotifyNewHomeless(_ context.Context, _, _, _, _ string) error {
	return nil
}
func (m *mockNotifier) NotifyPotentialMatch(_ context.Context, _ string, _ float64, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return nil
}

// --- Mock MatchRepository ---

type mockMatchRepo struct {
	items []*matching.Match
}

func (m *mockMatchRepo) Create(_ context.Context, match *matching.Match) error {
	m.items = append(m.items, match)
	return nil
}

func (m *mockMatchRepo) FindByID(_ context.Context, id string) (*matching.Match, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, matching.ErrMatchNotFound
}

func (m *mockMatchRepo) FindByHomelessID(_ context.Context, hid string) ([]*matching.Match, error) {
	var result []*matching.Match
	for _, item := range m.items {
		if item.HomelessID == hid {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockMatchRepo) FindByMissingID(_ context.Context, mid string) ([]*matching.Match, error) {
	var result []*matching.Match
	for _, item := range m.items {
		if item.MissingID == mid {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockMatchRepo) UpdateStatus(_ context.Context, id string, status matching.MatchStatus) error {
	for _, item := range m.items {
		if item.ID == id {
			item.Status = status
			return nil
		}
	}
	return matching.ErrMatchNotFound
}

// --- Mock HomelessRepo ---

type mockHomelessRepo struct {
	items []*homeless.Homeless
}

func (m *mockHomelessRepo) Create(_ context.Context, h *homeless.Homeless) error { return nil }
func (m *mockHomelessRepo) FindByID(_ context.Context, id string) (*homeless.Homeless, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, homeless.ErrHomelessNotFound
}
func (m *mockHomelessRepo) FindAll(_ context.Context) ([]*homeless.Homeless, error) { return nil, nil }
func (m *mockHomelessRepo) Count(_ context.Context) (int64, error)                  { return 0, nil }
func (m *mockHomelessRepo) CountByGender(_ context.Context) ([]homeless.GenderStat, error) {
	return nil, nil
}

// --- Mock MissingRepo ---

type mockMissingRepo struct {
	items []*missing.Missing
}

func (m *mockMissingRepo) Create(_ context.Context, mi *missing.Missing) error { return nil }
func (m *mockMissingRepo) FindByID(_ context.Context, id string) (*missing.Missing, error) {
	return nil, nil
}
func (m *mockMissingRepo) Update(_ context.Context, mi *missing.Missing) error { return nil }
func (m *mockMissingRepo) Delete(_ context.Context, id string) error           { return nil }
func (m *mockMissingRepo) FindByUserID(_ context.Context, uid string) ([]*missing.Missing, error) {
	return nil, nil
}
func (m *mockMissingRepo) FindAll(_ context.Context, opts missing.ListOptions) ([]*missing.Missing, string, error) {
	return nil, "", nil
}
func (m *mockMissingRepo) Count(_ context.Context) (int64, error) { return 0, nil }
func (m *mockMissingRepo) Search(_ context.Context, q string, l int) ([]*missing.Missing, error) {
	return nil, nil
}
func (m *mockMissingRepo) CountByGender(_ context.Context) ([]missing.GenderStat, error) {
	return nil, nil
}
func (m *mockMissingRepo) CountByYear(_ context.Context) ([]missing.YearStat, error) {
	return nil, nil
}
func (m *mockMissingRepo) CountChildren(_ context.Context) (int64, error) { return 0, nil }
func (m *mockMissingRepo) FindLocations(_ context.Context, l int) ([]missing.LocationPoint, error) {
	return nil, nil
}
func (m *mockMissingRepo) UpdateAgeProgressionURLs(_ context.Context, _ string, _ []string) error {
	return nil
}
func (m *mockMissingRepo) FindCandidates(_ context.Context, _ missing.CandidateFilter) ([]*missing.Missing, error) {
	return m.items, nil
}

// --- Tests ---

func TestProcessFaceMatching_NoPhoto(t *testing.T) {
	hRepo := &mockHomelessRepo{items: []*homeless.Homeless{
		{ID: "h1", Name: "Carlos", Gender: shared.GenderMale},
	}}
	svc := matching.NewService(&mockMissingRepo{}, hRepo, &mockMatchRepo{}, &mockComparer{score: 0.9}, nil, nil)

	err := svc.ProcessFaceMatching(context.Background(), "h1")
	require.NoError(t, err)
}

func TestProcessFaceMatching_ScoreBelow06_NoMatch(t *testing.T) {
	hRepo := &mockHomelessRepo{items: []*homeless.Homeless{
		{ID: "h1", Name: "Carlos", PhotoURL: "http://photo1.jpg", Gender: shared.GenderMale, Skin: shared.SkinBrown, BirthDate: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)},
	}}
	mRepo := &mockMissingRepo{items: []*missing.Missing{
		{ID: "m1", Name: "João", PhotoURL: "http://photo2.jpg", Gender: shared.GenderMale, Skin: shared.SkinBrown},
	}}
	matchRepo := &mockMatchRepo{}

	svc := matching.NewService(mRepo, hRepo, matchRepo, &mockComparer{score: 0.4}, nil, nil)

	err := svc.ProcessFaceMatching(context.Background(), "h1")
	require.NoError(t, err)
	assert.Len(t, matchRepo.items, 0)
}

func TestProcessFaceMatching_Score06_SavesMatch(t *testing.T) {
	hRepo := &mockHomelessRepo{items: []*homeless.Homeless{
		{ID: "h1", Name: "Carlos", PhotoURL: "http://photo1.jpg", Gender: shared.GenderMale, Skin: shared.SkinBrown, BirthDate: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)},
	}}
	mRepo := &mockMissingRepo{items: []*missing.Missing{
		{ID: "m1", Name: "João", PhotoURL: "http://photo2.jpg", Gender: shared.GenderMale, Skin: shared.SkinBrown},
	}}
	matchRepo := &mockMatchRepo{}

	svc := matching.NewService(mRepo, hRepo, matchRepo, &mockComparer{score: 0.7}, nil, nil)

	err := svc.ProcessFaceMatching(context.Background(), "h1")
	require.NoError(t, err)
	assert.Len(t, matchRepo.items, 1)
	assert.Equal(t, matching.MatchStatusPending, matchRepo.items[0].Status)
}

func TestProcessFaceMatching_Score08_NotifiesAndSaves(t *testing.T) {
	hRepo := &mockHomelessRepo{items: []*homeless.Homeless{
		{ID: "h1", Name: "Carlos", PhotoURL: "http://photo1.jpg", Gender: shared.GenderMale, Skin: shared.SkinBrown, BirthDate: time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)},
	}}
	mRepo := &mockMissingRepo{items: []*missing.Missing{
		{ID: "m1", Name: "João", PhotoURL: "http://photo2.jpg", Gender: shared.GenderMale, Skin: shared.SkinBrown},
	}}
	matchRepo := &mockMatchRepo{}
	notifier := &mockNotifier{}

	svc := matching.NewService(mRepo, hRepo, matchRepo, &mockComparer{score: 0.85}, nil, notifier)

	err := svc.ProcessFaceMatching(context.Background(), "h1")
	require.NoError(t, err)
	assert.Len(t, matchRepo.items, 1)

	time.Sleep(100 * time.Millisecond)
	notifier.mu.Lock()
	assert.Equal(t, 1, notifier.calls)
	notifier.mu.Unlock()
}

func TestProcessFaceMatching_HomelessNotFound(t *testing.T) {
	hRepo := &mockHomelessRepo{}
	svc := matching.NewService(&mockMissingRepo{}, hRepo, &mockMatchRepo{}, &mockComparer{}, nil, nil)

	err := svc.ProcessFaceMatching(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestUpdateStatus_Valid(t *testing.T) {
	matchRepo := &mockMatchRepo{items: []*matching.Match{
		{ID: "match-1", Status: matching.MatchStatusPending},
	}}
	svc := matching.NewService(nil, nil, matchRepo, nil, nil, nil)

	err := svc.UpdateStatus(context.Background(), "match-1", matching.MatchStatusConfirmed)
	require.NoError(t, err)
	assert.Equal(t, matching.MatchStatusConfirmed, matchRepo.items[0].Status)
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	svc := matching.NewService(nil, nil, &mockMatchRepo{}, nil, nil, nil)

	err := svc.UpdateStatus(context.Background(), "match-1", "invalid")
	assert.ErrorIs(t, err, matching.ErrInvalidMatch)
}

func TestEntity_Validate(t *testing.T) {
	m := &matching.Match{HomelessID: "h1", MissingID: "m1", Score: 0.5}
	assert.NoError(t, m.Validate())
}

func TestEntity_Validate_MissingFields(t *testing.T) {
	m := &matching.Match{}
	assert.ErrorIs(t, m.Validate(), matching.ErrInvalidMatch)
}

func TestEntity_Validate_InvalidScore(t *testing.T) {
	m := &matching.Match{HomelessID: "h1", MissingID: "m1", Score: 1.5}
	assert.ErrorIs(t, m.Validate(), matching.ErrInvalidMatch)
}
