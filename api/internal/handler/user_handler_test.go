package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l3co/traceo-api/internal/domain/user"
	"github.com/l3co/traceo-api/internal/handler"
	"github.com/l3co/traceo-api/internal/handler/middleware"
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
	users   map[string]string
	nextUID string
}

func newMockAuth() *mockAuth {
	return &mockAuth{
		users:   make(map[string]string),
		nextUID: "uid-123",
	}
}

func (m *mockAuth) CreateUser(_ context.Context, email, _ string) (string, error) {
	uid := m.nextUID
	m.users[uid] = email
	return uid, nil
}

func (m *mockAuth) VerifyToken(_ context.Context, token string) (string, error) {
	if token == "valid-token" {
		return "uid-123", nil
	}
	return "", user.ErrInvalidPassword
}

func (m *mockAuth) DeleteUser(_ context.Context, uid string) error {
	delete(m.users, uid)
	return nil
}

func (m *mockAuth) ChangePassword(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockAuth) SendPasswordResetEmail(_ context.Context, _ string) error {
	return nil
}

// --- Helpers ---

func setupUserHandler() (*handler.UserHandler, *mockRepo, *mockAuth) {
	repo := newMockRepo()
	auth := newMockAuth()
	svc := user.NewService(repo, auth)
	h := handler.NewUserHandler(svc)
	return h, repo, auth
}

func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withAuthContext(r *http.Request, uid string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, uid)
	return r.WithContext(ctx)
}

func jsonBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

// --- Tests: Create ---

func TestUserHandler_Create_Success(t *testing.T) {
	h, _, _ := setupUserHandler()

	body := jsonBody(t, map[string]interface{}{
		"name":           "João Silva",
		"email":          "joao@email.com",
		"password":       "secret123",
		"accepted_terms": true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "João Silva", resp["name"])
	assert.Equal(t, "joao@email.com", resp["email"])
	assert.Equal(t, "uid-123", resp["id"])
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	h, _, _ := setupUserHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Create_MissingFields(t *testing.T) {
	h, _, _ := setupUserHandler()

	body := jsonBody(t, map[string]interface{}{
		"email": "joao@email.com",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Create_DuplicateEmail(t *testing.T) {
	h, repo, _ := setupUserHandler()
	repo.users["existing"] = &user.User{ID: "existing", Email: "joao@email.com"}

	body := jsonBody(t, map[string]interface{}{
		"name":           "João",
		"email":          "joao@email.com",
		"password":       "secret123",
		"accepted_terms": true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

// --- Tests: FindByID ---

func TestUserHandler_FindByID_Success(t *testing.T) {
	h, repo, _ := setupUserHandler()
	repo.users["uid-123"] = &user.User{ID: "uid-123", Name: "João", Email: "joao@email.com"}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/uid-123", nil)
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "uid-123")
	rec := httptest.NewRecorder()

	h.FindByID(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "João", resp["name"])
}

func TestUserHandler_FindByID_NotFound(t *testing.T) {
	h, _, _ := setupUserHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/nonexistent", nil)
	req = withChiURLParam(req, "id", "nonexistent")
	req = withAuthContext(req, "uid-123")
	rec := httptest.NewRecorder()

	h.FindByID(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// --- Tests: Update ---

func TestUserHandler_Update_Success(t *testing.T) {
	h, repo, _ := setupUserHandler()
	repo.users["uid-123"] = &user.User{ID: "uid-123", Name: "João", Email: "joao@email.com"}

	body := jsonBody(t, map[string]interface{}{
		"name":       "João Silva",
		"phone":      "11999999999",
		"cell_phone": "11888888888",
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/uid-123", body)
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "uid-123")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "João Silva", resp["name"])
}

func TestUserHandler_Update_ForbiddenOtherUser(t *testing.T) {
	h, repo, _ := setupUserHandler()
	repo.users["uid-123"] = &user.User{ID: "uid-123", Name: "João"}

	body := jsonBody(t, map[string]interface{}{"name": "Hack"})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/uid-123", body)
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "other-uid")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// --- Tests: Delete ---

func TestUserHandler_Delete_Success(t *testing.T) {
	h, repo, _ := setupUserHandler()
	repo.users["uid-123"] = &user.User{ID: "uid-123", Name: "João"}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/uid-123", nil)
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "uid-123")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, repo.users)
}

func TestUserHandler_Delete_ForbiddenOtherUser(t *testing.T) {
	h, _, _ := setupUserHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/uid-123", nil)
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "other-uid")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// --- Tests: ChangePassword ---

func TestUserHandler_ChangePassword_Success(t *testing.T) {
	h, _, _ := setupUserHandler()

	body := jsonBody(t, map[string]interface{}{
		"new_password": "newpass123",
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/uid-123/password", body)
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "uid-123")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUserHandler_ChangePassword_ForbiddenOtherUser(t *testing.T) {
	h, _, _ := setupUserHandler()

	body := jsonBody(t, map[string]interface{}{
		"new_password": "newpass123",
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/uid-123/password", body)
	req.Header.Set("Content-Type", "application/json")
	req = withChiURLParam(req, "id", "uid-123")
	req = withAuthContext(req, "other-uid")
	rec := httptest.NewRecorder()

	h.ChangePassword(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
