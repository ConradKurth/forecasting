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

	fmt.Println(config.Values.CORS.AllowedOrigins)
	for _, origin := range config.Values.CORS.AllowedOrigins {
		fmt.Println(origin)
	}
	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"http://localhost:5173", "https://ea0b79250aad.ngrok.app"},
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
		Debug:            true,
	}))
	// r.Use(cors.AllowAll().Handler)

	// Add other middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Initialize routes
	oauth.InitRoutes(r)
	dashboard.InitRoutes(r)

	fmt.Println("Server listening on :3001")
	log.Fatal(http.ListenAndServe(":3001", r))
}
