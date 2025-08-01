package worker

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/hibiken/asynq"
)

// Server wraps asynq.Server with our configuration
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer creates a new worker server with proper configuration and middleware
func NewServer(serviceFactory *factory.ServiceFactory) *Server {
	// Create Redis connection config
	redisOpt := asynq.RedisClientOpt{
		Addr: config.Values.Redis.URL,
	}

	// Create worker server with configuration
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// Add error handler
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Error("Task failed", "type", task.Type(), "payload", string(task.Payload()), "error", err)
			}),
		},
	)

	// Create multiplexer with middleware
	mux := asynq.NewServeMux()
	
	// Add logging middleware
	mux.Use(loggingMiddleware())
	
	// Add recovery middleware
	mux.Use(recoveryMiddleware())

	// Create worker and register handlers
	worker := New(serviceFactory)
	worker.RegisterHandlers(mux)

	return &Server{
		server: srv,
		mux:    mux,
	}
}

// Run starts the worker server
func (s *Server) Run() error {
	return s.server.Run(s.mux)
}

// Shutdown gracefully shuts down the worker server
func (s *Server) Shutdown() {
	s.server.Shutdown()
}

// loggingMiddleware logs information about job processing
func loggingMiddleware() asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			start := time.Now()
			
			logger.Info("Started processing task", "type", task.Type(), "payload", string(task.Payload()))
			
			err := next.ProcessTask(ctx, task)
			
			duration := time.Since(start)
			if err != nil {
				logger.Error("Failed to process task", "type", task.Type(), "duration", duration, "error", err)
			} else {
				logger.Info("Successfully processed task", "type", task.Type(), "duration", duration)
			}
			
			return err
		})
	}
}

// recoveryMiddleware recovers from panics in task handlers
func recoveryMiddleware() asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic recovered in task handler", "type", task.Type(), "panic", r)
					err = asynq.SkipRetry
				}
			}()
			
			return next.ProcessTask(ctx, task)
		})
	}
}
