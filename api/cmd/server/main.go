package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/time/rate"

	"github.com/l3co/traceo-api/internal/config"
	"github.com/l3co/traceo-api/internal/domain/homeless"
	"github.com/l3co/traceo-api/internal/domain/matching"
	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/domain/sighting"
	"github.com/l3co/traceo-api/internal/domain/user"
	"github.com/l3co/traceo-api/internal/handler"
	"github.com/l3co/traceo-api/internal/handler/middleware"
	"github.com/l3co/traceo-api/internal/i18n"
	"github.com/l3co/traceo-api/internal/infrastructure/ai"
	"github.com/l3co/traceo-api/internal/infrastructure/firebase"
	"github.com/l3co/traceo-api/internal/infrastructure/notification"
	"github.com/l3co/traceo-api/internal/worker"

	_ "github.com/l3co/traceo-api/docs/swagger"
)

// @title           Traceo API
// @version         0.1.0
// @description     API para a plataforma Traceo — localização de pessoas desaparecidas no Brasil.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @accept          json
// @produce         json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg := config.Load()

	setupLogger(cfg)

	if err := i18n.Init(cfg.DefaultLanguage); err != nil {
		slog.Error("failed to initialize i18n", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx := context.Background()
	fbClient, err := firebase.NewClient(ctx, cfg.FirebaseProjectID)
	if err != nil {
		slog.Error("failed to initialize firebase", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer fbClient.Close()

	authService := firebase.NewAuthService(fbClient.Auth)
	userRepo := firebase.NewUserRepository(fbClient.Firestore)
	userService := user.NewService(userRepo, authService)

	missingRepo := firebase.NewMissingRepository(fbClient.Firestore)
	missingService := missing.NewService(missingRepo)

	var emailSender *notification.EmailSender
	if cfg.ResendAPIKey != "" {
		emailSender = notification.NewEmailSender(cfg.ResendAPIKey, cfg.ResendFromEmail)
	}
	var telegramSender *notification.TelegramSender
	if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		telegramSender = notification.NewTelegramSender(cfg.TelegramBotToken, cfg.TelegramChatID)
	}
	notifier := notification.NewService(emailSender, telegramSender)

	sightingRepo := firebase.NewSightingRepository(fbClient.Firestore)
	sightingService := sighting.NewService(sightingRepo, missingRepo, notifier)

	homelessRepo := firebase.NewHomelessRepository(fbClient.Firestore)
	homelessService := homeless.NewService(homelessRepo, notifier)

	matchRepo := firebase.NewMatchRepository(fbClient.Firestore)

	var faceComparer matching.FaceComparer
	var aiWorker *worker.AIWorker
	if cfg.GeminiAPIKey != "" {
		geminiClient, err := ai.NewGeminiClient(ctx, cfg.GeminiAPIKey)
		if err != nil {
			slog.Error("failed to initialize gemini client", slog.String("error", err.Error()))
		} else {
			defer geminiClient.Close()
			faceComparer = newGeminiComparer(geminiClient)
		}
	}

	var faceDescriber matching.FaceDescriber
	if faceComparer != nil {
		faceDescriber = faceComparer.(*geminiComparer)
	}
	matchingService := matching.NewService(missingRepo, homelessRepo, matchRepo, faceComparer, faceDescriber, notifier)

	if faceComparer != nil {
		aiWorker = worker.NewAIWorker(matchingService, 3)
		defer aiWorker.Shutdown()
	}
	_ = aiWorker

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(userService)
	missingHandler := handler.NewMissingHandler(missingService)
	sightingHandler := handler.NewSightingHandler(sightingService)
	homelessHandler := handler.NewHomelessHandler(homelessService)
	matchHandler := handler.NewMatchHandler(matchingService)
	metaHandler := handler.NewMetaHandler(missingService)
	sitemapHandler := handler.NewSitemapHandler(missingService, homelessService)
	healthHandler := handler.NewHealthHandler(fbClient.Firestore, "1.0.0")

	r := setupRouter(cfg, authService, userHandler, authHandler, missingHandler, sightingHandler, homelessHandler, matchHandler, metaHandler, sitemapHandler, healthHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting",
			slog.String("port", cfg.Port),
			slog.String("environment", cfg.Environment),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	sig := <-quit
	slog.Info("shutdown signal received", slog.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", slog.String("error", err.Error()))
	}

	slog.Info("server stopped gracefully")
}

type geminiComparer struct {
	client *ai.GeminiClient
}

func newGeminiComparer(client *ai.GeminiClient) *geminiComparer {
	return &geminiComparer{client: client}
}

func (g *geminiComparer) DescribeFace(ctx context.Context, photoURL string, currentAge int, gender string) (string, error) {
	return g.client.DescribeFace(ctx, photoURL, currentAge, gender)
}

func (g *geminiComparer) CompareFaces(ctx context.Context, photo1URL, photo2URL string) (*matching.FaceComparisonResult, error) {
	result, err := g.client.CompareFaces(ctx, photo1URL, photo2URL)
	if err != nil {
		return nil, err
	}
	return &matching.FaceComparisonResult{
		SimilarityScore:   result.SimilarityScore,
		Analysis:          result.Analysis,
		MatchingFeatures:  result.MatchingFeatures,
		DifferentFeatures: result.DifferentFeatures,
		Confidence:        result.Confidence,
	}, nil
}

func setupLogger(cfg *config.Config) {
	var h slog.Handler
	if cfg.IsDevelopment() {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(h))
}

func setupRouter(
	cfg *config.Config,
	authService *firebase.AuthService,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	missingHandler *handler.MissingHandler,
	sightingHandler *handler.SightingHandler,
	homelessHandler *handler.HomelessHandler,
	matchHandler *handler.MatchHandler,
	metaHandler *handler.MetaHandler,
	sitemapHandler *handler.SitemapHandler,
	healthHandler *handler.HealthHandler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.BodyLimit(1 << 20)) // 1 MB
	r.Use(i18n.Middleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	globalLimiter := middleware.NewRateLimiter(rate.Every(time.Second/4), 50) // ~200 req/min
	r.Use(globalLimiter.Handler)

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/robots.txt", metaHandler.RobotsTxt)
	r.Get("/sitemap.xml", sitemapHandler.Serve)
	r.Get("/share/missing/{id}", metaHandler.ServeMissingMeta)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.Check)

		r.Post("/users", userHandler.Create)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)

		r.Get("/missing", missingHandler.List)
		r.Get("/missing/search", missingHandler.Search)
		r.Get("/missing/stats", missingHandler.Stats)
		r.Get("/missing/locations", missingHandler.Locations)
		r.Get("/missing/{id}", missingHandler.FindByID)
		r.Get("/missing/{id}/age-progression", missingHandler.GetAgeProgression)
		r.Get("/missing/{id}/sightings", sightingHandler.FindByMissingID)
		r.Get("/sightings/{sightingId}", sightingHandler.FindByID)

		r.Get("/homeless", homelessHandler.List)
		r.Get("/homeless/stats", homelessHandler.Stats)
		r.Get("/homeless/{id}", homelessHandler.FindByID)
		r.Post("/homeless", homelessHandler.Create)

		r.Get("/homeless/{id}/matches", matchHandler.FindByHomelessID)
		r.Get("/missing/{id}/matches", matchHandler.FindByMissingID)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))

			r.Get("/users/{id}", userHandler.FindByID)
			r.Put("/users/{id}", userHandler.Update)
			r.Delete("/users/{id}", userHandler.Delete)
			r.Patch("/users/{id}/password", userHandler.ChangePassword)

			r.Post("/missing", missingHandler.Create)
			r.Put("/missing/{id}", missingHandler.Update)
			r.Delete("/missing/{id}", missingHandler.Delete)
			r.Get("/users/{id}/missing", missingHandler.FindByUserID)

			r.Post("/missing/{id}/sightings", sightingHandler.Create)
			r.Patch("/missing/{id}/status", missingHandler.UpdateStatus)
			r.Patch("/matches/{id}", matchHandler.UpdateStatus)
		})
	})

	return r
}
