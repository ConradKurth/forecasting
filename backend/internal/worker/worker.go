package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/internal/shopify"
	"github.com/ConradKurth/forecasting/backend/pkg/logger"
	"github.com/hibiken/asynq"
)

// Worker handles background job processing
type Worker struct {
	serviceFactory *factory.ServiceFactory
}

// New creates a new worker instance
func New(serviceFactory *factory.ServiceFactory) *Worker {
	return &Worker{
		serviceFactory: serviceFactory,
	}
}

// RegisterHandlers registers all task handlers with the given mux
func (w *Worker) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeShopifyStoreSync, w.HandleShopifyStoreSync)
}

// HandleShopifyStoreSync processes Shopify store synchronization tasks
func (w *Worker) HandleShopifyStoreSync(ctx context.Context, t *asynq.Task) error {
	var payload ShopifyStoreSyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal shopify store sync payload: %w", err)
	}

	// Create Shopify API client
	shopifyClient, err := shopify.NewClient(payload.ShopName, payload.Token)
	if err != nil {
		return fmt.Errorf("failed to create Shopify client: %w", err)
	}

	// Fetch shop information from Shopify API
	logger.Info("Fetching shop info from Shopify API", "shop", payload.ShopName, "user_id", payload.UserID)
	shopInfo, err := shopifyClient.GetShopInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch shop info from Shopify API: %w", err)
	}

	logger.Info("Successfully fetched shop info", 
		"shop", payload.ShopName, 
		"name", shopInfo.Name, 
		"currency", shopInfo.Currency, 
		"timezone", shopInfo.Timezone)

	// Get Shopify store service
	shopifyStoreService := w.serviceFactory.CreateShopifyStoreService()

	// Update store with fetched information
	_, err = shopifyStoreService.CreateOrUpdateStore(
		ctx,
		payload.ShopName,
		&shopInfo.Name,
		&shopInfo.Timezone,
		&shopInfo.Currency,
	)
	if err != nil {
		return fmt.Errorf("failed to update store in database: %w", err)
	}

	logger.Info("Successfully synced Shopify store data", 
		"shop", payload.ShopName, 
		"user_id", payload.UserID,
		"name", shopInfo.Name, 
		"currency", shopInfo.Currency, 
		"timezone", shopInfo.Timezone)

	return nil
}
