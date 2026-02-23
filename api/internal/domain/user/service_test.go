package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l3co/traceo-api/internal/domain/user"
)

// --- Mocks ---

type mockRepo struct {
	users map[string]*user.User
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*user.User)}
}

func (m *mockRepo) Create(_ context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (m *mockRepo) Update(_ context.Context, u *user.User) error {
	if _, ok := m.users[u.ID]; !ok {
		return user.ErrUserNotFound
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return user.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

type mockAuth struct {
	users    map[string]string // uid -> email
	nextUID  string
	deleted  []string
	resetted []string
}

func newMockAuth() *mockAuth {
	return &mockAuth{
		users:   make(map[string]string),
		nextUID: "firebase-uid-123",
	}
}

func (m *mockAuth) CreateUser(_ context.Context, email, _ string) (string, error) {
	uid := m.nextUID
	m.users[uid] = email
	return uid, nil
}

func (m *mockAuth) VerifyToken(_ context.Context, token string) (string, error) {
	if token == "" {
		return "", user.ErrInvalidPassword
	}
	return "verified-uid", nil
}

func (m *mockAuth) DeleteUser(_ context.Context, uid string) error {
	delete(m.users, uid)
	m.deleted = append(m.deleted, uid)
	return nil
}

func (m *mockAuth) ChangePassword(_ context.Context, _ string, newPassword string) error {
	if newPassword == "" {
		return user.ErrInvalidPassword
	}
	return nil
}

func (m *mockAuth) SendPasswordResetEmail(_ context.Context, _ string) error {
	return nil
}

// --- Tests: Create ---

func TestCreate_Success(t *testing.T) {
	repo := newMockRepo()
	auth := newMockAuth()
	svc := user.NewService(repo, auth)

	created, err := svc.Create(context.Background(), &user.CreateInput{
		Name:          "João Silva",
		Email:         "joao@email.com",
		Password:      "secret123",
		AcceptedTerms: true,
	})

	require.NoError(t, err)
	assert.Equal(t, "João Silva", created.Name)
	assert.Equal(t, "joao@email.com", created.Email)
	assert.Equal(t, "firebase-uid-123", created.ID)
	assert.True(t, created.AcceptedTerms)
	assert.NotZero(t, created.CreatedAt)
}

func TestCreate_SanitizesInput(t *testing.T) {
	repo := newMockRepo()
	auth := newMockAuth()
	svc := user.NewService(repo, auth)

	created, err := svc.Create(context.Background(), &user.CreateInput{
		Name:          "  Maria  ",
		Email:         "  MARIA@Email.COM  ",
		Password:      "secret",
		AcceptedTerms: true,
	})

	require.NoError(t, err)
	assert.Equal(t, "Maria", created.Name)
	assert.Equal(t, "maria@email.com", created.Email)
}

func TestCreate_MissingRequiredFields(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	tests := []struct {
		name  string
		input user.CreateInput
	}{
		{"missing name", user.CreateInput{Email: "a@b.c", Password: "123456", AcceptedTerms: true}},
		{"missing email", user.CreateInput{Name: "João", Password: "123456", AcceptedTerms: true}},
		{"missing password", user.CreateInput{Name: "João", Email: "a@b.c", AcceptedTerms: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), &tt.input)
			assert.ErrorIs(t, err, user.ErrInvalidInput)
		})
	}
}

func TestCreate_TermsNotAccepted(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	_, err := svc.Create(context.Background(), &user.CreateInput{
		Name:          "João",
		Email:         "joao@email.com",
		Password:      "secret",
		AcceptedTerms: false,
	})

	assert.ErrorIs(t, err, user.ErrTermsNotAccepted)
}

func TestCreate_DuplicateEmail(t *testing.T) {
	repo := newMockRepo()
	repo.users["existing"] = &user.User{ID: "existing", Email: "joao@email.com"}
	svc := user.NewService(repo, newMockAuth())

	_, err := svc.Create(context.Background(), &user.CreateInput{
		Name:          "Outro João",
		Email:         "joao@email.com",
		Password:      "secret",
		AcceptedTerms: true,
	})

	assert.ErrorIs(t, err, user.ErrEmailAlreadyExists)
}

// --- Tests: FindByID ---

func TestFindByID_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["uid-1"] = &user.User{ID: "uid-1", Name: "João"}
	svc := user.NewService(repo, newMockAuth())

	found, err := svc.FindByID(context.Background(), "uid-1")

	require.NoError(t, err)
	assert.Equal(t, "João", found.Name)
}

func TestFindByID_NotFound(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	_, err := svc.FindByID(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestFindByID_EmptyID(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	_, err := svc.FindByID(context.Background(), "")

	assert.ErrorIs(t, err, user.ErrInvalidInput)
}

// --- Tests: Update ---

func TestUpdate_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["uid-1"] = &user.User{ID: "uid-1", Name: "João"}
	svc := user.NewService(repo, newMockAuth())

	updated, err := svc.Update(context.Background(), "uid-1", &user.UpdateInput{
		Name:      "João Silva",
		Phone:     "11999999999",
		CellPhone: "11888888888",
	})

	require.NoError(t, err)
	assert.Equal(t, "João Silva", updated.Name)
	assert.Equal(t, "11999999999", updated.Phone)
	assert.NotZero(t, updated.UpdatedAt)
}

func TestUpdate_NotFound(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	_, err := svc.Update(context.Background(), "nonexistent", &user.UpdateInput{Name: "João"})

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestUpdate_EmptyName(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	_, err := svc.Update(context.Background(), "uid-1", &user.UpdateInput{Name: ""})

	assert.ErrorIs(t, err, user.ErrInvalidInput)
}

// --- Tests: Delete ---

func TestDelete_Success(t *testing.T) {
	repo := newMockRepo()
	auth := newMockAuth()
	repo.users["uid-1"] = &user.User{ID: "uid-1", Name: "João"}
	svc := user.NewService(repo, auth)

	err := svc.Delete(context.Background(), "uid-1")

	require.NoError(t, err)
	assert.Empty(t, repo.users)
	assert.Contains(t, auth.deleted, "uid-1")
}

func TestDelete_NotFound(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	err := svc.Delete(context.Background(), "nonexistent")

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

// --- Tests: ChangePassword ---

func TestChangePassword_Success(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	err := svc.ChangePassword(context.Background(), "uid-1", "newpass123")

	assert.NoError(t, err)
}

func TestChangePassword_EmptyFields(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	assert.ErrorIs(t, svc.ChangePassword(context.Background(), "", "newpass"), user.ErrInvalidInput)
	assert.ErrorIs(t, svc.ChangePassword(context.Background(), "uid", ""), user.ErrInvalidInput)
}

// --- Tests: ForgotPassword ---

func TestForgotPassword_Success(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	err := svc.ForgotPassword(context.Background(), "joao@email.com")

	assert.NoError(t, err)
}

func TestForgotPassword_EmptyEmail(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockAuth())

	err := svc.ForgotPassword(context.Background(), "")

	assert.ErrorIs(t, err, user.ErrInvalidInput)
}
