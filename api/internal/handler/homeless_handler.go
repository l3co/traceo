package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/shared"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type HomelessHandler struct {
	service *homeless.Service
}

func NewHomelessHandler(service *homeless.Service) *HomelessHandler {
	return &HomelessHandler{service: service}
}

// --- DTOs ---

type CreateHomelessRequest struct {
	Name      string  `json:"name"`
	Nickname  string  `json:"nickname"`
	BirthDate string  `json:"birth_date"`
	Gender    string  `json:"gender"`
	Eyes      string  `json:"eyes"`
	Hair      string  `json:"hair"`
	Skin      string  `json:"skin"`
	PhotoURL  string  `json:"photo_url"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

type HomelessResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Nickname  string  `json:"nickname,omitempty"`
	BirthDate string  `json:"birth_date,omitempty"`
	Age       int     `json:"age"`
	Gender    string  `json:"gender"`
	Eyes      string  `json:"eyes"`
	Hair      string  `json:"hair"`
	Skin      string  `json:"skin"`
	PhotoURL  string  `json:"photo_url,omitempty"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Slug      string  `json:"slug"`
	CreatedAt string  `json:"created_at"`
}

type HomelessStatsResponse struct {
	Total    int64              `json:"total"`
	ByGender []GenderStatDTO    `json:"by_gender"`
}

func toHomelessResponse(h *homeless.Homeless) HomelessResponse {
	var birthStr string
	if !h.BirthDate.IsZero() {
		birthStr = h.BirthDate.Format("02/01/2006")
	}
	return HomelessResponse{
		ID:        h.ID,
		Name:      h.Name,
		Nickname:  h.Nickname,
		BirthDate: birthStr,
		Age:       h.Age(),
		Gender:    string(h.Gender),
		Eyes:      string(h.Eyes),
		Hair:      string(h.Hair),
		Skin:      string(h.Skin),
		PhotoURL:  h.PhotoURL,
		Lat:       h.Location.Lat,
		Lng:       h.Location.Lng,
		Slug:      h.Slug,
		CreatedAt: h.CreatedAt.Format(time.RFC3339),
	}
}

// @Summary      Cadastrar morador de rua
// @Description  Registra uma pessoa em situação de rua que quer ser encontrada
// @Tags         homeless
// @Accept       json
// @Produce      json
// @Param        body  body      CreateHomelessRequest  true  "Dados"
// @Success      201   {object}  HomelessResponse
// @Failure      400   {object}  httputil.ErrorResponse
// @Router       /api/v1/homeless [post]
func (h *HomelessHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateHomelessRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var birthDate time.Time
	if req.BirthDate != "" {
		parsed, err := time.Parse("02/01/2006", req.BirthDate)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid birth_date format, use DD/MM/YYYY")
			return
		}
		birthDate = parsed
	}

	input := homeless.CreateInput{
		Name:      req.Name,
		Nickname:  req.Nickname,
		BirthDate: birthDate,
		Gender:    shared.Gender(req.Gender),
		Eyes:      shared.EyeColor(req.Eyes),
		Hair:      shared.HairColor(req.Hair),
		Skin:      shared.SkinColor(req.Skin),
		PhotoURL:  req.PhotoURL,
		Lat:       req.Lat,
		Lng:       req.Lng,
	}

	result, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, homeless.ErrInvalidHomeless) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to create homeless")
		return
	}

	httputil.JSON(w, http.StatusCreated, toHomelessResponse(result))
}

// @Summary      Listar moradores de rua
// @Description  Retorna todos os registros de moradores de rua
// @Tags         homeless
// @Produce      json
// @Success      200  {array}  HomelessResponse
// @Router       /api/v1/homeless [get]
func (h *HomelessHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.FindAll(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list homeless")
		return
	}

	resp := make([]HomelessResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toHomelessResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Buscar morador de rua por ID
// @Description  Retorna um registro específico
// @Tags         homeless
// @Produce      json
// @Param        id   path      string  true  "ID"
// @Success      200  {object}  HomelessResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /api/v1/homeless/{id} [get]
func (h *HomelessHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, homeless.ErrHomelessNotFound) {
			httputil.Error(w, http.StatusNotFound, "homeless not found")
			return
		}
		if errors.Is(err, homeless.ErrInvalidHomeless) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to find homeless")
		return
	}

	httputil.JSON(w, http.StatusOK, toHomelessResponse(result))
}

// @Summary      Estatísticas de moradores de rua
// @Description  Retorna total e distribuição por gênero
// @Tags         homeless
// @Produce      json
// @Success      200  {object}  HomelessStatsResponse
// @Router       /api/v1/homeless/stats [get]
func (h *HomelessHandler) Stats(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.Count(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	byGender, err := h.service.CountByGender(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	genderDTOs := make([]GenderStatDTO, 0, len(byGender))
	for _, g := range byGender {
		genderDTOs = append(genderDTOs, GenderStatDTO{Gender: string(g.Gender), Count: g.Count})
	}

	httputil.JSON(w, http.StatusOK, HomelessStatsResponse{
		Total:    count,
		ByGender: genderDTOs,
	})
}
