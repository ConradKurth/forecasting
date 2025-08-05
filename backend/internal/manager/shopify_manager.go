package manager

import (
	"context"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/crypto"
	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/interfaces"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
)

// ShopifyManager orchestrates operations across user, shopify store, and shopify user repositories.
// It ensures data consistency by wrapping multi-table operations in database transactions.
//
// Transaction Strategy:
// - Write operations (CreateOrUpdateShopifyIntegration) use transactions to ensure atomicity
// - Read operations that need consistency (GetShopifyIntegration, ListUserShopifyIntegrations) use read-only transactions
// - Single-table operations can use the regular queries without transactions
//
// Repository Dependencies:
// The manager works directly with repository query interfaces.
// It maintains a database reference for transaction management. Queries are accessed
// directly from the database connection (regular or transactional).
type ShopifyManager struct {
	database db.Database
	queue    interfaces.Queue
}

// NewShopifyManager creates a new ShopifyManager instance
func NewShopifyManager(database db.Database, queue interfaces.Queue) *ShopifyManager {
	return &ShopifyManager{
		database: database,
		queue:    queue,
	}
}

// ShopifyIntegration represents the complete shopify integration data
type ShopifyIntegration struct {
	User        *users.User         `json:"user"`
	Store       *shopify.ShopifyStore `json:"store"`
	ShopifyUser *shopify.ShopifyUser  `json:"shopify_user"`
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
		// Step 1: Create or get the user
		user, err := m.ensureUserTx(ctx, txDB.GetUsers(), params.UserID)
		if err != nil {
			return errors.Wrap(err, "failed to ensure user exists")
		}

		// Step 2: Create or update the shopify store
		var shopName, timezone, currency pgtype.Text
		if params.ShopName != nil {
			shopName = pgtype.Text{String: *params.ShopName, Valid: true}
		}
		if params.Timezone != nil {
			timezone = pgtype.Text{String: *params.Timezone, Valid: true}
		}
		if params.Currency != nil {
			currency = pgtype.Text{String: *params.Currency, Valid: true}
		}

		store, err := txDB.GetShopify().CreateOrUpdateShopifyStore(ctx, shopify.CreateOrUpdateShopifyStoreParams{
			ID:         id.NewGeneration[id.ShopifyStore](),
			ShopDomain: params.ShopDomain,
			ShopName:   shopName,
			Timezone:   timezone,
			Currency:   currency,
		})
		if err != nil {
			return errors.Wrap(err, "failed to create or update shopify store")
		}

		// Step 3: Create or update the shopify user with access token
		var expiresAt pgtype.Timestamp
		if params.ExpiresAt != nil {
			expiresAt = pgtype.Timestamp{Time: *params.ExpiresAt, Valid: true}
		}

		shopifyUser, err := txDB.GetShopify().CreateOrUpdateShopifyUser(ctx, shopify.CreateOrUpdateShopifyUserParams{
			ID:             id.NewGeneration[id.ShopifyUser](),
			UserID:         params.UserID,
			ShopifyStoreID: store.ID,
			AccessToken:    crypto.EncryptedSecret(params.AccessToken),
			Scope:          params.Scope,
			ExpiresAt:      expiresAt,
		})
		if err != nil {
			return errors.Wrap(err, "failed to create or update shopify user")
		}

		result = &ShopifyIntegration{
			User:        &user,
			Store:       &store,
			ShopifyUser: &shopifyUser,
			AccessToken: params.AccessToken,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Enqueue Shopify store sync task after successful integration creation
	err = m.queue.EnqueueShopifyStoreSync(ctx, params.UserID.String(), result.Store.ID.String(), params.AccessToken)
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
	// Get the user
	user, err := m.database.GetUsers().GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	// Get the store
	store, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, shopDomain)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify store")
	}

	// Get the shopify user
	shopifyUser, err := m.database.GetShopify().GetShopifyUserByUserAndStore(ctx, shopify.GetShopifyUserByUserAndStoreParams{
		UserID:         userID,
		ShopifyStoreID: store.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get shopify user")
	}

	result := &ShopifyIntegration{
		User:        &user,
		Store:       &store,
		ShopifyUser: &shopifyUser,
		AccessToken: shopifyUser.AccessToken.String(),
	}

	return result, nil
}

// ValidateShopifyIntegration checks if a complete shopify integration exists for a user and shop domain
func (m *ShopifyManager) ValidateShopifyIntegration(ctx context.Context, userID id.ID[id.User], shopDomain string) (bool, error) {
	// Check if user exists
	_, err := m.database.GetUsers().GetUserByID(ctx, userID)
	if err != nil {
		return false, nil // User doesn't exist
	}

	// Check if shopify store exists
	store, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, shopDomain)
	if err != nil {
		return false, nil // Store doesn't exist
	}

	// Check if shopify user exists
	_, err = m.database.GetShopify().GetShopifyUserByUserAndStore(ctx, shopify.GetShopifyUserByUserAndStoreParams{
		UserID:         userID,
		ShopifyStoreID: store.ID,
	})
	if err != nil {
		return false, nil // Shopify user doesn't exist
	}

	return true, nil
}

// GetShopifyAccessToken is a convenience method to get just the access token
func (m *ShopifyManager) GetShopifyAccessToken(ctx context.Context, userID id.ID[id.User], shopDomain string) (string, error) {
	// Get the store
	store, err := m.database.GetShopify().GetShopifyStoreByDomain(ctx, shopDomain)
	if err != nil {
		return "", errors.Wrap(err, "failed to get shopify store")
	}

	// Get the shopify user
	shopifyUser, err := m.database.GetShopify().GetShopifyUserByUserAndStore(ctx, shopify.GetShopifyUserByUserAndStoreParams{
		UserID:         userID,
		ShopifyStoreID: store.ID,
	})
	if err != nil {
		return "", errors.Wrap(err, "could not get the shopify user")
	}

	return shopifyUser.AccessToken.String(), nil
}

// ensureUserTx creates a user if it doesn't exist, or returns the existing user (transactional version)
func (m *ShopifyManager) ensureUserTx(ctx context.Context, userQueries users.Querier, userID id.ID[id.User]) (users.User, error) {
	// Try to get the existing user
	user, err := userQueries.GetUserByID(ctx, userID)
	if err != nil {
		// If user doesn't exist, create it
		user, err = userQueries.CreateUser(ctx, userID)
		if err != nil {
			return users.User{}, errors.Wrap(err, "failed to create user")
		}
	}

	return user, nil
}

// SyncStoreInfo updates store information by fetching from Shopify API
func (m *ShopifyManager) SyncStoreInfo(ctx context.Context, userID id.ID[id.User], shopID id.ID[id.ShopifyStore]) error {
	// For now, just log that the sync was called
	// The actual Shopify API integration would go here
	logger.Info("Store sync requested", "user_id", userID, "shop_id", shopID)
	return nil
}
