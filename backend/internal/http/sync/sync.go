package sync

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/ConradKurth/forecasting/backend/internal/http/response"
	"github.com/ConradKurth/forecasting/backend/internal/manager"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/go-chi/chi/v5"
)

// InitRoutes initializes sync-related routes
func InitRoutes(r *chi.Mux, syncManager *manager.InventorySyncManager) {
	r.Route("/v1/sync", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/shopify", response.Wrap(TriggerShopifySync(syncManager)))
		r.Get("/status/{shop_domain}", response.Wrap(GetSyncStatus(syncManager)))
	})
}

// TriggerSyncRequest represents a sync trigger request
type TriggerSyncRequest struct {
	ShopDomain string `json:"shop_domain"`
	Force      bool   `json:"force,omitempty"`
}

// SyncStatusResponse represents the response for sync status
type SyncStatusResponse struct {
	IntegrationID string    `json:"integration_id"`
	Status        string    `json:"status"`
	LastSynced    time.Time `json:"last_synced,omitempty"`
	Error         string    `json:"error,omitempty"`
}

// TriggerShopifySync triggers a sync for a Shopify shop
// POST /api/v1/sync/shopify
func TriggerShopifySync(syncManager *manager.InventorySyncManager) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			return response.InternalServerError("User not found in context", nil)
		}

		userID, err := id.New[id.User](user.UserID)
		if err != nil {
			logger.Error("Invalid user ID", "user_id", user.UserID, "error", err)
			return response.BadRequest("Invalid user ID", nil)
		}

		var req TriggerSyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("Failed to decode sync request", "error", err)
			return response.BadRequest("Invalid request body", nil)
		}

		if req.ShopDomain == "" {
			return response.BadRequest("shop_domain is required", nil)
		}

		// Delegate to manager
		result, err := syncManager.TriggerShopifySync(r.Context(), manager.SyncRequest{
			UserID:     userID,
			ShopDomain: req.ShopDomain,
			Force:      req.Force,
		})
		if err != nil {
			logger.Error("Sync trigger failed", "error", err, "user_id", userID, "shop_domain", req.ShopDomain)
			return response.InternalServerError("Failed to trigger sync", err)
		}

		return response.JSON(w, http.StatusOK, SyncStatusResponse{
			IntegrationID: result.IntegrationID,
			Status:        string(result.Status),
			Error:         result.Error,
		})
	}
}

// GetSyncStatus gets the sync status for a shop
// GET /api/v1/sync/status/{shop_domain}
func GetSyncStatus(syncManager *manager.InventorySyncManager) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			return response.InternalServerError("User not found in context", nil)
		}

		userID, err := id.New[id.User](user.UserID)
		if err != nil {
			logger.Error("Invalid user ID", "user_id", user.UserID, "error", err)
			return response.BadRequest("Invalid user ID", nil)
		}

		shopDomain := chi.URLParam(r, "shop_domain")
		if shopDomain == "" {
			return response.BadRequest("shop_domain is required", nil)
		}

		// Delegate to manager
		result, err := syncManager.GetSyncStatus(r.Context(), userID, shopDomain)
		if err != nil {
			logger.Error("Failed to get sync status", "error", err, "user_id", userID, "shop_domain", shopDomain)
			return response.InternalServerError("Failed to get sync status", err)
		}

		return response.JSON(w, http.StatusOK, SyncStatusResponse{
			IntegrationID: result.IntegrationID,
			Status:        string(result.Status),
			Error:         result.Error,
		})
	}
}
