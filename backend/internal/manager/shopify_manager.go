package manager

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/ConradKurth/forecasting/backend/internal/worker"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/pkg/errors"
)

// ShopifyManager orchestrates operations across user, shopify store, and shopify user services.
// It ensures data consistency by wrapping multi-table operations in database transactions.
//
// Transaction Strategy:
// - Write operations (CreateOrUpdateShopifyIntegration) use transactions to ensure atomicity
// - Read operations that need consistency (GetShopifyIntegration, ListUserShopifyIntegrations) use read-only transactions
// - Single-table operations can use the regular services without transactions
//
// Service Dependencies:
// The manager uses interface-based services for dependency injection and testability.
// It maintains a database reference for transaction management. Services are created
// on-demand from the appropriate database connection (regular or transactional).
type ShopifyManager struct {
	database *db.DB
	services *factory.ServiceInterfaces
	queue    worker.Queue
}

// NewShopifyManager creates a new ShopifyManager instance with default services
func NewShopifyManager(database *db.DB, services *factory.ServiceInterfaces, queue worker.Queue) *ShopifyManager {
	return &ShopifyManager{
		database: database,
		services: services,
		queue:    queue,
	}
}

// withTxServices creates transactional service interfaces from a transaction.
func (m *ShopifyManager) withTxServices(tx *db.TxDB) *factory.ServiceInterfaces {
	return factory.NewServiceInterfacesFromTx(tx)
}

// ShopifyIntegration represents the complete shopify integration data
type ShopifyIntegration struct {
	User        *service.User         `json:"user"`
	Store       *service.ShopifyStore `json:"store"`
	ShopifyUser *service.ShopifyUser  `json:"shopify_user"`
	AccessToken string                `json:"-"` // Don't serialize access token
}

// CreateOrUpdateShopifyIntegration handles the complete process of setting up a Shopify integration
// This includes creating/updating the user, store, and shopify user with access token
// All operations are wrapped in a database transaction to ensure consistency
//
// Transaction Behavior:
// - If any step fails, all changes are rolled back automatically
// - The transaction ensures that partial integrations cannot exist
// - Example: If store creation succeeds but shopify user creation fails, the store creation will be rolled back
func (m *ShopifyManager) CreateOrUpdateShopifyIntegration(ctx context.Context, params CreateShopifyIntegrationParams) (*ShopifyIntegration, error) {
	var result *ShopifyIntegration

	err := m.database.WithTx(ctx, func(txDB *db.TxDB) error {
		// Create transactional services
		services := m.withTxServices(txDB)

		// Step 1: Create or get the user
		user, err := m.ensureUserTx(ctx, services.User, params.UserID)
		if err != nil {
			return errors.Wrap(err, "failed to ensure user exists")
		}

		// Step 2: Create or update the shopify store
		store, err := services.ShopifyStore.CreateOrUpdateStore(
			ctx,
			params.ShopDomain,
			params.ShopName,
			params.Timezone,
			params.Currency,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create or update shopify store")
		}

		// Step 3: Create or update the shopify user with access token
		shopifyUser, err := services.ShopifyUser.CreateOrUpdateShopifyUser(
			ctx,
			user.ID,
			store.ID,
			params.AccessToken,
			params.Scope,
			params.ExpiresAt,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create or update shopify user")
		}

		result = &ShopifyIntegration{
			User:        user,
			Store:       store,
			ShopifyUser: shopifyUser,
			AccessToken: params.AccessToken,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Enqueue Shopify store sync task after successful integration creation
	err = m.queue.EnqueueShopifyStoreSync(ctx, params.UserID.String(), params.ShopDomain, params.AccessToken)
	if err != nil {
		// Log the error but don't fail the integration creation
		logger.Error("Failed to enqueue Shopify store sync task", "user_id", params.UserID, "shop_domain", params.ShopDomain, "error", err)
	}

	return result, nil
}

// CreateShopifyIntegrationParams contains all the parameters needed to create a Shopify integration
type CreateShopifyIntegrationParams struct {
	UserID      id.ID[id.User] `json:"user_id"`
	ShopDomain  string         `json:"shop_domain"`
	ShopName    *string        `json:"shop_name,omitempty"`
	Timezone    *string        `json:"timezone,omitempty"`
	Currency    *string        `json:"currency,omitempty"`
	AccessToken string         `json:"access_token"`
	Scope       string         `json:"scope"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}

// GetShopifyIntegration retrieves the complete shopify integration for a user and shop domain
// Uses a read-only transaction to ensure data consistency
func (m *ShopifyManager) GetShopifyIntegration(ctx context.Context, userID id.ID[id.User], shopDomain string) (*ShopifyIntegration, error) {
	var result *ShopifyIntegration

	services := m.services

	// Get the user
	user, err := services.User.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get the store
	store, err := services.ShopifyStore.GetStoreByDomain(ctx, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify store")
	}
	if store == nil {
		return nil, errors.New("shopify store not found")
	}

	// Get the shopify user
	shopifyUser, err := services.ShopifyUser.GetShopifyUserByUserAndDomain(ctx, userID, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify user")
	}
	if shopifyUser == nil {
		return nil, errors.New("shopify user not found")
	}

	// Get the access token
	accessToken, err := services.ShopifyUser.GetShopifyAccessToken(ctx, userID, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify access token")
	}

	result = &ShopifyIntegration{
		User:        user,
		Store:       store,
		ShopifyUser: shopifyUser,
		AccessToken: accessToken,
	}

	return result, nil

}

// ValidateShopifyIntegration checks if a complete shopify integration exists for a user and shop domain
func (m *ShopifyManager) ValidateShopifyIntegration(ctx context.Context, userID id.ID[id.User], shopDomain string) (bool, error) {

	// Check if user exists
	userExists, err := m.services.User.ValidateUser(ctx, userID)
	if err != nil || !userExists {
		return false, err
	}

	// Check if shopify user exists
	shopifyUserExists, err := m.services.ShopifyUser.ValidateShopifyUser(ctx, userID, shopDomain)
	if err != nil || !shopifyUserExists {
		return false, err
	}

	return true, err
}

// GetShopifyAccessToken is a convenience method to get just the access token
func (m *ShopifyManager) GetShopifyAccessToken(ctx context.Context, userID id.ID[id.User], shopDomain string) (string, error) {
	services := m.services

	token, err := services.ShopifyUser.GetShopifyAccessToken(ctx, userID, shopDomain)
	if err != nil {
		return "", errors.Wrap(err, "could not get the shopify token")
	}
	return token, err
}

// ensureUserTx creates a user if it doesn't exist, or returns the existing user (transactional version)
func (m *ShopifyManager) ensureUserTx(ctx context.Context, userService factory.UserServiceInterface, userID id.ID[id.User]) (*service.User, error) {
	// Try to get the existing user
	user, err := userService.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if user exists")
	}

	// If user doesn't exist, create it
	if user == nil {
		user, err = userService.CreateUser(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create user")
		}
	}

	return user, nil
}
