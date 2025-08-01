package manager

import (
	"context"
	"log"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/pkg/errors"
)

// ShopifyManager orchestrates operations across user, shopify store, and shopify user services
type ShopifyManager struct {
	userService         *service.UserService
	shopifyStoreService *service.ShopifyStoreService
	shopifyUserService  *service.ShopifyUserService
}

// NewShopifyManager creates a new ShopifyManager instance
func NewShopifyManager(
	userService *service.UserService,
	shopifyStoreService *service.ShopifyStoreService,
	shopifyUserService *service.ShopifyUserService,
) *ShopifyManager {
	return &ShopifyManager{
		userService:         userService,
		shopifyStoreService: shopifyStoreService,
		shopifyUserService:  shopifyUserService,
	}
}

// ShopifyIntegration represents the complete shopify integration data
type ShopifyIntegration struct {
	User         *service.User         `json:"user"`
	Store        *service.ShopifyStore `json:"store"`
	ShopifyUser  *service.ShopifyUser  `json:"shopify_user"`
	AccessToken  string                `json:"-"` // Don't serialize access token
}

// CreateOrUpdateShopifyIntegration handles the complete process of setting up a Shopify integration
// This includes creating/updating the user, store, and shopify user with access token
func (m *ShopifyManager) CreateOrUpdateShopifyIntegration(ctx context.Context, params CreateShopifyIntegrationParams) (*ShopifyIntegration, error) {
	// Step 1: Create or get the user
	user, err := m.ensureUser(ctx, params.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure user exists")
	}

	// Step 2: Create or update the shopify store
	store, err := m.shopifyStoreService.CreateOrUpdateStore(
		ctx,
		params.ShopDomain,
		params.ShopName,
		params.Timezone,
		params.Currency,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create or update shopify store")
	}

	// Step 3: Create or update the shopify user with access token
	shopifyUser, err := m.shopifyUserService.CreateOrUpdateShopifyUser(
		ctx,
		user.ID,
		store.ID,
		params.AccessToken,
		params.Scope,
		params.ExpiresAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create or update shopify user")
	}

	return &ShopifyIntegration{
		User:        user,
		Store:       store,
		ShopifyUser: shopifyUser,
		AccessToken: params.AccessToken,
	}, nil
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
func (m *ShopifyManager) GetShopifyIntegration(ctx context.Context, userID id.ID[id.User], shopDomain string) (*ShopifyIntegration, error) {
	// Get the user
	user, err := m.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get the store
	store, err := m.shopifyStoreService.GetStoreByDomain(ctx, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify store")
	}
	if store == nil {
		return nil, errors.New("shopify store not found")
	}

	// Get the shopify user
	shopifyUser, err := m.shopifyUserService.GetShopifyUserByUserAndDomain(ctx, userID, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify user")
	}
	if shopifyUser == nil {
		return nil, errors.New("shopify user not found")
	}

	// Get the access token
	accessToken, err := m.shopifyUserService.GetShopifyAccessToken(ctx, userID, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify access token")
	}

	return &ShopifyIntegration{
		User:        user,
		Store:       store,
		ShopifyUser: shopifyUser,
		AccessToken: accessToken,
	}, nil
}

// ValidateShopifyIntegration checks if a complete shopify integration exists for a user and shop domain
func (m *ShopifyManager) ValidateShopifyIntegration(ctx context.Context, userID id.ID[id.User], shopDomain string) (bool, error) {
	// Check if user exists
	userExists, err := m.userService.ValidateUser(ctx, userID)
	if err != nil || !userExists {
		return false, err
	}

	// Check if shopify user exists
	shopifyUserExists, err := m.shopifyUserService.ValidateShopifyUser(ctx, userID, shopDomain)
	if err != nil || !shopifyUserExists {
		return false, err
	}

	return true, nil
}

// GetShopifyAccessToken is a convenience method to get just the access token
func (m *ShopifyManager) GetShopifyAccessToken(ctx context.Context, userID id.ID[id.User], shopDomain string) (string, error) {
	return m.shopifyUserService.GetShopifyAccessToken(ctx, userID, shopDomain)
}

// ListUserShopifyIntegrations retrieves all shopify integrations for a user
func (m *ShopifyManager) ListUserShopifyIntegrations(ctx context.Context, userID id.ID[id.User]) ([]*ShopifyIntegration, error) {
	// Get the user
	user, err := m.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get all shopify users for this user
	shopifyUsers, err := m.shopifyUserService.GetShopifyUsersByUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify users")
	}

	// For each shopify user, get the corresponding store and build the integration
	integrations := make([]*ShopifyIntegration, 0, len(shopifyUsers))
	for _, shopifyUser := range shopifyUsers {
		// Get the store by ID
		store, err := m.shopifyStoreService.GetStoreByID(ctx, shopifyUser.ShopifyStoreID)
		if err != nil {
			log.Printf("Failed to get store for shopify user %s with store ID %s: %v", shopifyUser.ID, shopifyUser.ShopifyStoreID, err)
			continue // Skip this integration if we can't get the store
		}
		
		integrations = append(integrations, &ShopifyIntegration{
			User:        user,
			Store:       store,
			ShopifyUser: shopifyUser,
			AccessToken: "", // Don't return access token in list operations
		})
	}

	return integrations, nil
}

// ensureUser creates a user if it doesn't exist, or returns the existing user
func (m *ShopifyManager) ensureUser(ctx context.Context, userID id.ID[id.User]) (*service.User, error) {
	// Try to get the existing user
	user, err := m.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if user exists")
	}

	// If user doesn't exist, create it
	if user == nil {
		user, err = m.userService.CreateUser(ctx, userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create user")
		}
		log.Printf("Created new user: %s", userID)
	}

	return user, nil
}
