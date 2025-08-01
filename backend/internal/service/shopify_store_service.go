package service

import (
	"context"
	"log"

	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

// ShopifyStoreRepository defines the interface for shopify store operations
type ShopifyStoreRepository interface {
	GetShopifyStoreByDomain(ctx context.Context, shopDomain string) (shopify.ShopifyStore, error)
	GetShopifyStoreByID(ctx context.Context, id id.ID[id.ShopifyStore]) (shopify.ShopifyStore, error)
	CreateShopifyStore(ctx context.Context, arg shopify.CreateShopifyStoreParams) (shopify.ShopifyStore, error)
	UpdateShopifyStore(ctx context.Context, arg shopify.UpdateShopifyStoreParams) (shopify.ShopifyStore, error)
	CreateOrUpdateShopifyStore(ctx context.Context, arg shopify.CreateOrUpdateShopifyStoreParams) (shopify.ShopifyStore, error)
}

// ShopifyStoreService provides business logic for shopify store operations
type ShopifyStoreService struct {
	storeRepo ShopifyStoreRepository
}

// NewShopifyStoreService creates a new ShopifyStoreService instance
func NewShopifyStoreService(storeRepo ShopifyStoreRepository) *ShopifyStoreService {
	return &ShopifyStoreService{
		storeRepo: storeRepo,
	}
}

// ShopifyStore represents a shopify store in the service layer
type ShopifyStore struct {
	ID         id.ID[id.ShopifyStore] `json:"id"`
	ShopDomain string                 `json:"shop_domain"`
	ShopName   *string                `json:"shop_name,omitempty"`
	Timezone   *string                `json:"timezone,omitempty"`
	Currency   *string                `json:"currency,omitempty"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

// convertDomainShopifyStore converts a repository shopify store to a service shopify store
func (s *ShopifyStoreService) convertDomainShopifyStore(domainStore shopify.ShopifyStore) *ShopifyStore {
	store := &ShopifyStore{
		ID:         domainStore.ID,
		ShopDomain: domainStore.ShopDomain,
		CreatedAt:  domainStore.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  domainStore.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if domainStore.ShopName.Valid {
		store.ShopName = &domainStore.ShopName.String
	}
	if domainStore.Timezone.Valid {
		store.Timezone = &domainStore.Timezone.String
	}
	if domainStore.Currency.Valid {
		store.Currency = &domainStore.Currency.String
	}

	return store
}

// GetStoreByDomain retrieves a shopify store by domain
func (s *ShopifyStoreService) GetStoreByDomain(ctx context.Context, domain string) (*ShopifyStore, error) {
	domainStore, err := s.storeRepo.GetShopifyStoreByDomain(ctx, domain)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed to get shopify store by domain %s: %v", domain, err)
		return nil, errors.Wrapf(err, "failed to get shopify store")
	}

	return s.convertDomainShopifyStore(domainStore), nil
}

// GetStoreByID retrieves a shopify store by ID
func (s *ShopifyStoreService) GetStoreByID(ctx context.Context, storeID id.ID[id.ShopifyStore]) (*ShopifyStore, error) {
	domainStore, err := s.storeRepo.GetShopifyStoreByID(ctx, storeID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		log.Printf("Failed to get shopify store by ID %s: %v", storeID, err)
		return nil, errors.Wrapf(err, "failed to get shopify store")
	}

	return s.convertDomainShopifyStore(domainStore), nil
}

// CreateOrUpdateStore creates or updates a shopify store
func (s *ShopifyStoreService) CreateOrUpdateStore(ctx context.Context, shopDomain string, shopName, timezone, currency *string) (*ShopifyStore, error) {
	params := shopify.CreateOrUpdateShopifyStoreParams{
		ID:         id.NewGeneration[id.ShopifyStore](),
		ShopDomain: shopDomain,
	}

	if shopName != nil {
		params.ShopName.Valid = true
		params.ShopName.String = *shopName
	}
	if timezone != nil {
		params.Timezone.Valid = true
		params.Timezone.String = *timezone
	}
	if currency != nil {
		params.Currency.Valid = true
		params.Currency.String = *currency
	}

	domainStore, err := s.storeRepo.CreateOrUpdateShopifyStore(ctx, params)
	if err != nil {
		log.Printf("Failed to create or update shopify store for domain %s: %v", shopDomain, err)
		return nil, errors.Wrapf(err, "failed to create or update shopify store")
	}

	return s.convertDomainShopifyStore(domainStore), nil
}
