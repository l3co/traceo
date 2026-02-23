package handler

import (
	"net/http"

	"github.com/l3co/traceo-api/internal/domain/user"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type AuthHandler struct {
	service *user.Service
}

func NewAuthHandler(service *user.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// @Summary      Recuperar senha
// @Description  Envia email de recuperação de senha
// @Tags         auth
// @Accept       json
// @Param        body  body  ForgotPasswordRequest  true  "Email do usuário"
// @Success      204   "Sem conteúdo"
// @Failure      400   {object}  httputil.ErrorResponse  "Email inválido"
// @Router       /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to send password reset email")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
