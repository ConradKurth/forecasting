package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/http/dashboard"
	"github.com/ConradKurth/forecasting/backend/internal/http/oauth"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Initialize logger
	logger.Init(logger.Level(config.Values.Logging.Level))
	
	// Run database migrations
	logger.Info("Running database migrations...")
	if err := db.RunMigrations(config.Values.Database.URL); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		panic(err)
	}
	logger.Info("Database migrations completed successfully")
	
	// Initialize database
	database, err := db.New()
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		panic(err)
	}
	defer database.Close()

	// Initialize services
	userService := service.NewUserService(database.Users)

	r := chi.NewRouter()

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: config.Values.CORS.AllowedOrigins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Add other middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize routes
	oauth.InitRoutes(r, userService)
	dashboard.InitRoutes(r, userService)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":3001",
		Handler: r,
	}

	// Channel to listen for interrupt signal to terminate gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening on :3001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			panic(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-quit
	logger.Info("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
		panic(err)
	}

	logger.Info("Server exited")
}
