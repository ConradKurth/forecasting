package manager

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/interfaces"
	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

// SyncStatus represents the status of a synchronization operation
type SyncStatus string

const (
	// Standard sync statuses
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"

	// Special result statuses
	SyncStatusSyncStarted     SyncStatus = "sync_started"
	SyncStatusNeverSynced     SyncStatus = "never_synced"
	SyncStatusPartialSyncOnly SyncStatus = "partial_sync_only"
)

// EntityType represents the type of entity being synchronized
type EntityType string

const (
	EntityTypeFullSync       EntityType = "full_sync"
	EntityTypeProduct        EntityType = "product"
	EntityTypeInventoryItem  EntityType = "inventory_item"
	EntityTypeInventoryLevel EntityType = "inventory_level"
	EntityTypeOrder          EntityType = "order"
	EntityTypeLocation       EntityType = "location"
)

// Type conversion helpers to keep conversions DRY and centralized

// ConvertSyncStatus provides bidirectional conversion between core and manager SyncStatus types
type ConvertSyncStatus struct{}

func (ConvertSyncStatus) FromCore(coreStatus core.SyncStatus) SyncStatus {
	switch coreStatus {
	case core.SyncStatusPending:
		return SyncStatusPending
	case core.SyncStatusInProgress:
		return SyncStatusInProgress
	case core.SyncStatusCompleted:
		return SyncStatusCompleted
	case core.SyncStatusFailed:
		return SyncStatusFailed
	default:
		return SyncStatusPending
	}
}

func (ConvertSyncStatus) ToCore(managerStatus SyncStatus) core.SyncStatus {
	switch managerStatus {
	case SyncStatusPending:
		return core.SyncStatusPending
	case SyncStatusInProgress:
		return core.SyncStatusInProgress
	case SyncStatusCompleted:
		return core.SyncStatusCompleted
	case SyncStatusFailed:
		return core.SyncStatusFailed
	default:
		return core.SyncStatusPending
	}
}

// ConvertEntityType provides bidirectional conversion between core and manager EntityType types
type ConvertEntityType struct{}

func (ConvertEntityType) FromCore(coreEntity core.EntityType) EntityType {
	switch coreEntity {
	case core.EntityTypeFullSync:
		return EntityTypeFullSync
	case core.EntityTypeProducts:
		return EntityTypeProduct
	case core.EntityTypeInventory:
		return EntityTypeInventoryItem
	case core.EntityTypeOrders:
		return EntityTypeOrder
	case core.EntityTypeLocations:
		return EntityTypeLocation
	default:
		return EntityTypeFullSync
	}
}

func (ConvertEntityType) ToCore(managerEntity EntityType) core.EntityType {
	switch managerEntity {
	case EntityTypeFullSync:
		return core.EntityTypeFullSync
	case EntityTypeProduct:
		return core.EntityTypeProducts
	case EntityTypeInventoryItem:
		return core.EntityTypeInventory
	case EntityTypeOrder:
		return core.EntityTypeOrders
	case EntityTypeLocation:
		return core.EntityTypeLocations
	default:
		return core.EntityTypeFullSync
	}
}

// Global converter instances for easy access
var (
	ConvertSync   = ConvertSyncStatus{}
	ConvertEntity = ConvertEntityType{}
)

// InventorySyncManager orchestrates inventory synchronization operations
// It ensures data consistency by wrapping operations in database transactions
// and coordinates between multiple repositories for complex sync workflows
type InventorySyncManager struct {
	database       db.Database
	shopifyManager *ShopifyManager
	queue          interfaces.Queue
}

// NewInventorySyncManager creates a new InventorySyncManager instance
func NewInventorySyncManager(database db.Database, shopifyManager *ShopifyManager, queue interfaces.Queue) *InventorySyncManager {
	return &InventorySyncManager{
		database:       database,
		shopifyManager: shopifyManager,
		queue:          queue,
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
	IntegrationID string     `json:"integration_id"`
	Status        SyncStatus `json:"status"`
	Error         string     `json:"error,omitempty"`
}

// TriggerShopifySync orchestrates a complete Shopify synchronization process
// This includes validating access, creating integrations, and triggering async sync
func (m *InventorySyncManager) TriggerShopifySync(ctx context.Context, req SyncRequest) (*SyncResult, error) {
	// Validate user exists
	_, err := m.database.GetUsers().GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	// Get shop and validate it exists
	shop, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, req.ShopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "shop not found")
	}

	// Get shopify user to get access token
	shopifyUser, err := m.database.GetShopify().GetShopifyUserByUserAndStore(ctx, shopify.GetShopifyUserByUserAndStoreParams{
		UserID:         req.UserID,
		ShopifyStoreID: shop.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get access token")
	}

	accessToken := shopifyUser.AccessToken.String()
	if accessToken == "" {
		return nil, errors.New("no access token found for user and shop")
	}

	// Create or get platform integration
	integration, err := m.getOrCreateShopifyIntegration(ctx, shop.ID, req.ShopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get/create integration")
	}

	// Check if sync should proceed
	if !req.Force {
		shouldSkip, skipReason, err := m.shouldSkipSync(ctx, integration.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to check sync status")
		}
		if shouldSkip {
			return &SyncResult{
				IntegrationID: integration.ID.String(),
				Status:        skipReason,
			}, nil
		}
	}

	// Enqueue async sync task
	err = m.queue.EnqueueShopifyInventorySync(ctx, integration.ID.String(), req.ShopDomain, accessToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to enqueue sync task")
	}

	return &SyncResult{
		IntegrationID: integration.ID.String(),
		Status:        SyncStatusSyncStarted,
	}, nil
}

// GetSyncStatus retrieves the current synchronization status for a shop
func (m *InventorySyncManager) GetSyncStatus(ctx context.Context, userID id.ID[id.User], shopDomain string) (*SyncResult, error) {
	// Get shop
	shop, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "shop not found")
	}

	// Get integration
	integration, err := m.database.GetCore().GetPlatformIntegrationByShopAndType(ctx, core.GetPlatformIntegrationByShopAndTypeParams{
		ShopID:       shop.ID,
		PlatformType: core.PlatformTypeShopify,
	})
	if err != nil {
		return nil, errors.Wrap(err, "integration not found")
	}

	// Get sync states
	syncStates, err := m.database.GetCore().GetSyncStatesByIntegrationID(ctx, integration.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sync states")
	}

	result := &SyncResult{
		IntegrationID: integration.ID.String(),
	}

	if len(syncStates) == 0 {
		result.Status = SyncStatusNeverSynced
	} else {
		// Find the most recent full_sync
		for _, state := range syncStates {
			if state.EntityType == core.EntityTypeFullSync {
				result.Status = ConvertSync.FromCore(state.SyncStatus)
				if state.ErrorMessage.Valid {
					result.Error = state.ErrorMessage.String
				}
				break
			}
		}
		if result.Status == "" {
			result.Status = SyncStatusPartialSyncOnly
		}
	}

	return result, nil
}

// getOrCreateShopifyIntegration gets or creates a Shopify platform integration
func (m *InventorySyncManager) getOrCreateShopifyIntegration(ctx context.Context, shopID id.ID[id.ShopifyStore], shopDomain string) (core.PlatformIntegration, error) {
	// Try to get existing integration
	integration, err := m.database.GetCore().GetPlatformIntegrationByShopAndType(ctx, core.GetPlatformIntegrationByShopAndTypeParams{
		ShopID:       shopID,
		PlatformType: core.PlatformTypeShopify,
	})
	if err == nil {
		return integration, nil
	}

	// Create new integration
	integration, err = m.database.GetCore().CreatePlatformIntegration(ctx, core.CreatePlatformIntegrationParams{
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
func (m *InventorySyncManager) shouldSkipSync(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) (bool, SyncStatus, error) {
	syncState, err := m.database.GetCore().GetSyncState(ctx, core.GetSyncStateParams{
		IntegrationID: integrationID,
		EntityType:    core.EntityTypeFullSync,
	})
	if err != nil {
		// No sync state exists, proceed with sync
		return false, "", nil
	}

	if syncState.SyncStatus == core.SyncStatusInProgress {
		return true, SyncStatusInProgress, nil
	}

	// Check if recently synced (within 30 minutes)
	if syncState.SyncStatus == core.SyncStatusCompleted &&
		syncState.LastSyncedAt.Valid &&
		time.Since(syncState.LastSyncedAt.Time) < 30*time.Minute { // 30 minutes
		return true, SyncStatusCompleted, nil
	}

	return false, "", nil
}

// SyncInventory performs inventory synchronization for a platform integration
func (m *InventorySyncManager) SyncInventory(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) error {
	// For now, just log that the sync was called
	// The actual inventory sync logic would go here
	logger.Info("Inventory sync requested", "integration_id", integrationID)
	return nil
}
