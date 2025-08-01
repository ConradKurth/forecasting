package service

import (
	"context"
	"database/sql"
	"log"

	"github.com/ConradKurth/forecasting/backend/internal/crypto"
	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/pkg/errors"
)

// UserRepository defines the interface that the UserService needs from the repository layer.
// This follows the dependency inversion principle - the service defines what it needs
// rather than depending on the full repository interface.
type UserRepository interface {
	GetUserByShopDomain(ctx context.Context, shopDomain string) (users.User, error)
	CreateOrUpdateUser(ctx context.Context, arg users.CreateOrUpdateUserParams) (users.User, error)
}

// UserService provides business logic for user operations
type UserService struct {
	userRepo UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// User represents a user in the service layer
type User struct {
	ID         id.ID[id.User] `json:"id"`
	ShopDomain string         `json:"shop_domain"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

// convertDomainUser converts a repository user to a service user
func (s *UserService) convertDomainUser(domainUser users.User) *User {
	return &User{
		ID:         domainUser.ID,
		ShopDomain: domainUser.ShopDomain,
		CreatedAt:  domainUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  domainUser.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetUser retrieves a user by shop domain
func (s *UserService) GetUser(ctx context.Context, shopDomain string) (*User, error) {
	domainUser, err := s.userRepo.GetUserByShopDomain(ctx, shopDomain)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed to get user by shop domain %s: %v", shopDomain, err)
		return nil, errors.Wrapf(err, "failed to get user")
	}

	return s.convertDomainUser(domainUser), nil
}

// CreateOrUpdateUser creates or updates a user with their Shopify access token
func (s *UserService) CreateOrUpdateUser(ctx context.Context, shopDomain, accessToken string) (*User, error) {
	domainUser, err := s.userRepo.CreateOrUpdateUser(ctx, users.CreateOrUpdateUserParams{
		ID:          id.NewGeneration[id.User](),
		ShopDomain:  shopDomain,
		AccessToken: crypto.EncryptedSecret(accessToken),
	})
	if err != nil {
		log.Printf("Failed to create or update user for shop %s: %v", shopDomain, err)
		return nil, errors.Wrapf(err, "failed to create or update user")
	}

	return s.convertDomainUser(domainUser), nil
}

// GetShopifyAccessToken retrieves and decrypts the Shopify access token for a shop
// This method handles the complete decryption process and should be used for making
// authenticated Shopify API calls
func (s *UserService) GetShopifyAccessToken(ctx context.Context, shopDomain string) (string, error) {
	domainUser, err := s.userRepo.GetUserByShopDomain(ctx, shopDomain)
	if err != nil {
		log.Printf("Failed to get user by shop domain %s: %v", shopDomain, err)
		return "", errors.Wrapf(err, "failed to get user for token decryption")
	}

	// The access token is automatically decrypted when we call String() on the EncryptedSecret
	accessToken := domainUser.AccessToken.String()
	if accessToken == "" {
		return "", errors.Errorf("no access token found for shop domain: %s", shopDomain)
	}

	return accessToken, nil
}

// ValidateUser checks if a user exists for the given shop domain
func (s *UserService) ValidateUser(ctx context.Context, shopDomain string) (bool, error) {
	_, err := s.userRepo.GetUserByShopDomain(ctx, shopDomain)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
