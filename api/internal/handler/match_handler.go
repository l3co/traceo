package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/l3co/traceo-api/internal/domain/matching"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type MatchHandler struct {
	service *matching.Service
}

func NewMatchHandler(service *matching.Service) *MatchHandler {
	return &MatchHandler{service: service}
}

// --- DTOs ---

type MatchResponse struct {
	ID             string  `json:"id"`
	HomelessID     string  `json:"homeless_id"`
	MissingID      string  `json:"missing_id"`
	Score          float64 `json:"score"`
	Status         string  `json:"status"`
	GeminiAnalysis string  `json:"gemini_analysis,omitempty"`
	CreatedAt      string  `json:"created_at"`
	ReviewedAt     string  `json:"reviewed_at,omitempty"`
}

type UpdateMatchStatusRequest struct {
	Status string `json:"status"`
}

func toMatchResponse(m *matching.Match) MatchResponse {
	resp := MatchResponse{
		ID:             m.ID,
		HomelessID:     m.HomelessID,
		MissingID:      m.MissingID,
		Score:          m.Score,
		Status:         string(m.Status),
		GeminiAnalysis: m.GeminiAnalysis,
		CreatedAt:      m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if !m.ReviewedAt.IsZero() {
		resp.ReviewedAt = m.ReviewedAt.Format("2006-01-02T15:04:05Z")
	}
	return resp
}

// @Summary      Listar matches de um homeless
// @Description  Retorna candidatos de match para um homeless
// @Tags         matches
// @Produce      json
// @Param        id   path      string  true  "Homeless ID"
// @Success      200  {array}   MatchResponse
// @Router       /api/v1/homeless/{id}/matches [get]
func (h *MatchHandler) FindByHomelessID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	items, err := h.service.FindByHomelessID(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list matches")
		return
	}

	resp := make([]MatchResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toMatchResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Listar matches de um desaparecido
// @Description  Retorna matches encontrados para um missing
// @Tags         matches
// @Produce      json
// @Param        id   path      string  true  "Missing ID"
// @Success      200  {array}   MatchResponse
// @Router       /api/v1/missing/{id}/matches [get]
func (h *MatchHandler) FindByMissingID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	items, err := h.service.FindByMissingID(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list matches")
		return
	}

	resp := make([]MatchResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toMatchResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Atualizar status do match
// @Description  Confirmar ou rejeitar match
// @Tags         matches
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Match ID"
// @Param        body  body      UpdateMatchStatusRequest  true  "Status"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  httputil.ErrorResponse
// @Failure      404   {object}  httputil.ErrorResponse
// @Router       /api/v1/matches/{id} [patch]
func (h *MatchHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateMatchStatusRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.service.UpdateStatus(r.Context(), id, matching.MatchStatus(req.Status))
	if err != nil {
		if errors.Is(err, matching.ErrInvalidMatch) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, matching.ErrMatchNotFound) {
			httputil.Error(w, http.StatusNotFound, "match not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update match")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
