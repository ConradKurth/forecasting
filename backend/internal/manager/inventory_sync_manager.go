package manager

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

// InventorySyncManager orchestrates inventory synchronization operations
// It ensures data consistency by wrapping operations in database transactions
// and coordinates between multiple services for complex sync workflows
type InventorySyncManager struct {
	database       *db.DB
	serviceFactory *factory.ServiceInterfaceFactory
	shopifyManager *ShopifyManager
}

// NewInventorySyncManager creates a new InventorySyncManager instance
func NewInventorySyncManager(database *db.DB, shopifyManager *ShopifyManager, serviceFactory *factory.ServiceInterfaceFactory) *InventorySyncManager {
	return &InventorySyncManager{
		database:       database,
		serviceFactory: serviceFactory,
		shopifyManager: shopifyManager,
	}
}

// SyncRequest represents a synchronization request
type SyncRequest struct {
	UserID     id.ID[id.User] `json:"user_id"`
	ShopDomain string         `json:"shop_domain"`
	Force      bool           `json:"force,omitempty"`
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	IntegrationID string `json:"integration_id"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
}

// TriggerShopifySync orchestrates a complete Shopify synchronization process
// This includes validating access, creating integrations, and triggering async sync
func (m *InventorySyncManager) TriggerShopifySync(ctx context.Context, req SyncRequest) (*SyncResult, error) {
	var result *SyncResult

	err := m.database.WithTx(ctx, func(txDB *db.TxDB) error {
		// Create transactional services
		txServices := m.serviceFactory.FromTx(txDB)

		// Validate user exists
		user, err := txServices.User.GetUser(ctx, req.UserID)
		if err != nil {
			return errors.Wrap(err, "failed to get user")
		}
		if user == nil {
			return errors.New("user not found")
		}

		// Get shop and validate it exists
		shop, err := txServices.ShopifyStore.GetStoreByDomain(ctx, req.ShopDomain)
		if err != nil || shop == nil {
			return errors.Wrap(err, "shop not found")
		}

		// Get access token using services
		accessToken, err := txServices.ShopifyUser.GetShopifyAccessToken(ctx, req.UserID, req.ShopDomain)
		if err != nil {
			return errors.Wrap(err, "failed to get access token")
		}
		if accessToken == "" {
			return errors.New("no access token found for user and shop")
		}

		// Create or get platform integration
		integration, err := m.getOrCreateShopifyIntegration(ctx, shop.ID, req.ShopDomain)
		if err != nil {
			return errors.Wrap(err, "failed to get/create integration")
		}

		// Check if sync should proceed
		if !req.Force {
			shouldSkip, skipReason, err := m.shouldSkipSync(ctx, integration.ID)
			if err != nil {
				return errors.Wrap(err, "failed to check sync status")
			}
			if shouldSkip {
				result = &SyncResult{
					IntegrationID: integration.ID.String(),
					Status:        string(skipReason),
				}
				return nil
			}
		}

		// Trigger async sync
		go func() {
			syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute) // 30 minutes
			defer cancel()

			inventoryService := service.NewCoreInventoryService(txDB.GetCore())

			err := inventoryService.SyncShopifyData(syncCtx, integration.ID, req.ShopDomain, accessToken)
			if err != nil {
				logger.Error("Sync failed",
					"integration_id", integration.ID,
					"shop_domain", req.ShopDomain,
					"user_id", req.UserID,
					"error", err,
				)
			}
		}()

		result = &SyncResult{
			IntegrationID: integration.ID.String(),
			Status:        "sync_started",
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetSyncStatus retrieves the current synchronization status for a shop
func (m *InventorySyncManager) GetSyncStatus(ctx context.Context, userID id.ID[id.User], shopDomain string) (*SyncResult, error) {
	shopStore := m.serviceFactory.FromDB().ShopifyStore
	coreStore := m.serviceFactory.FromDB().Core

	// Get shop
	shop, err := shopStore.GetStoreByDomain(ctx, shopDomain)
	if err != nil || shop == nil {
		return nil, errors.Wrap(err, "shop not found")
	}

	// Get integration
	integration, err := coreStore.GetPlatformIntegrationByShopAndType(ctx, core.GetPlatformIntegrationByShopAndTypeParams{
		ShopID:       shop.ID,
		PlatformType: core.PlatformTypeShopify,
	})
	if err != nil {
		return nil, errors.Wrap(err, "integration not found")
	}

	// Get sync states
	syncStates, err := coreStore.GetSyncStatesByIntegrationID(ctx, integration.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sync states")
	}

	result := &SyncResult{
		IntegrationID: integration.ID.String(),
	}

	if len(syncStates) == 0 {
		result.Status = "never_synced"
	} else {
		// Find the most recent full_sync
		for _, state := range syncStates {
			if state.EntityType == core.EntityTypeFullSync {
				result.Status = string(state.SyncStatus)
				if state.ErrorMessage.Valid {
					result.Error = state.ErrorMessage.String
				}
				break
			}
		}
		if result.Status == "" {
			result.Status = "partial_sync_only"
		}
	}

	return result, nil
}

// getOrCreateShopifyIntegration gets or creates a Shopify platform integration
func (m *InventorySyncManager) getOrCreateShopifyIntegration(ctx context.Context, shopID id.ID[id.ShopifyStore], shopDomain string) (core.PlatformIntegration, error) {
	// Try to get existing integration

	services := m.serviceFactory.FromDB()

	integration, err := services.Core.GetPlatformIntegrationByShopAndType(ctx, core.GetPlatformIntegrationByShopAndTypeParams{
		ShopID:       shopID,
		PlatformType: core.PlatformTypeShopify,
	})
	if err == nil {
		return integration, nil
	}

	// Create new integration
	integration, err = services.Core.CreatePlatformIntegration(ctx, core.CreatePlatformIntegrationParams{
		ID:             id.NewGeneration[id.PlatformIntegration](),
		ShopID:         shopID,
		PlatformType:   core.PlatformTypeShopify,
		PlatformShopID: shopDomain,
		IsActive:       pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		return core.PlatformIntegration{}, errors.Wrap(err, "failed to create platform integration")
	}

	return integration, nil
}

// shouldSkipSync determines if a sync should be skipped based on current state
func (m *InventorySyncManager) shouldSkipSync(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) (bool, core.SyncStatus, error) {

	services := m.serviceFactory.FromDB()

	syncState, err := services.Core.GetSyncState(ctx, core.GetSyncStateParams{
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
	})
	if err != nil {
		// No sync state exists, proceed with sync
		return false, "", nil
	}

	if syncState.SyncStatus == core.SyncStatusInProgress {
		return true, core.SyncStatusInProgress, nil
	}

	// Check if recently synced (within 30 minutes)
	if syncState.SyncStatus == core.SyncStatusCompleted &&
		syncState.LastSyncedAt.Valid &&
		time.Since(syncState.LastSyncedAt.Time) < 30*time.Minute { // 30 minutes
		return true, core.SyncStatusCompleted, nil
	}

	return false, "", nil
}
