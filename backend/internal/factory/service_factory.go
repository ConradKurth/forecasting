package factory

import (
	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/ConradKurth/forecasting/backend/internal/service"
)

// DatabaseConnection is an interface that both *db.DB and *db.TxDB implement
// This allows the factory to work with both regular and transactional connections
type DatabaseConnection interface {
	GetUsers() users.Querier
	GetShopify() shopify.Querier
}

// ServiceFactory creates services from database connections
// It can work with both regular (*db.DB) and transactional (*db.TxDB) connections
type ServiceFactory struct {
	dbConn DatabaseConnection
}

// NewServiceFactory creates a new service factory with the given database connection
func NewServiceFactory(dbConn DatabaseConnection) *ServiceFactory {
	return &ServiceFactory{dbConn: dbConn}
}

// CreateUserService creates a UserService from the current database connection
func (f *ServiceFactory) CreateUserService() *service.UserService {
	return service.NewUserService(f.dbConn.GetUsers())
}

// CreateShopifyStoreService creates a ShopifyStoreService from the current database connection
func (f *ServiceFactory) CreateShopifyStoreService() *service.ShopifyStoreService {
	return service.NewShopifyStoreService(f.dbConn.GetShopify())
}

// CreateShopifyUserService creates a ShopifyUserService from the current database connection
func (f *ServiceFactory) CreateShopifyUserService() *service.ShopifyUserService {
	return service.NewShopifyUserService(f.dbConn.GetShopify())
}

// Interface-based service creation methods for dependency injection

// CreateUserServiceInterface creates a UserService and returns it as an interface
func (f *ServiceFactory) CreateUserServiceInterface() UserServiceInterface {
	return f.CreateUserService()
}

// CreateShopifyStoreServiceInterface creates a ShopifyStoreService and returns it as an interface
func (f *ServiceFactory) CreateShopifyStoreServiceInterface() ShopifyStoreServiceInterface {
	return f.CreateShopifyStoreService()
}

// CreateShopifyUserServiceInterface creates a ShopifyUserService and returns it as an interface
func (f *ServiceFactory) CreateShopifyUserServiceInterface() ShopifyUserServiceInterface {
	return f.CreateShopifyUserService()
}

// Services is a convenience struct that holds all service instances (concrete types)
type Services struct {
	User         *service.UserService
	ShopifyStore *service.ShopifyStoreService
	ShopifyUser  *service.ShopifyUserService
}

// CreateAllServices creates all services at once and returns them in a Services struct
func (f *ServiceFactory) CreateAllServices() *Services {
	return &Services{
		User:         f.CreateUserService(),
		ShopifyStore: f.CreateShopifyStoreService(),
		ShopifyUser:  f.CreateShopifyUserService(),
	}
}

// CreateAllServiceInterfaces creates all services and returns them as interfaces for dependency injection
func (f *ServiceFactory) CreateAllServiceInterfaces() *ServiceInterfaces {
	return &ServiceInterfaces{
		User:         f.CreateUserServiceInterface(),
		ShopifyStore: f.CreateShopifyStoreServiceInterface(),
		ShopifyUser:  f.CreateShopifyUserServiceInterface(),
	}
}

// Convenience functions for common patterns

// NewServicesFromDB creates services from a regular database connection
func NewServicesFromDB(database *db.DB) *Services {
	factory := NewServiceFactory(database)
	return factory.CreateAllServices()
}

// NewServicesFromTx creates services from a transactional database connection
func NewServicesFromTx(txDB *db.TxDB) *Services {
	factory := NewServiceFactory(txDB)
	return factory.CreateAllServices()
}

// Interface-based convenience functions for dependency injection

// NewServiceInterfacesFromDB creates service interfaces from a regular database connection
func NewServiceInterfacesFromDB(database *db.DB) *ServiceInterfaces {
	factory := NewServiceFactory(database)
	return factory.CreateAllServiceInterfaces()
}

// NewServiceInterfacesFromTx creates service interfaces from a transactional database connection
func NewServiceInterfacesFromTx(txDB *db.TxDB) *ServiceInterfaces {
	factory := NewServiceFactory(txDB)
	return factory.CreateAllServiceInterfaces()
}
