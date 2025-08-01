package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/http/dashboard"
	"github.com/ConradKurth/forecasting/backend/internal/http/oauth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {

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
	oauth.InitRoutes(r)
	dashboard.InitRoutes(r)

	fmt.Println("Server listening on :3001")
	log.Fatal(http.ListenAndServe(":3001", r))
}
