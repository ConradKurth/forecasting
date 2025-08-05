package factory

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
)

// Service interfaces for dependency injection
// These interfaces define the contracts that services must implement

// UserServiceInterface defines the contract for user operations
type UserServiceInterface interface {
	GetUser(ctx context.Context, userID id.ID[id.User]) (*service.User, error)
	CreateUser(ctx context.Context, userID id.ID[id.User]) (*service.User, error)
	ValidateUser(ctx context.Context, userID id.ID[id.User]) (bool, error)
}

// ShopifyStoreServiceInterface defines the contract for shopify store operations
type ShopifyStoreServiceInterface interface {
	CreateOrUpdateStore(ctx context.Context, shopDomain string, shopName, timezone, currency *string) (*service.ShopifyStore, error)
	GetStoreByDomain(ctx context.Context, domain string) (*service.ShopifyStore, error)
	GetStoreByID(ctx context.Context, storeID id.ID[id.ShopifyStore]) (*service.ShopifyStore, error)
}

// ShopifyUserServiceInterface defines the contract for shopify user operations
type ShopifyUserServiceInterface interface {
	CreateOrUpdateShopifyUser(ctx context.Context, userID id.ID[id.User], storeID id.ID[id.ShopifyStore], accessToken, scope string, expiresAt *time.Time) (*service.ShopifyUser, error)
	GetShopifyUserByUserAndDomain(ctx context.Context, userID id.ID[id.User], shopDomain string) (*service.ShopifyUser, error)
	GetShopifyUsersByUser(ctx context.Context, userID id.ID[id.User]) ([]*service.ShopifyUser, error)
	GetShopifyAccessToken(ctx context.Context, userID id.ID[id.User], shopDomain string) (string, error)
	ValidateShopifyUser(ctx context.Context, userID id.ID[id.User], shopDomain string) (bool, error)
}

// CoreServiceInterface defines the contract for core platform operations
type CoreServiceInterface interface {
	GetPlatformIntegrationByShopAndType(ctx context.Context, arg core.GetPlatformIntegrationByShopAndTypeParams) (core.PlatformIntegration, error)
	CreatePlatformIntegration(ctx context.Context, arg core.CreatePlatformIntegrationParams) (core.PlatformIntegration, error)
	GetSyncState(ctx context.Context, arg core.GetSyncStateParams) (core.SyncState, error)
	GetSyncStatesByIntegrationID(ctx context.Context, integrationID id.ID[id.PlatformIntegration]) ([]core.SyncState, error)
}

// ServiceInterfaces holds all service interfaces for dependency injection
type ServiceInterfaces struct {
	User         UserServiceInterface
	ShopifyStore ShopifyStoreServiceInterface
	ShopifyUser  ShopifyUserServiceInterface
	Core         CoreServiceInterface
}
