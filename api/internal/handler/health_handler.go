package handler

import (
	"net/http"
	"time"

	"github.com/l3co/traceo-api/internal/i18n"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type HealthHandler struct {
	startTime time.Time
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// @Summary      Health check
// @Description  Verifica se a API está operacional / Checks if the API is operational
// @Tags         system
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /api/v1/health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, HealthResponse{
		Status:  i18n.T(r.Context(), "StatusOk"),
		Message: i18n.T(r.Context(), "HealthOk"),
		Uptime:  time.Since(h.startTime).String(),
	})
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Uptime  string `json:"uptime"`
}
