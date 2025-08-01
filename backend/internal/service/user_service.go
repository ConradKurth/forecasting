package service

import (
	"context"
	"database/sql"
	"log"

	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/pkg/errors"
)

// UserRepository defines the interface that the UserService needs from the repository layer.
// This follows the dependency inversion principle - the service defines what it needs
// rather than depending on the full repository interface.
type UserRepository interface {
	GetUserByID(ctx context.Context, argID id.ID[id.User]) (users.User, error)
	CreateUser(ctx context.Context, argID id.ID[id.User]) (users.User, error)
	UpdateUser(ctx context.Context, argID id.ID[id.User]) (users.User, error)
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
	ID        id.ID[id.User] `json:"id"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

// convertDomainUser converts a repository user to a service user
func (s *UserService) convertDomainUser(domainUser users.User) *User {
	return &User{
		ID:        domainUser.ID,
		CreatedAt: domainUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: domainUser.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID id.ID[id.User]) (*User, error) {
	domainUser, err := s.userRepo.GetUserByID(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed to get user by ID %s: %v", userID, err)
		return nil, errors.Wrapf(err, "failed to get user")
	}

	return s.convertDomainUser(domainUser), nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, userID id.ID[id.User]) (*User, error) {
	domainUser, err := s.userRepo.CreateUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to create user %s: %v", userID, err)
		return nil, errors.Wrapf(err, "failed to create user")
	}

	return s.convertDomainUser(domainUser), nil
}

// UpdateUser updates a user's updated_at timestamp
func (s *UserService) UpdateUser(ctx context.Context, userID id.ID[id.User]) (*User, error) {
	domainUser, err := s.userRepo.UpdateUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to update user %s: %v", userID, err)
		return nil, errors.Wrapf(err, "failed to update user")
	}

	return s.convertDomainUser(domainUser), nil
}

// ValidateUser checks if a user exists for the given user ID
func (s *UserService) ValidateUser(ctx context.Context, userID id.ID[id.User]) (bool, error) {
	_, err := s.userRepo.GetUserByID(ctx, userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
