package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/l3co/traceo-api/internal/domain/user"
	"github.com/l3co/traceo-api/internal/handler/middleware"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type UserHandler struct {
	service *user.Service
}

func NewUserHandler(service *user.Service) *UserHandler {
	return &UserHandler{service: service}
}

// @Summary      Criar usuário
// @Description  Registra um novo usuário na plataforma
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserRequest  true  "Dados do usuário"
// @Success      201   {object}  UserResponse
// @Failure      400   {object}  httputil.ErrorResponse  "Dados inválidos"
// @Failure      409   {object}  httputil.ErrorResponse  "Email já existe"
// @Router       /api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.service.Create(r.Context(), &user.CreateInput{
		Name:          httputil.SanitizeString(req.Name),
		Email:         req.Email,
		Password:      req.Password,
		Phone:         httputil.SanitizeString(req.Phone),
		CellPhone:     httputil.SanitizeString(req.CellPhone),
		AcceptedTerms: req.AcceptedTerms,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailAlreadyExists):
			httputil.Error(w, http.StatusConflict, "email already exists")
		case errors.Is(err, user.ErrTermsNotAccepted):
			httputil.Error(w, http.StatusBadRequest, "terms must be accepted")
		case errors.Is(err, user.ErrInvalidInput):
			httputil.Error(w, http.StatusBadRequest, err.Error())
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	resp := toUserResponse(created.ID, created.Name, created.Email, created.Phone, created.CellPhone, created.AvatarURL, created.CreatedAt, created.UpdatedAt)
	httputil.JSON(w, http.StatusCreated, resp)
}

// @Summary      Buscar usuário por ID
// @Description  Retorna os dados de um usuário pelo seu ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "ID do usuário"
// @Success      200  {object}  UserResponse
// @Failure      404  {object}  httputil.ErrorResponse  "Usuário não encontrado"
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	found, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to find user")
		return
	}

	resp := toUserResponse(found.ID, found.Name, found.Email, found.Phone, found.CellPhone, found.AvatarURL, found.CreatedAt, found.UpdatedAt)
	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Atualizar usuário
// @Description  Atualiza os dados do perfil do usuário autenticado
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "ID do usuário"
// @Param        body  body      UpdateUserRequest  true  "Dados atualizados"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  httputil.ErrorResponse  "Dados inválidos"
// @Failure      403   {object}  httputil.ErrorResponse  "Sem permissão"
// @Failure      404   {object}  httputil.ErrorResponse  "Usuário não encontrado"
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	authUID := middleware.GetUserID(r.Context())

	if id != authUID {
		httputil.Error(w, http.StatusForbidden, "cannot update another user")
		return
	}

	var req UpdateUserRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.Update(r.Context(), id, &user.UpdateInput{
		Name:      httputil.SanitizeString(req.Name),
		Phone:     httputil.SanitizeString(req.Phone),
		CellPhone: httputil.SanitizeString(req.CellPhone),
	})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	resp := toUserResponse(updated.ID, updated.Name, updated.Email, updated.Phone, updated.CellPhone, updated.AvatarURL, updated.CreatedAt, updated.UpdatedAt)
	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Deletar usuário
// @Description  Remove o usuário autenticado da plataforma
// @Tags         users
// @Param        id   path  string  true  "ID do usuário"
// @Success      204  "Sem conteúdo"
// @Failure      403  {object}  httputil.ErrorResponse  "Sem permissão"
// @Failure      404  {object}  httputil.ErrorResponse  "Usuário não encontrado"
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	authUID := middleware.GetUserID(r.Context())

	if id != authUID {
		httputil.Error(w, http.StatusForbidden, "cannot delete another user")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			httputil.Error(w, http.StatusNotFound, "user not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Alterar senha
// @Description  Altera a senha do usuário autenticado
// @Tags         users
// @Accept       json
// @Param        id    path  string                 true  "ID do usuário"
// @Param        body  body  ChangePasswordRequest  true  "Nova senha"
// @Success      204   "Sem conteúdo"
// @Failure      400   {object}  httputil.ErrorResponse  "Dados inválidos"
// @Failure      403   {object}  httputil.ErrorResponse  "Sem permissão"
// @Security     BearerAuth
// @Router       /api/v1/users/{id}/password [patch]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	authUID := middleware.GetUserID(r.Context())

	if id != authUID {
		httputil.Error(w, http.StatusForbidden, "cannot change another user's password")
		return
	}

	var req ChangePasswordRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ChangePassword(r.Context(), id, req.NewPassword); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
