package factory_test

import (
	"context"
	"testing"
	"time"

	"github.com/ConradKurth/forecasting/backend/internal/factory"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/ConradKurth/forecasting/backend/pkg/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations that implement the factory interfaces
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUser(ctx context.Context, userID id.ID[id.User]) (*service.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.User), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, userID id.ID[id.User]) (*service.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*service.User), args.Error(1)
}

func (m *MockUserService) ValidateUser(ctx context.Context, userID id.ID[id.User]) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type MockShopifyStoreService struct {
	mock.Mock
}

func (m *MockShopifyStoreService) CreateOrUpdateStore(ctx context.Context, shopDomain string, shopName, timezone, currency *string) (*service.ShopifyStore, error) {
	args := m.Called(ctx, shopDomain, shopName, timezone, currency)
	return args.Get(0).(*service.ShopifyStore), args.Error(1)
}

func (m *MockShopifyStoreService) GetStoreByDomain(ctx context.Context, domain string) (*service.ShopifyStore, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ShopifyStore), args.Error(1)
}

func (m *MockShopifyStoreService) GetStoreByID(ctx context.Context, storeID id.ID[id.ShopifyStore]) (*service.ShopifyStore, error) {
	args := m.Called(ctx, storeID)
	return args.Get(0).(*service.ShopifyStore), args.Error(1)
}

type MockShopifyUserService struct {
	mock.Mock
}

func (m *MockShopifyUserService) CreateOrUpdateShopifyUser(ctx context.Context, userID id.ID[id.User], storeID id.ID[id.ShopifyStore], accessToken, scope string, expiresAt *time.Time) (*service.ShopifyUser, error) {
	args := m.Called(ctx, userID, storeID, accessToken, scope, expiresAt)
	return args.Get(0).(*service.ShopifyUser), args.Error(1)
}

func (m *MockShopifyUserService) GetShopifyUserByUserAndDomain(ctx context.Context, userID id.ID[id.User], shopDomain string) (*service.ShopifyUser, error) {
	args := m.Called(ctx, userID, shopDomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ShopifyUser), args.Error(1)
}

func (m *MockShopifyUserService) GetShopifyUsersByUser(ctx context.Context, userID id.ID[id.User]) ([]*service.ShopifyUser, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*service.ShopifyUser), args.Error(1)
}

func (m *MockShopifyUserService) GetShopifyAccessToken(ctx context.Context, userID id.ID[id.User], shopDomain string) (string, error) {
	args := m.Called(ctx, userID, shopDomain)
	return args.String(0), args.Error(1)
}

func (m *MockShopifyUserService) ValidateShopifyUser(ctx context.Context, userID id.ID[id.User], shopDomain string) (bool, error) {
	args := m.Called(ctx, userID, shopDomain)
	return args.Bool(0), args.Error(1)
}

// Example: Component that accepts service interfaces for dependency injection
type UserManager struct {
	userService factory.UserServiceInterface
}

func NewUserManager(userService factory.UserServiceInterface) *UserManager {
	return &UserManager{userService: userService}
}

func (m *UserManager) GetUserInfo(ctx context.Context, userID id.ID[id.User]) (*service.User, error) {
	return m.userService.GetUser(ctx, userID)
}

// Test showing interface-based dependency injection
func TestInterfaceBasedDependencyInjection(t *testing.T) {
	// Create mock service
	mockUserService := &MockUserService{}
	
	// Create component with injected interface
	userManager := NewUserManager(mockUserService)
	
	// Setup expectations
	ctx := context.Background()
	userID := id.New[id.User]()
	expectedUser := &service.User{ID: userID}
	
	mockUserService.On("GetUser", ctx, userID).Return(expectedUser, nil)
	
	// Test the component
	user, err := userManager.GetUserInfo(ctx, userID)
	
	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockUserService.AssertExpectations(t)
}

// Example: HTTP Handler that accepts service interfaces
type UserHandler struct {
	services *factory.ServiceInterfaces
}

func NewUserHandler(services *factory.ServiceInterfaces) *UserHandler {
	return &UserHandler{services: services}
}

func (h *UserHandler) HandleGetUser(ctx context.Context, userID id.ID[id.User]) (*service.User, error) {
	return h.services.User.GetUser(ctx, userID)
}

// Test showing how to inject mock services into handlers
func TestHandlerWithMockServices(t *testing.T) {
	// Create mock services
	mockUserService := &MockUserService{}
	mockStoreService := &MockShopifyStoreService{}
	mockShopifyUserService := &MockShopifyUserService{}
	
	// Create service interfaces struct with mocks
	services := &factory.ServiceInterfaces{
		User:         mockUserService,
		ShopifyStore: mockStoreService,
		ShopifyUser:  mockShopifyUserService,
	}
	
	// Create handler with injected mocks
	handler := NewUserHandler(services)
	
	// Setup expectations
	ctx := context.Background()
	userID := id.New[id.User]()
	expectedUser := &service.User{ID: userID}
	
	mockUserService.On("GetUser", ctx, userID).Return(expectedUser, nil)
	
	// Test the handler
	user, err := handler.HandleGetUser(ctx, userID)
	
	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockUserService.AssertExpectations(t)
}

// Example: Testing that real services implement the interfaces
func TestServiceImplementsInterface(t *testing.T) {
	// This test ensures that our actual services implement the interfaces
	// If they don't, this will fail at compile time
	
	var userService factory.UserServiceInterface = &service.UserService{}
	var storeService factory.ShopifyStoreServiceInterface = &service.ShopifyStoreService{}
	var shopifyUserService factory.ShopifyUserServiceInterface = &service.ShopifyUserService{}
	
	// Just verify they're not nil (compiler already checked interface compliance)
	assert.NotNil(t, userService)
	assert.NotNil(t, storeService)
	assert.NotNil(t, shopifyUserService)
}

// Example: Production code using factory interfaces
func ExampleProductionUsage() {
	// In production, create real services via factory
	// services := factory.NewServiceInterfacesFromDB(database)
	
	// In tests, create mock services
	mockUserService := &MockUserService{}
	services := &factory.ServiceInterfaces{
		User: mockUserService,
		// ... other services
	}
	
	// Both work the same way due to interfaces
	handler := NewUserHandler(services)
	_ = handler
}
