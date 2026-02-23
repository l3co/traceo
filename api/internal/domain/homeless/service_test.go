package homeless_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/shared"
)

// --- Mock Notifier ---

type mockNotifier struct {
	mu    sync.Mutex
	calls []string
}

func (m *mockNotifier) NotifySighting(_ context.Context, _, _ string) error { return nil }

func (m *mockNotifier) NotifyNewHomeless(_ context.Context, name, _, _, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, name)
	return nil
}

func (m *mockNotifier) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// --- Mock Repository ---

type mockRepo struct {
	items []*homeless.Homeless
}

func (m *mockRepo) Create(_ context.Context, h *homeless.Homeless) error {
	m.items = append(m.items, h)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*homeless.Homeless, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, homeless.ErrHomelessNotFound
}

func (m *mockRepo) FindAll(_ context.Context) ([]*homeless.Homeless, error) {
	return m.items, nil
}

func (m *mockRepo) Count(_ context.Context) (int64, error) {
	return int64(len(m.items)), nil
}

func (m *mockRepo) CountByGender(_ context.Context) ([]homeless.GenderStat, error) {
	counts := map[shared.Gender]int64{}
	for _, item := range m.items {
		counts[item.Gender]++
	}
	var result []homeless.GenderStat
	for g, c := range counts {
		result = append(result, homeless.GenderStat{Gender: g, Count: c})
	}
	return result, nil
}

// --- Helpers ---

func validInput() homeless.CreateInput {
	return homeless.CreateInput{
		Name:      "Carlos Souza",
		Nickname:  "Carlão",
		BirthDate: time.Date(1975, 3, 15, 0, 0, 0, 0, time.UTC),
		Gender:    shared.GenderMale,
		Eyes:      shared.EyeBrown,
		Hair:      shared.HairBlack,
		Skin:      shared.SkinBrown,
		PhotoURL:  "https://example.com/photo.jpg",
		Lat:       -23.5505,
		Lng:       -46.6333,
	}
}

// --- Tests: Create ---

func TestCreate_Success(t *testing.T) {
	repo := &mockRepo{}
	notifier := &mockNotifier{}
	svc := homeless.NewService(repo, notifier)

	result, err := svc.Create(context.Background(), validInput())

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "Carlos Souza", result.Name)
	assert.Equal(t, "Carlão", result.Nickname)
	assert.NotEmpty(t, result.Slug)
	assert.Len(t, repo.items, 1)
}

func TestCreate_SanitizesInput(t *testing.T) {
	repo := &mockRepo{}
	svc := homeless.NewService(repo, nil)

	input := validInput()
	input.Name = "<script>xss</script>Carlos"

	result, err := svc.Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "Carlos", result.Name)
}

func TestCreate_MissingName(t *testing.T) {
	repo := &mockRepo{}
	svc := homeless.NewService(repo, nil)

	input := validInput()
	input.Name = ""

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, homeless.ErrInvalidHomeless)
}

func TestCreate_InvalidGender(t *testing.T) {
	repo := &mockRepo{}
	svc := homeless.NewService(repo, nil)

	input := validInput()
	input.Gender = "invalid"

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, homeless.ErrInvalidHomeless)
}

func TestCreate_NotifierCalled(t *testing.T) {
	repo := &mockRepo{}
	notifier := &mockNotifier{}
	svc := homeless.NewService(repo, notifier)

	_, err := svc.Create(context.Background(), validInput())
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, notifier.CallCount())
}

// --- Tests: FindByID ---

func TestFindByID_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := homeless.NewService(repo, nil)

	created, _ := svc.Create(context.Background(), validInput())

	found, err := svc.FindByID(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestFindByID_EmptyID(t *testing.T) {
	svc := homeless.NewService(&mockRepo{}, nil)

	_, err := svc.FindByID(context.Background(), "")

	assert.ErrorIs(t, err, homeless.ErrInvalidHomeless)
}

func TestFindByID_NotFound(t *testing.T) {
	svc := homeless.NewService(&mockRepo{}, nil)

	_, err := svc.FindByID(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, homeless.ErrHomelessNotFound)
}

// --- Tests: FindAll ---

func TestFindAll_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := homeless.NewService(repo, nil)

	svc.Create(context.Background(), validInput())
	svc.Create(context.Background(), validInput())

	all, err := svc.FindAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, all, 2)
}

// --- Tests: Count ---

func TestCount_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := homeless.NewService(repo, nil)

	svc.Create(context.Background(), validInput())

	count, err := svc.Count(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// --- Tests: Entity ---

func TestEntity_Age(t *testing.T) {
	h := &homeless.Homeless{
		BirthDate: time.Date(1975, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	assert.True(t, h.Age() >= 51)
}

func TestEntity_Validate(t *testing.T) {
	h := &homeless.Homeless{
		Name:   "Test",
		Gender: shared.GenderMale,
		Eyes:   shared.EyeBrown,
		Hair:   shared.HairBlack,
		Skin:   shared.SkinBrown,
	}
	assert.NoError(t, h.Validate())
}
