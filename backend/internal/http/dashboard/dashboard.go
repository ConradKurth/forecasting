package dashboard

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux, userService *service.UserService) {
	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/profile", GetProfile(userService))
		r.Get("/data", GetDashboardData)
	})
}

func GetProfile(userService *service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}

		// Get full user details from the database
		dbUser, err := userService.GetUser(r.Context(), user.Shop)
		if err != nil {
			log.Printf("Failed to get user details for shop %s: %v", user.Shop, err)
			http.Error(w, "Failed to get user details", http.StatusInternalServerError)
			return
		}

		if dbUser == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         dbUser.ID,
			"shop":       dbUser.ShopDomain,
			"userId":     user.UserID,
			"createdAt":  dbUser.CreatedAt,
			"updatedAt":  dbUser.UpdatedAt,
		})
	}
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
