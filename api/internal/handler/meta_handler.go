package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/l3co/traceo-api/internal/domain/missing"
)

type MetaHandler struct {
	missingService *missing.Service
}

func NewMetaHandler(missingService *missing.Service) *MetaHandler {
	return &MetaHandler{missingService: missingService}
}

var botUserAgents = []string{
	"googlebot", "bingbot", "facebookexternalhit",
	"twitterbot", "whatsapp", "telegrambot",
	"linkedinbot", "slackbot", "discordbot",
}

func isBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, bot := range botUserAgents {
		if strings.Contains(ua, bot) {
			return true
		}
	}
	return false
}

// @Summary      Meta tags para bots (Open Graph)
// @Description  Retorna HTML com Open Graph tags para compartilhamento social
// @Tags         seo
// @Produce      html
// @Param        id   path  string  true  "Missing ID"
// @Router       /share/missing/{id} [get]
func (h *MetaHandler) ServeMissingMeta(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !isBot(r.UserAgent()) {
		http.Redirect(w, r, fmt.Sprintf("/missing/%s", id), http.StatusFound)
		return
	}

	m, err := h.missingService.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	dateStr := ""
	if !m.DateOfDisappearance.IsZero() {
		dateStr = m.DateOfDisappearance.Format("02/01/2006")
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8" />
    <title>%s - Desaparecido | Traceo</title>
    <meta name="description" content="Desaparecido desde %s. Ajude a encontrar." />
    <meta property="og:title" content="%s - Desaparecido" />
    <meta property="og:description" content="Desaparecido desde %s. Ajude a encontrar." />
    <meta property="og:image" content="%s" />
    <meta property="og:url" content="https://traceo.me/missing/%s" />
    <meta property="og:type" content="article" />
    <meta property="og:site_name" content="Traceo" />
    <meta name="twitter:card" content="summary_large_image" />
</head>
<body>
    <script>window.location.href='/missing/%s';</script>
    <noscript><a href="/missing/%s">Ver perfil de %s</a></noscript>
</body>
</html>`, m.Name, dateStr, m.Name, dateStr, m.PhotoURL, m.ID, m.ID, m.ID, m.Name)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (h *MetaHandler) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	body := `User-agent: *
Allow: /
Sitemap: https://traceo.me/sitemap.xml
`
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(body))
}
