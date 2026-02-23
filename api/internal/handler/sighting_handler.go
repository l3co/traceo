package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/l3co/traceo-api/internal/domain/sighting"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type SightingHandler struct {
	service *sighting.Service
}

func NewSightingHandler(service *sighting.Service) *SightingHandler {
	return &SightingHandler{service: service}
}

// --- DTOs ---

type CreateSightingRequest struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Observation string  `json:"observation"`
}

type SightingResponse struct {
	ID          string  `json:"id"`
	MissingID   string  `json:"missing_id"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Observation string  `json:"observation"`
	CreatedAt   string  `json:"created_at"`
}

func toSightingResponse(s *sighting.Sighting) SightingResponse {
	return SightingResponse{
		ID:          s.ID,
		MissingID:   s.MissingID,
		Lat:         s.Location.Lat,
		Lng:         s.Location.Lng,
		Observation: s.Observation,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary      Registrar avistamento
// @Description  Registra um avistamento de pessoa desaparecida
// @Tags         sightings
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "ID do desaparecido"
// @Param        body  body      CreateSightingRequest   true  "Dados do avistamento"
// @Success      201   {object}  SightingResponse
// @Failure      400   {object}  httputil.ErrorResponse
// @Router       /api/v1/missing/{id}/sightings [post]
func (h *SightingHandler) Create(w http.ResponseWriter, r *http.Request) {
	missingID := chi.URLParam(r, "id")

	var req CreateSightingRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := sighting.CreateInput{
		MissingID:   missingID,
		Lat:         req.Lat,
		Lng:         req.Lng,
		Observation: req.Observation,
	}

	result, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, sighting.ErrInvalidSighting) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to create sighting")
		return
	}

	httputil.JSON(w, http.StatusCreated, toSightingResponse(result))
}

// @Summary      Listar avistamentos de um desaparecido
// @Description  Retorna todos os avistamentos de uma pessoa desaparecida
// @Tags         sightings
// @Produce      json
// @Param        id   path      string  true  "ID do desaparecido"
// @Success      200  {array}   SightingResponse
// @Router       /api/v1/missing/{id}/sightings [get]
func (h *SightingHandler) FindByMissingID(w http.ResponseWriter, r *http.Request) {
	missingID := chi.URLParam(r, "id")

	items, err := h.service.FindByMissingID(r.Context(), missingID)
	if err != nil {
		if errors.Is(err, sighting.ErrInvalidSighting) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to list sightings")
		return
	}

	resp := make([]SightingResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toSightingResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Buscar avistamento por ID
// @Description  Retorna um avistamento específico
// @Tags         sightings
// @Produce      json
// @Param        sightingId  path      string  true  "ID do avistamento"
// @Success      200         {object}  SightingResponse
// @Failure      404         {object}  httputil.ErrorResponse
// @Router       /api/v1/sightings/{sightingId} [get]
func (h *SightingHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "sightingId")

	result, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sighting.ErrSightingNotFound) {
			httputil.Error(w, http.StatusNotFound, "sighting not found")
			return
		}
		if errors.Is(err, sighting.ErrInvalidSighting) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to find sighting")
		return
	}

	httputil.JSON(w, http.StatusOK, toSightingResponse(result))
}
