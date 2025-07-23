package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux) {
	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/profile", GetProfile)
		r.Get("/data", GetDashboardData)
	})
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"shop":   user.Shop,
		"userId": user.UserID,
	})
}

func GetDashboardData(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Welcome to your dashboard!",
		"shop":    user.Shop,
		"data":    []string{"Sample data 1", "Sample data 2", "Sample data 3"},
	})
}
