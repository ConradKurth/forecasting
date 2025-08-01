package dashboard

import (
	"log"
	"net/http"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/ConradKurth/forecasting/backend/internal/http/response"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux, userService *service.UserService) {
	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/profile", response.Wrap(GetProfile(userService)))
		r.Get("/data", response.Wrap(GetDashboardData))
	})
}

func GetProfile(userService *service.UserService) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			return response.InternalServerError("User not found in context", nil)
		}

		// Get full user details from the database
		dbUser, err := userService.GetUser(r.Context(), user.Shop)
		if err != nil {
			log.Printf("Failed to get user details for shop %s: %v", user.Shop, err)
			return response.InternalServerError("Failed to get user details", err)
		}

		if dbUser == nil {
			return response.UserNotFound()
		}

		return response.JSON(w, http.StatusOK, map[string]interface{}{
			"id":         dbUser.ID,
			"shop":       dbUser.ShopDomain,
			"userId":     user.UserID,
			"createdAt":  dbUser.CreatedAt,
			"updatedAt":  dbUser.UpdatedAt,
		})
	}
}

func GetDashboardData(w http.ResponseWriter, r *http.Request) error {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		return response.InternalServerError("User not found in context", nil)
	}

	return response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Welcome to your dashboard!",
		"shop":    user.Shop,
		"data":    []string{"Sample data 1", "Sample data 2", "Sample data 3"},
	})
}
