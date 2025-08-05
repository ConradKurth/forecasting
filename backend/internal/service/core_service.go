package service

import (
	"context"

	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
)

// CoreService handles core platform operations
type CoreService struct {
	coreRepo core.Querier
}

// NewCoreService creates a new CoreService instance
func NewCoreService(coreRepo core.Querier) *CoreService {
	return &CoreService{
		coreRepo: coreRepo,
	}
}

// GetPlatformIntegrationByShopAndType retrieves a platform integration by shop and type
func (s *CoreService) GetPlatformIntegrationByShopAndType(ctx context.Context, arg core.GetPlatformIntegrationByShopAndTypeParams) (core.PlatformIntegration, error) {
	return s.coreRepo.GetPlatformIntegrationByShopAndType(ctx, arg)
}

// CreatePlatformIntegration creates a new platform integration
func (s *CoreService) CreatePlatformIntegration(ctx context.Context, arg core.CreatePlatformIntegrationParams) (core.PlatformIntegration, error) {
	return s.coreRepo.CreatePlatformIntegration(ctx, arg)
}

// GetSyncState retrieves a sync state by integration and entity type
func (s *CoreService) GetSyncState(ctx context.Context, arg core.GetSyncStateParams) (core.SyncState, error) {
	return s.coreRepo.GetSyncState(ctx, arg)
}

// GetSyncStatesByIntegrationID retrieves all sync states for an integration
func (s *CoreService) GetSyncStatesByIntegrationID(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) ([]core.SyncState, error) {
	return s.coreRepo.GetSyncStatesByIntegrationID(ctx, integrationID)
}
