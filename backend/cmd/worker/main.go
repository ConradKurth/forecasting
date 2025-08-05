package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/manager"
	"github.com/ConradKurth/forecasting/backend/internal/worker"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init(logger.Level(config.Values.Logging.Level))

	// Initialize database (config is loaded automatically)
	database, err := db.New()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	logger.Info("Database connected successfully")

	// Initialize worker queue client
	workerQueue := worker.NewClient()
	defer func() {
		if err := workerQueue.Close(); err != nil {
			logger.Error("Failed to close worker queue", "error", err)
		}
	}()

	// Initialize managers
	shopifyManager := manager.NewShopifyManager(database, workerQueue)
	syncManager := manager.NewInventorySyncManager(database, shopifyManager, workerQueue)

	// Create worker server with middleware and proper configuration
	server := worker.NewServer(shopifyManager, syncManager)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Run(); err != nil {
			logger.Error("Worker server failed", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("Worker server started. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutting down worker server...")

	server.Shutdown()
	logger.Info("Worker server stopped.")
}
