package sighting_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/domain/sighting"
)

// --- Mock Notifier ---

type mockNotifier struct {
	mu    sync.Mutex
	calls []string
}

func (m *mockNotifier) NotifySighting(_ context.Context, userEmail, observation string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, userEmail+":"+observation)
	return nil
}

func (m *mockNotifier) NotifyNewHomeless(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *mockNotifier) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// --- Mock Sighting Repository ---

type mockSightingRepo struct {
	items []*sighting.Sighting
}

func (m *mockSightingRepo) Create(_ context.Context, s *sighting.Sighting) error {
	m.items = append(m.items, s)
	return nil
}

func (m *mockSightingRepo) FindByID(_ context.Context, id string) (*sighting.Sighting, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, sighting.ErrSightingNotFound
}

func (m *mockSightingRepo) FindByMissingID(_ context.Context, missingID string) ([]*sighting.Sighting, error) {
	var result []*sighting.Sighting
	for _, item := range m.items {
		if item.MissingID == missingID {
			result = append(result, item)
		}
	}
	return result, nil
}

// --- Mock Missing Repository (minimal) ---

type mockMissingRepo struct {
	items []*missing.Missing
}

func (m *mockMissingRepo) FindByID(_ context.Context, id string) (*missing.Missing, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, missing.ErrMissingNotFound
}

func (m *mockMissingRepo) Create(_ context.Context, mi *missing.Missing) error { return nil }
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

// --- Helpers ---

func newTestService() (*sighting.Service, *mockSightingRepo, *mockMissingRepo, *mockNotifier) {
	sRepo := &mockSightingRepo{}
	mRepo := &mockMissingRepo{
		items: []*missing.Missing{
			{
				ID:     "missing-1",
				UserID: "user-1",
				Name:   "João Silva",
			},
		},
	}
	notifier := &mockNotifier{}
	svc := sighting.NewService(sRepo, mRepo, notifier)
	return svc, sRepo, mRepo, notifier
}

func validSightingInput() sighting.CreateInput {
	return sighting.CreateInput{
		MissingID:   "missing-1",
		Lat:         -23.5505,
		Lng:         -46.6333,
		Observation: "Seen near bus station",
	}
}

// --- Tests: Create ---

func TestCreate_Success(t *testing.T) {
	svc, repo, _, _ := newTestService()

	result, err := svc.Create(context.Background(), validSightingInput())

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "missing-1", result.MissingID)
	assert.Equal(t, -23.5505, result.Location.Lat)
	assert.Equal(t, "Seen near bus station", result.Observation)
	assert.Len(t, repo.items, 1)
}

func TestCreate_SanitizesObservation(t *testing.T) {
	svc, _, _, _ := newTestService()

	input := validSightingInput()
	input.Observation = "<script>alert('xss')</script>Seen downtown"

	result, err := svc.Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "Seen downtown", result.Observation)
}

func TestCreate_MissingNotFound(t *testing.T) {
	svc, _, _, _ := newTestService()

	input := validSightingInput()
	input.MissingID = "nonexistent"

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, sighting.ErrInvalidSighting)
}

func TestCreate_MissingLocation(t *testing.T) {
	svc, _, _, _ := newTestService()

	input := validSightingInput()
	input.Lat = 0
	input.Lng = 0

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, sighting.ErrInvalidSighting)
}

func TestCreate_EmptyObservation(t *testing.T) {
	svc, _, _, _ := newTestService()

	input := validSightingInput()
	input.Observation = ""

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, sighting.ErrInvalidSighting)
}

func TestCreate_NotifierCalled(t *testing.T) {
	svc, _, _, notifier := newTestService()

	_, err := svc.Create(context.Background(), validSightingInput())
	require.NoError(t, err)

	// Give goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, notifier.CallCount())
}

// --- Tests: FindByID ---

func TestFindByID_Success(t *testing.T) {
	svc, _, _, _ := newTestService()

	created, _ := svc.Create(context.Background(), validSightingInput())

	found, err := svc.FindByID(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestFindByID_EmptyID(t *testing.T) {
	svc, _, _, _ := newTestService()

	_, err := svc.FindByID(context.Background(), "")

	assert.ErrorIs(t, err, sighting.ErrInvalidSighting)
}

func TestFindByID_NotFound(t *testing.T) {
	svc, _, _, _ := newTestService()

	_, err := svc.FindByID(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, sighting.ErrSightingNotFound)
}

// --- Tests: FindByMissingID ---

func TestFindByMissingID_Success(t *testing.T) {
	svc, _, _, _ := newTestService()

	svc.Create(context.Background(), validSightingInput())
	svc.Create(context.Background(), validSightingInput())

	results, err := svc.FindByMissingID(context.Background(), "missing-1")

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFindByMissingID_EmptyID(t *testing.T) {
	svc, _, _, _ := newTestService()

	_, err := svc.FindByMissingID(context.Background(), "")

	assert.ErrorIs(t, err, sighting.ErrInvalidSighting)
}

func TestFindByMissingID_NoResults(t *testing.T) {
	svc, _, _, _ := newTestService()

	results, err := svc.FindByMissingID(context.Background(), "missing-1")

	require.NoError(t, err)
	assert.Empty(t, results)
}

// --- Tests: Entity Validation ---

func TestSighting_Validate(t *testing.T) {
	s := &sighting.Sighting{
		MissingID:   "m-1",
		Location:    sighting.GeoPoint{Lat: -23.5, Lng: -46.6},
		Observation: "Seen at station",
	}
	assert.NoError(t, s.Validate())
}

func TestSighting_Validate_MissingID(t *testing.T) {
	s := &sighting.Sighting{
		Location:    sighting.GeoPoint{Lat: -23.5, Lng: -46.6},
		Observation: "Seen at station",
	}
	assert.ErrorIs(t, s.Validate(), sighting.ErrInvalidSighting)
}
