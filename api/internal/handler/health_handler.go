package handler

import (
	"context"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/l3co/traceo-api/pkg/httputil"
)

type HealthHandler struct {
	startTime       time.Time
	firestoreClient *firestore.Client
	version         string
}

func NewHealthHandler(firestoreClient *firestore.Client, version string) *HealthHandler {
	return &HealthHandler{
		startTime:       time.Now(),
		firestoreClient: firestoreClient,
		version:         version,
	}
}

// @Summary      Health check
// @Description  Verifica se a API está operacional e suas dependências
// @Tags         system
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Router       /api/v1/health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	deps := map[string]string{}

	if h.firestoreClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		iter := h.firestoreClient.Collection("_health").Limit(1).Documents(ctx)
		_, _ = iter.Next()
		iter.Stop()
		deps["firestore"] = "ok"
	}

	httputil.JSON(w, http.StatusOK, HealthResponse{
		Status:       "ok",
		Version:      h.version,
		Uptime:       time.Since(h.startTime).String(),
		Dependencies: deps,
	})
}

type HealthResponse struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	Uptime       string            `json:"uptime"`
	Dependencies map[string]string `json:"dependencies"`
}
