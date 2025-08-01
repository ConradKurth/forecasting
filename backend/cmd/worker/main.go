package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/internal/worker"
	"github.com/hibiken/asynq"
)

func main() {
	// Initialize database (config is loaded automatically)
	database, err := db.New()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create service factory
	serviceFactory := factory.NewServiceFactory(database)

	// Create Redis connection config
	redisOpt := asynq.RedisClientOpt{
		Addr: config.Values.Redis.URL,
	}

	// Create worker server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// Create worker with services
	w := worker.New(serviceFactory)

	// Register task handlers
	mux := asynq.NewServeMux()
	w.RegisterHandlers(mux)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Worker server failed: %v", err)
		}
	}()

	log.Println("Worker server started. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down worker server...")

	srv.Shutdown()
	log.Println("Worker server stopped.")
}
