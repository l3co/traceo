package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/handler/middleware"
	"github.com/l3co/traceo-api/pkg/httputil"
)

type MissingHandler struct {
	service *missing.Service
}

func NewMissingHandler(service *missing.Service) *MissingHandler {
	return &MissingHandler{service: service}
}

// --- Request/Response DTOs ---

type CreateMissingRequest struct {
	Name                string  `json:"name" validate:"required,max=200"`
	Nickname            string  `json:"nickname,omitempty" validate:"omitempty,max=100"`
	BirthDate           string  `json:"birth_date,omitempty"`
	DateOfDisappearance string  `json:"date_of_disappearance,omitempty"`
	Height              string  `json:"height,omitempty" validate:"omitempty,max=20"`
	Clothes             string  `json:"clothes,omitempty" validate:"omitempty,max=500"`
	Gender              string  `json:"gender" validate:"required"`
	Eyes                string  `json:"eyes" validate:"required"`
	Hair                string  `json:"hair" validate:"required"`
	Skin                string  `json:"skin" validate:"required"`
	PhotoURL            string  `json:"photo_url,omitempty"`
	Lat                 float64 `json:"lat"`
	Lng                 float64 `json:"lng"`
	EventReport         string  `json:"event_report,omitempty" validate:"omitempty,max=2000"`
	TattooDescription   string  `json:"tattoo_description,omitempty" validate:"omitempty,max=500"`
	ScarDescription     string  `json:"scar_description,omitempty" validate:"omitempty,max=500"`
}

type UpdateMissingRequest struct {
	Name                string  `json:"name" validate:"required,max=200"`
	Nickname            string  `json:"nickname,omitempty" validate:"omitempty,max=100"`
	BirthDate           string  `json:"birth_date,omitempty"`
	DateOfDisappearance string  `json:"date_of_disappearance,omitempty"`
	Height              string  `json:"height,omitempty" validate:"omitempty,max=20"`
	Clothes             string  `json:"clothes,omitempty" validate:"omitempty,max=500"`
	Gender              string  `json:"gender" validate:"required"`
	Eyes                string  `json:"eyes" validate:"required"`
	Hair                string  `json:"hair" validate:"required"`
	Skin                string  `json:"skin" validate:"required"`
	PhotoURL            string  `json:"photo_url,omitempty"`
	Lat                 float64 `json:"lat"`
	Lng                 float64 `json:"lng"`
	Status              string  `json:"status,omitempty"`
	EventReport         string  `json:"event_report,omitempty" validate:"omitempty,max=2000"`
	TattooDescription   string  `json:"tattoo_description,omitempty" validate:"omitempty,max=500"`
	ScarDescription     string  `json:"scar_description,omitempty" validate:"omitempty,max=500"`
}

type MissingResponse struct {
	ID                  string  `json:"id"`
	UserID              string  `json:"user_id"`
	Name                string  `json:"name"`
	Nickname            string  `json:"nickname,omitempty"`
	BirthDate           string  `json:"birth_date,omitempty"`
	DateOfDisappearance string  `json:"date_of_disappearance,omitempty"`
	Height              string  `json:"height,omitempty"`
	Clothes             string  `json:"clothes,omitempty"`
	Gender              string  `json:"gender"`
	Eyes                string  `json:"eyes"`
	Hair                string  `json:"hair"`
	Skin                string  `json:"skin"`
	PhotoURL            string  `json:"photo_url,omitempty"`
	Lat                 float64 `json:"lat"`
	Lng                 float64 `json:"lng"`
	Status              string  `json:"status"`
	EventReport         string  `json:"event_report,omitempty"`
	TattooDescription   string  `json:"tattoo_description,omitempty"`
	ScarDescription     string  `json:"scar_description,omitempty"`
	WasChild            bool    `json:"was_child"`
	Slug                string  `json:"slug"`
	HasTattoo           bool    `json:"has_tattoo"`
	HasScar             bool    `json:"has_scar"`
	Age                 int     `json:"age"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type MissingListResponse struct {
	Items      []MissingResponse `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

const dateFormat = "02/01/2006"

func toMissingResponse(m *missing.Missing) MissingResponse {
	resp := MissingResponse{
		ID:                m.ID,
		UserID:            m.UserID,
		Name:              m.Name,
		Nickname:          m.Nickname,
		Height:            m.Height,
		Clothes:           m.Clothes,
		Gender:            string(m.Gender),
		Eyes:              string(m.Eyes),
		Hair:              string(m.Hair),
		Skin:              string(m.Skin),
		PhotoURL:          m.PhotoURL,
		Lat:               m.Location.Lat,
		Lng:               m.Location.Lng,
		Status:            string(m.Status),
		EventReport:       m.EventReport,
		TattooDescription: m.TattooDescription,
		ScarDescription:   m.ScarDescription,
		WasChild:          m.WasChild,
		Slug:              m.Slug,
		HasTattoo:         m.HasTattoo(),
		HasScar:           m.HasScar(),
		Age:               m.Age(),
		CreatedAt:         m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         m.UpdatedAt.Format(time.RFC3339),
	}
	if !m.BirthDate.IsZero() {
		resp.BirthDate = m.BirthDate.Format(dateFormat)
	}
	if !m.DateOfDisappearance.IsZero() {
		resp.DateOfDisappearance = m.DateOfDisappearance.Format(dateFormat)
	}
	return resp
}

func parseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// @Summary      Cadastrar desaparecido
// @Description  Registra uma nova pessoa desaparecida
// @Tags         missing
// @Accept       json
// @Produce      json
// @Param        body  body      CreateMissingRequest  true  "Dados do desaparecido"
// @Success      201   {object}  MissingResponse
// @Failure      400   {object}  httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/missing [post]
func (h *MissingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateMissingRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r.Context())

	input := &missing.CreateInput{
		UserID:              userID,
		Name:                httputil.SanitizeString(req.Name),
		Nickname:            httputil.SanitizeString(req.Nickname),
		BirthDate:           parseDate(req.BirthDate),
		DateOfDisappearance: parseDate(req.DateOfDisappearance),
		Height:              httputil.SanitizeString(req.Height),
		Clothes:             httputil.SanitizeString(req.Clothes),
		Gender:              missing.Gender(req.Gender),
		Eyes:                missing.EyeColor(req.Eyes),
		Hair:                missing.HairColor(req.Hair),
		Skin:                missing.SkinColor(req.Skin),
		PhotoURL:            req.PhotoURL,
		Location:            missing.GeoPoint{Lat: req.Lat, Lng: req.Lng},
		EventReport:         httputil.SanitizeString(req.EventReport),
		TattooDescription:   httputil.SanitizeString(req.TattooDescription),
		ScarDescription:     httputil.SanitizeString(req.ScarDescription),
	}

	created, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, missing.ErrInvalidMissing) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to create missing person")
		return
	}

	httputil.JSON(w, http.StatusCreated, toMissingResponse(created))
}

// @Summary      Buscar desaparecido por ID
// @Description  Retorna dados de um desaparecido
// @Tags         missing
// @Produce      json
// @Param        id   path      string  true  "ID do desaparecido"
// @Success      200  {object}  MissingResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /api/v1/missing/{id} [get]
func (h *MissingHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	found, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, missing.ErrMissingNotFound) {
			httputil.Error(w, http.StatusNotFound, "missing person not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to find missing person")
		return
	}

	httputil.JSON(w, http.StatusOK, toMissingResponse(found))
}

// @Summary      Listar desaparecidos
// @Description  Retorna lista paginada de desaparecidos (cursor-based)
// @Tags         missing
// @Produce      json
// @Param        size   query     int     false  "Tamanho da página"  default(20)
// @Param        after  query     string  false  "Cursor para próxima página"
// @Success      200    {object}  MissingListResponse
// @Router       /api/v1/missing [get]
func (h *MissingHandler) List(w http.ResponseWriter, r *http.Request) {
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	after := r.URL.Query().Get("after")

	opts := missing.ListOptions{
		PageSize: size,
		After:    after,
	}

	items, nextCursor, err := h.service.List(r.Context(), opts)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list missing persons")
		return
	}

	resp := MissingListResponse{
		Items:      make([]MissingResponse, 0, len(items)),
		NextCursor: nextCursor,
	}
	for _, item := range items {
		resp.Items = append(resp.Items, toMissingResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Atualizar desaparecido
// @Description  Atualiza dados de um desaparecido (somente o dono)
// @Tags         missing
// @Accept       json
// @Produce      json
// @Param        id    path      string                true  "ID do desaparecido"
// @Param        body  body      UpdateMissingRequest  true  "Dados atualizados"
// @Success      200   {object}  MissingResponse
// @Failure      400   {object}  httputil.ErrorResponse
// @Failure      403   {object}  httputil.ErrorResponse
// @Failure      404   {object}  httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/missing/{id} [put]
func (h *MissingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req UpdateMissingRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	input := &missing.UpdateInput{
		Name:                httputil.SanitizeString(req.Name),
		Nickname:            httputil.SanitizeString(req.Nickname),
		BirthDate:           parseDate(req.BirthDate),
		DateOfDisappearance: parseDate(req.DateOfDisappearance),
		Height:              httputil.SanitizeString(req.Height),
		Clothes:             httputil.SanitizeString(req.Clothes),
		Gender:              missing.Gender(req.Gender),
		Eyes:                missing.EyeColor(req.Eyes),
		Hair:                missing.HairColor(req.Hair),
		Skin:                missing.SkinColor(req.Skin),
		PhotoURL:            req.PhotoURL,
		Location:            missing.GeoPoint{Lat: req.Lat, Lng: req.Lng},
		Status:              missing.Status(req.Status),
		EventReport:         httputil.SanitizeString(req.EventReport),
		TattooDescription:   httputil.SanitizeString(req.TattooDescription),
		ScarDescription:     httputil.SanitizeString(req.ScarDescription),
	}

	updated, err := h.service.Update(r.Context(), id, userID, input)
	if err != nil {
		switch {
		case errors.Is(err, missing.ErrMissingNotFound):
			httputil.Error(w, http.StatusNotFound, "missing person not found")
		case errors.Is(err, missing.ErrInvalidMissing):
			httputil.Error(w, http.StatusForbidden, err.Error())
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to update missing person")
		}
		return
	}

	httputil.JSON(w, http.StatusOK, toMissingResponse(updated))
}

// @Summary      Deletar desaparecido
// @Description  Remove um desaparecido (somente o dono)
// @Tags         missing
// @Param        id   path  string  true  "ID do desaparecido"
// @Success      204  "Sem conteúdo"
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/missing/{id} [delete]
func (h *MissingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	err := h.service.Delete(r.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, missing.ErrMissingNotFound):
			httputil.Error(w, http.StatusNotFound, "missing person not found")
		case errors.Is(err, missing.ErrInvalidMissing):
			httputil.Error(w, http.StatusForbidden, err.Error())
		default:
			httputil.Error(w, http.StatusInternalServerError, "failed to delete missing person")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Listar desaparecidos de um usuário
// @Description  Retorna todos os desaparecidos cadastrados por um usuário
// @Tags         missing
// @Produce      json
// @Param        id   path      string  true  "ID do usuário"
// @Success      200  {array}   MissingResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id}/missing [get]
func (h *MissingHandler) FindByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	items, err := h.service.FindByUserID(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list user's missing persons")
		return
	}

	resp := make([]MissingResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toMissingResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// @Summary      Buscar desaparecidos
// @Description  Busca por prefixo no nome
// @Tags         missing
// @Produce      json
// @Param        q      query     string  true   "Termo de busca"
// @Param        limit  query     int     false  "Limite de resultados"  default(20)
// @Success      200    {array}   MissingResponse
// @Router       /api/v1/missing/search [get]
func (h *MissingHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	items, err := h.service.Search(r.Context(), q, limit)
	if err != nil {
		if errors.Is(err, missing.ErrInvalidMissing) {
			httputil.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to search")
		return
	}

	resp := make([]MissingResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toMissingResponse(item))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// --- Stats DTOs ---

type StatsResponse struct {
	Total      int64           `json:"total"`
	ByGender   []GenderStatDTO `json:"by_gender"`
	ChildCount int64           `json:"child_count"`
	ByYear     []YearStatDTO   `json:"by_year"`
}

type GenderStatDTO struct {
	Gender string `json:"gender"`
	Count  int64  `json:"count"`
}

type YearStatDTO struct {
	Year  int   `json:"year"`
	Count int64 `json:"count"`
}

// @Summary      Estatísticas de desaparecidos
// @Description  Retorna totais, por gênero, crianças e por ano (goroutines paralelas)
// @Tags         missing
// @Produce      json
// @Success      200  {object}  StatsResponse
// @Router       /api/v1/missing/stats [get]
func (h *MissingHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	resp := StatsResponse{
		Total:      stats.Total,
		ChildCount: stats.ChildCount,
		ByGender:   make([]GenderStatDTO, 0, len(stats.ByGender)),
		ByYear:     make([]YearStatDTO, 0, len(stats.ByYear)),
	}
	for _, g := range stats.ByGender {
		resp.ByGender = append(resp.ByGender, GenderStatDTO{Gender: g.Gender, Count: g.Count})
	}
	for _, y := range stats.ByYear {
		resp.ByYear = append(resp.ByYear, YearStatDTO{Year: y.Year, Count: y.Count})
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// --- Locations DTOs ---

type LocationPointDTO struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Status string  `json:"status"`
}

type LocationsResponse struct {
	Locations []LocationPointDTO `json:"locations"`
}

// @Summary      Localizações de desaparecidos
// @Description  Retorna coordenadas para exibição no mapa
// @Tags         missing
// @Produce      json
// @Param        limit  query     int  false  "Limite"  default(100)
// @Success      200    {object}  LocationsResponse
// @Router       /api/v1/missing/locations [get]
func (h *MissingHandler) Locations(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	points, err := h.service.FindLocations(r.Context(), limit)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get locations")
		return
	}

	resp := LocationsResponse{
		Locations: make([]LocationPointDTO, 0, len(points)),
	}
	for _, p := range points {
		resp.Locations = append(resp.Locations, LocationPointDTO{
			ID:     p.ID,
			Name:   p.Name,
			Lat:    p.Lat,
			Lng:    p.Lng,
			Status: string(p.Status),
		})
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// --- Age Progression ---

type AgeProgressionResponse struct {
	MissingID string   `json:"missing_id"`
	URLs      []string `json:"urls"`
}

// @Summary      Obter projeções de idade
// @Description  Retorna URLs das imagens de age progression geradas por IA
// @Tags         missing
// @Produce      json
// @Param        id  path  string  true  "Missing ID"
// @Success      200  {object}  AgeProgressionResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /api/v1/missing/{id}/age-progression [get]
func (h *MissingHandler) GetAgeProgression(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	m, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, missing.ErrMissingNotFound) {
			httputil.Error(w, http.StatusNotFound, "missing not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to find missing")
		return
	}

	urls := m.AgeProgressionURLs
	if urls == nil {
		urls = []string{}
	}

	httputil.JSON(w, http.StatusOK, AgeProgressionResponse{
		MissingID: id,
		URLs:      urls,
	})
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// @Summary      Alterar status do desaparecido
// @Description  Marca como encontrado ou reativa busca
// @Tags         missing
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Missing ID"
// @Param        body  body      UpdateStatusRequest   true  "Status"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  httputil.ErrorResponse
// @Failure      404   {object}  httputil.ErrorResponse
// @Router       /api/v1/missing/{id}/status [patch]
func (h *MissingHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateStatusRequest
	if err := httputil.DecodeAndValidate(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	status := missing.Status(req.Status)
	if !status.IsValid() {
		httputil.Error(w, http.StatusBadRequest, "invalid status, use 'disappeared' or 'found'")
		return
	}

	existing, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, missing.ErrMissingNotFound) {
			httputil.Error(w, http.StatusNotFound, "missing not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to find missing")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if existing.UserID != userID {
		httputil.Error(w, http.StatusForbidden, "only the owner can change status")
		return
	}

	existing.Status = status
	if err := h.service.UpdateEntity(r.Context(), existing); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": string(status)})
}
