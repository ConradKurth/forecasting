package service

import (
	"context"
	"log"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/crypto"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

// ShopifyUserRepository defines the interface for shopify user operations
type ShopifyUserRepository interface {
	GetShopifyUserByUserAndDomain(ctx context.Context, arg shopify.GetShopifyUserByUserAndDomainParams) (shopify.ShopifyUser, error)
	CreateOrUpdateShopifyUser(ctx context.Context, arg shopify.CreateOrUpdateShopifyUserParams) (shopify.ShopifyUser, error)
	GetShopifyUsersByUser(ctx context.Context, userID id.ID[id.User]) ([]shopify.ShopifyUser, error)
}

// ShopifyUserService provides business logic for shopify user operations
type ShopifyUserService struct {
	shopifyUserRepo ShopifyUserRepository
}

// NewShopifyUserService creates a new ShopifyUserService instance
func NewShopifyUserService(shopifyUserRepo ShopifyUserRepository) *ShopifyUserService {
	return &ShopifyUserService{
		shopifyUserRepo: shopifyUserRepo,
	}
}

// ShopifyUser represents a shopify user in the service layer
type ShopifyUser struct {
	ID             id.ID[id.ShopifyUser]  `json:"id"`
	UserID         id.ID[id.User]         `json:"user_id"`
	ShopifyStoreID id.ID[id.ShopifyStore] `json:"shopify_store_id"`
	Scope          string                 `json:"scope"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

// convertDomainShopifyUser converts a repository shopify user to a service shopify user
func (s *ShopifyUserService) convertDomainShopifyUser(domainUser shopify.ShopifyUser) *ShopifyUser {
	user := &ShopifyUser{
		ID:             domainUser.ID,
		UserID:         domainUser.UserID,
		ShopifyStoreID: domainUser.ShopifyStoreID,
		Scope:          domainUser.Scope,
		CreatedAt:      domainUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      domainUser.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if domainUser.ExpiresAt.Valid {
		user.ExpiresAt = &domainUser.ExpiresAt.Time
	}

	return user
}

// GetShopifyUserByUserAndDomain retrieves a shopify user by user ID and shop domain
func (s *ShopifyUserService) GetShopifyUserByUserAndDomain(ctx context.Context, userID id.ID[id.User], shopDomain string) (*ShopifyUser, error) {
	domainUser, err := s.shopifyUserRepo.GetShopifyUserByUserAndDomain(ctx, shopify.GetShopifyUserByUserAndDomainParams{
		UserID:     userID,
		ShopDomain: shopDomain,
	})
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed to get shopify user for user %s and domain %s: %v", userID, shopDomain, err)
		return nil, errors.Wrapf(err, "failed to get shopify user")
	}

	return s.convertDomainShopifyUser(domainUser), nil
}

// CreateOrUpdateShopifyUser creates or updates a shopify user with access token
func (s *ShopifyUserService) CreateOrUpdateShopifyUser(ctx context.Context, userID id.ID[id.User], shopifyStoreID id.ID[id.ShopifyStore], accessToken, scope string, expiresAt *time.Time) (*ShopifyUser, error) {
	params := shopify.CreateOrUpdateShopifyUserParams{
		ID:             id.NewGeneration[id.ShopifyUser](),
		UserID:         userID,
		ShopifyStoreID: shopifyStoreID,
		AccessToken:    crypto.EncryptedSecret(accessToken),
		Scope:          scope,
	}

	if expiresAt != nil {
		params.ExpiresAt.Valid = true
		params.ExpiresAt.Time = *expiresAt
	}

	domainUser, err := s.shopifyUserRepo.CreateOrUpdateShopifyUser(ctx, params)
	if err != nil {
		log.Printf("Failed to create or update shopify user for user %s and store %s: %v", userID, shopifyStoreID, err)
		return nil, errors.Wrapf(err, "failed to create or update shopify user")
	}

	return s.convertDomainShopifyUser(domainUser), nil
}

// GetShopifyAccessToken retrieves and decrypts the Shopify access token for a user and shop domain
func (s *ShopifyUserService) GetShopifyAccessToken(ctx context.Context, userID id.ID[id.User], shopDomain string) (string, error) {
	domainUser, err := s.shopifyUserRepo.GetShopifyUserByUserAndDomain(ctx, shopify.GetShopifyUserByUserAndDomainParams{
		UserID:     userID,
		ShopDomain: shopDomain,
	})
	if err != nil {
		log.Printf("Failed to get shopify user for user %s and domain %s: %v", userID, shopDomain, err)
		return "", errors.Wrapf(err, "failed to get shopify user for token decryption")
	}

	// The access token is automatically decrypted when we call String() on the EncryptedSecret
	accessToken := domainUser.AccessToken.String()
	if accessToken == "" {
		return "", errors.Errorf("no access token found for user %s and shop domain: %s", userID, shopDomain)
	}

	return accessToken, nil
}

// GetShopifyUsersByUser retrieves all shopify users for a given user ID
func (s *ShopifyUserService) GetShopifyUsersByUser(ctx context.Context, userID id.ID[id.User]) ([]*ShopifyUser, error) {
	domainUsers, err := s.shopifyUserRepo.GetShopifyUsersByUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to get shopify users for user %s: %v", userID, err)
		return nil, errors.Wrapf(err, "failed to get shopify users")
	}

	serviceUsers := make([]*ShopifyUser, 0, len(domainUsers))
	for _, domainUser := range domainUsers {
		serviceUsers = append(serviceUsers, s.convertDomainShopifyUser(domainUser))
	}

	return serviceUsers, nil
}

// ValidateShopifyUser checks if a shopify user exists for the given user ID and shop domain
func (s *ShopifyUserService) ValidateShopifyUser(ctx context.Context, userID id.ID[id.User], shopDomain string) (bool, error) {
	_, err := s.shopifyUserRepo.GetShopifyUserByUserAndDomain(ctx, shopify.GetShopifyUserByUserAndDomainParams{
		UserID:     userID,
		ShopDomain: shopDomain,
	})
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
