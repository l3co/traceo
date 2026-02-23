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
	"github.com/l3co/traceo-api/internal/domain/missing"
	"github.com/l3co/traceo-api/internal/domain/sighting"
	"github.com/l3co/traceo-api/internal/domain/user"
	"github.com/l3co/traceo-api/internal/handler"
	"github.com/l3co/traceo-api/internal/handler/middleware"
	"github.com/l3co/traceo-api/internal/i18n"
	"github.com/l3co/traceo-api/internal/infrastructure/firebase"
	"github.com/l3co/traceo-api/internal/infrastructure/notification"

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
	notifier := notification.NewService(emailSender)

	sightingRepo := firebase.NewSightingRepository(fbClient.Firestore)
	sightingService := sighting.NewService(sightingRepo, missingRepo, notifier)

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(userService)
	missingHandler := handler.NewMissingHandler(missingService)
	sightingHandler := handler.NewSightingHandler(sightingService)

	r := setupRouter(cfg, authService, userHandler, authHandler, missingHandler, sightingHandler)

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

	healthHandler := handler.NewHealthHandler()

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.Check)

		r.Post("/users", userHandler.Create)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)

		r.Get("/missing", missingHandler.List)
		r.Get("/missing/search", missingHandler.Search)
		r.Get("/missing/stats", missingHandler.Stats)
		r.Get("/missing/locations", missingHandler.Locations)
		r.Get("/missing/{id}", missingHandler.FindByID)
		r.Get("/missing/{id}/sightings", sightingHandler.FindByMissingID)
		r.Get("/sightings/{sightingId}", sightingHandler.FindByID)

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
		})
	})

	return r
}
