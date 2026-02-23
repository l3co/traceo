package handler

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/missing"
)

type SitemapHandler struct {
	missingService  *missing.Service
	homelessService *homeless.Service
}

func NewSitemapHandler(missingService *missing.Service, homelessService *homeless.Service) *SitemapHandler {
	return &SitemapHandler{
		missingService:  missingService,
		homelessService: homelessService,
	}
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
	Freq    string `xml:"changefreq,omitempty"`
	Prio    string `xml:"priority,omitempty"`
}

// @Summary      Sitemap XML
// @Description  Sitemap dinâmico para SEO
// @Tags         seo
// @Produce      xml
// @Router       /sitemap.xml [get]
func (h *SitemapHandler) Serve(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")

	urls := []sitemapURL{
		{Loc: "https://traceo.me/", Freq: "daily", Prio: "1.0", LastMod: today},
		{Loc: "https://traceo.me/missing", Freq: "daily", Prio: "0.9", LastMod: today},
		{Loc: "https://traceo.me/homeless", Freq: "daily", Prio: "0.8", LastMod: today},
		{Loc: "https://traceo.me/faq", Freq: "monthly", Prio: "0.5"},
		{Loc: "https://traceo.me/terms", Freq: "monthly", Prio: "0.3"},
		{Loc: "https://traceo.me/privacy", Freq: "monthly", Prio: "0.3"},
	}

	missingList, _, _ := h.missingService.List(r.Context(), missing.ListOptions{PageSize: 500})
	for _, m := range missingList {
		lastMod := m.UpdatedAt.Format("2006-01-02")
		urls = append(urls, sitemapURL{
			Loc:     "https://traceo.me/missing/" + m.ID,
			LastMod: lastMod,
			Freq:    "weekly",
			Prio:    "0.8",
		})
	}

	homelessList, _ := h.homelessService.FindAll(r.Context())
	for _, h := range homelessList {
		urls = append(urls, sitemapURL{
			Loc:     "https://traceo.me/homeless/" + h.ID,
			LastMod: h.CreatedAt.Format("2006-01-02"),
			Freq:    "weekly",
			Prio:    "0.7",
		})
	}

	set := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(set)
}
