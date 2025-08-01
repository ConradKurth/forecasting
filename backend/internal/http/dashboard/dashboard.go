package dashboard

import (
	"log"
	"net/http"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/ConradKurth/forecasting/backend/internal/http/response"
	"github.com/ConradKurth/forecasting/backend/internal/manager"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux, shopifyManager *manager.ShopifyManager) {
	r.Route("/v1/dashboard", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/profile", response.Wrap(GetProfile(shopifyManager)))
		r.Get("/data", response.Wrap(GetDashboardData))
	})
}

func GetProfile(shopifyManager *manager.ShopifyManager) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			return response.InternalServerError("User not found in context", nil)
		}

		// Convert the user ID from string to typed ID
		userID, err := id.New[id.User](user.UserID)
		if err != nil {
			log.Printf("Failed to parse user ID %s: %v", user.UserID, err)
			return response.InternalServerError("Invalid user ID", err)
		}

		// Get the full shopify integration
		integration, err := shopifyManager.GetShopifyIntegration(r.Context(), userID, user.Shop)
		if err != nil {
			log.Printf("Failed to get shopify integration for user %s and shop %s: %v", userID, user.Shop, err)
			return response.InternalServerError("Failed to get user details", err)
		}

		return response.JSON(w, http.StatusOK, map[string]interface{}{
			"id":             integration.User.ID,
			"shop":           integration.Store.ShopDomain,
			"shop_name":      integration.Store.ShopName,
			"userId":         user.UserID,
			"createdAt":      integration.User.CreatedAt,
			"updatedAt":      integration.User.UpdatedAt,
			"store":          integration.Store,
			"shopify_user":   integration.ShopifyUser,
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
