package missing_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l3co/traceo-api/internal/domain/missing"
)

// --- Mock Repository ---

type mockRepo struct {
	items map[string]*missing.Missing
}

func newMockRepo() *mockRepo {
	return &mockRepo{items: make(map[string]*missing.Missing)}
}

func (m *mockRepo) Create(_ context.Context, item *missing.Missing) error {
	m.items[item.ID] = item
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*missing.Missing, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, missing.ErrMissingNotFound
	}
	return item, nil
}

func (m *mockRepo) Update(_ context.Context, item *missing.Missing) error {
	if _, ok := m.items[item.ID]; !ok {
		return missing.ErrMissingNotFound
	}
	m.items[item.ID] = item
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return missing.ErrMissingNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *mockRepo) FindByUserID(_ context.Context, userID string) ([]*missing.Missing, error) {
	var result []*missing.Missing
	for _, item := range m.items {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockRepo) FindAll(_ context.Context, opts missing.ListOptions) ([]*missing.Missing, string, error) {
	var result []*missing.Missing
	for _, item := range m.items {
		result = append(result, item)
	}
	if len(result) > opts.PageSize {
		result = result[:opts.PageSize]
	}
	cursor := ""
	if len(result) == opts.PageSize {
		cursor = result[len(result)-1].ID
	}
	return result, cursor, nil
}

func (m *mockRepo) Count(_ context.Context) (int64, error) {
	return int64(len(m.items)), nil
}

func (m *mockRepo) Search(_ context.Context, query string, limit int) ([]*missing.Missing, error) {
	var result []*missing.Missing
	q := strings.ToLower(query)
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.Name), q) {
			result = append(result, item)
		}
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockRepo) CountByGender(_ context.Context) ([]missing.GenderStat, error) {
	counts := map[string]int64{}
	for _, item := range m.items {
		counts[string(item.Gender)]++
	}
	var result []missing.GenderStat
	for g, c := range counts {
		result = append(result, missing.GenderStat{Gender: g, Count: c})
	}
	return result, nil
}

func (m *mockRepo) CountByYear(_ context.Context) ([]missing.YearStat, error) {
	counts := map[int]int64{}
	for _, item := range m.items {
		if !item.DateOfDisappearance.IsZero() {
			counts[item.DateOfDisappearance.Year()]++
		}
	}
	var result []missing.YearStat
	for y, c := range counts {
		result = append(result, missing.YearStat{Year: y, Count: c})
	}
	return result, nil
}

func (m *mockRepo) CountChildren(_ context.Context) (int64, error) {
	var count int64
	for _, item := range m.items {
		if item.WasChild {
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) FindLocations(_ context.Context, limit int) ([]missing.LocationPoint, error) {
	var result []missing.LocationPoint
	for _, item := range m.items {
		if item.Location.Lat != 0 || item.Location.Lng != 0 {
			result = append(result, missing.LocationPoint{
				ID:     item.ID,
				Name:   item.Name,
				Lat:    item.Location.Lat,
				Lng:    item.Location.Lng,
				Status: item.Status,
			})
		}
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockRepo) FindCandidates(_ context.Context, _ missing.CandidateFilter) ([]*missing.Missing, error) {
	var result []*missing.Missing
	for _, item := range m.items {
		result = append(result, item)
	}
	return result, nil
}

func (m *mockRepo) UpdateAgeProgressionURLs(_ context.Context, id string, urls []string) error {
	item, ok := m.items[id]
	if !ok {
		return missing.ErrMissingNotFound
	}
	_ = item
	_ = urls
	return nil
}

// --- Helpers ---

func validInput() *missing.CreateInput {
	return &missing.CreateInput{
		UserID:              "user-123",
		Name:                "João Silva",
		Nickname:            "Joãozinho",
		BirthDate:           time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		DateOfDisappearance: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		Height:              "175cm",
		Clothes:             "Camiseta azul, calça jeans",
		Gender:              missing.GenderMale,
		Eyes:                missing.EyeBrown,
		Hair:                missing.HairBlack,
		Skin:                missing.SkinBrown,
		Location:            missing.GeoPoint{Lat: -23.5505, Lng: -46.6333},
		EventReport:         "BO 123456",
		TattooDescription:   "Dragão no braço esquerdo",
		ScarDescription:     "",
	}
}

// --- Tests: Create ---

func TestCreate_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	result, err := svc.Create(context.Background(), validInput())

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "João Silva", result.Name)
	assert.Equal(t, "user-123", result.UserID)
	assert.Equal(t, missing.StatusDisappeared, result.Status)
	assert.NotEmpty(t, result.Slug)
	assert.False(t, result.WasChild)
	assert.True(t, result.HasTattoo())
	assert.False(t, result.HasScar())
	assert.NotZero(t, result.CreatedAt)
}

func TestCreate_WasChild(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	input := validInput()
	input.BirthDate = time.Date(2010, 6, 1, 0, 0, 0, 0, time.UTC)
	input.DateOfDisappearance = time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	result, err := svc.Create(context.Background(), input)

	require.NoError(t, err)
	assert.True(t, result.WasChild)
}

func TestCreate_SanitizesInput(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	input := validInput()
	input.Name = "<script>alert('xss')</script>João"

	result, err := svc.Create(context.Background(), input)

	require.NoError(t, err)
	assert.NotContains(t, result.Name, "<script>")
}

func TestCreate_MissingName(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	input := validInput()
	input.Name = ""

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

func TestCreate_MissingUserID(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	input := validInput()
	input.UserID = ""

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

func TestCreate_InvalidGender(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	input := validInput()
	input.Gender = "banana"

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

func TestCreate_FutureDateOfDisappearance(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	input := validInput()
	input.DateOfDisappearance = time.Now().Add(24 * time.Hour)

	_, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

// --- Tests: FindByID ---

func TestFindByID_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	created, _ := svc.Create(context.Background(), validInput())

	found, err := svc.FindByID(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestFindByID_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	_, err := svc.FindByID(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, missing.ErrMissingNotFound)
}

func TestFindByID_EmptyID(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	_, err := svc.FindByID(context.Background(), "")

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

// --- Tests: Update ---

func TestUpdate_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	created, _ := svc.Create(context.Background(), validInput())

	updated, err := svc.Update(context.Background(), created.ID, "user-123", &missing.UpdateInput{
		Name:   "João Silva Atualizado",
		Gender: missing.GenderMale,
		Eyes:   missing.EyeBrown,
		Hair:   missing.HairBlack,
		Skin:   missing.SkinBrown,
	})

	require.NoError(t, err)
	assert.Equal(t, "João Silva Atualizado", updated.Name)
}

func TestUpdate_NotOwner(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	created, _ := svc.Create(context.Background(), validInput())

	_, err := svc.Update(context.Background(), created.ID, "other-user", &missing.UpdateInput{
		Name:   "Hack",
		Gender: missing.GenderMale,
		Eyes:   missing.EyeBrown,
		Hair:   missing.HairBlack,
		Skin:   missing.SkinBrown,
	})

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	_, err := svc.Update(context.Background(), "nonexistent", "user-123", &missing.UpdateInput{
		Name:   "Test",
		Gender: missing.GenderMale,
		Eyes:   missing.EyeBrown,
		Hair:   missing.HairBlack,
		Skin:   missing.SkinBrown,
	})

	assert.ErrorIs(t, err, missing.ErrMissingNotFound)
}

// --- Tests: Delete ---

func TestDelete_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	created, _ := svc.Create(context.Background(), validInput())

	err := svc.Delete(context.Background(), created.ID, "user-123")

	assert.NoError(t, err)
	assert.Empty(t, repo.items)
}

func TestDelete_NotOwner(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	created, _ := svc.Create(context.Background(), validInput())

	err := svc.Delete(context.Background(), created.ID, "other-user")

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

// --- Tests: FindByUserID ---

func TestFindByUserID_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())

	input2 := validInput()
	input2.Name = "Maria"
	svc.Create(context.Background(), input2)

	results, err := svc.FindByUserID(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFindByUserID_EmptyID(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	_, err := svc.FindByUserID(context.Background(), "")

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

// --- Tests: List ---

func TestList_DefaultPageSize(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())

	results, _, err := svc.List(context.Background(), missing.ListOptions{PageSize: 0})

	require.NoError(t, err)
	assert.Len(t, results, 1)
}

// --- Tests: Count ---

func TestCount_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())
	svc.Create(context.Background(), validInput())

	count, err := svc.Count(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// --- Tests: Value Objects ---

func TestGender_IsValid(t *testing.T) {
	assert.True(t, missing.GenderMale.IsValid())
	assert.True(t, missing.GenderFemale.IsValid())
	assert.False(t, missing.Gender("banana").IsValid())
}

func TestGender_Label(t *testing.T) {
	assert.Equal(t, "Masculino", missing.GenderMale.Label())
	assert.Equal(t, "Feminino", missing.GenderFemale.Label())
}

func TestEyeColor_IsValid(t *testing.T) {
	assert.True(t, missing.EyeGreen.IsValid())
	assert.True(t, missing.EyeBlue.IsValid())
	assert.False(t, missing.EyeColor("red").IsValid())
}

func TestHairColor_IsValid(t *testing.T) {
	assert.True(t, missing.HairBlack.IsValid())
	assert.False(t, missing.HairColor("purple").IsValid())
}

func TestSkinColor_IsValid(t *testing.T) {
	assert.True(t, missing.SkinBrown.IsValid())
	assert.False(t, missing.SkinColor("green").IsValid())
}

func TestStatus_IsValid(t *testing.T) {
	assert.True(t, missing.StatusDisappeared.IsValid())
	assert.True(t, missing.StatusFound.IsValid())
	assert.False(t, missing.Status("unknown").IsValid())
}

// --- Tests: Entity behavior ---

// --- Tests: Search ---

func TestSearch_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())

	results, err := svc.Search(context.Background(), "João", 20)

	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestSearch_EmptyQuery(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	_, err := svc.Search(context.Background(), "", 20)

	assert.ErrorIs(t, err, missing.ErrInvalidMissing)
}

func TestSearch_NoResults(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())

	results, err := svc.Search(context.Background(), "Maria", 20)

	require.NoError(t, err)
	assert.Empty(t, results)
}

// --- Tests: GetStats (goroutines) ---

func TestGetStats_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())

	input2 := validInput()
	input2.Name = "Maria Souza"
	input2.Gender = missing.GenderFemale
	svc.Create(context.Background(), input2)

	stats, err := svc.GetStats(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(2), stats.Total)
	assert.NotEmpty(t, stats.ByGender)
	assert.NotEmpty(t, stats.ByYear)
}

func TestGetStats_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	stats, err := svc.GetStats(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.Total)
	assert.Equal(t, int64(0), stats.ChildCount)
}

// --- Tests: FindLocations ---

func TestFindLocations_Success(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	svc.Create(context.Background(), validInput())

	locs, err := svc.FindLocations(context.Background(), 100)

	require.NoError(t, err)
	assert.Len(t, locs, 1)
	assert.Equal(t, -23.5505, locs[0].Lat)
}

func TestFindLocations_DefaultLimit(t *testing.T) {
	repo := newMockRepo()
	svc := missing.NewService(repo)

	locs, err := svc.FindLocations(context.Background(), 0)

	require.NoError(t, err)
	assert.Empty(t, locs)
}

// --- Tests: Entity ---

func TestEntity_Age(t *testing.T) {
	m := &missing.Missing{
		BirthDate: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	age := m.Age()
	assert.True(t, age >= 35)
}

func TestEntity_HasTattoo(t *testing.T) {
	m := &missing.Missing{TattooDescription: "Dragon on left arm"}
	assert.True(t, m.HasTattoo())

	m2 := &missing.Missing{}
	assert.False(t, m2.HasTattoo())
}

func TestEntity_HasScar(t *testing.T) {
	m := &missing.Missing{ScarDescription: "Scar on forehead"}
	assert.True(t, m.HasScar())

	m2 := &missing.Missing{}
	assert.False(t, m2.HasScar())
}
