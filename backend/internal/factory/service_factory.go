package factory

import (
	"github.com/ConradKurth/forecasting/backend/internal/db"
	"github.com/ConradKurth/forecasting/backend/internal/repository/core"
	"github.com/ConradKurth/forecasting/backend/internal/repository/shopify"
	"github.com/ConradKurth/forecasting/backend/internal/repository/users"
	"github.com/ConradKurth/forecasting/backend/internal/service"
)

// DatabaseConnection is an interface that both *db.DB and *db.TxDB implement
// This allows the factory to work with both regular and transactional connections
type DatabaseConnection interface {
	GetUsers() users.Querier
	GetShopify() shopify.Querier
	GetCore() core.Querier
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

// CreateCoreInventoryService creates a CoreInventoryService from the current database connection
func (f *ServiceFactory) CreateCoreInventoryService() *service.CoreInventoryService {
	return service.NewCoreInventoryService(f.dbConn.GetCore())
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

// CreateCoreService creates a CoreService from the current database connection
func (f *ServiceFactory) CreateCoreService() *service.CoreService {
	return service.NewCoreService(f.dbConn.GetCore())
}

// CreateCoreServiceInterface creates a CoreService and returns it as an interface
func (f *ServiceFactory) CreateCoreServiceInterface() CoreServiceInterface {
	return f.CreateCoreService()
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
		Core:         f.CreateCoreServiceInterface(),
	}
}

// ServiceInterfaceFactory provides methods to create service interfaces from different database connections
type ServiceInterfaceFactory struct {
	database *db.DB
}

// NewServiceInterfaceFactory creates a new ServiceInterfaceFactory
func NewServiceInterfaceFactory(database *db.DB) *ServiceInterfaceFactory {
	return &ServiceInterfaceFactory{
		database: database,
	}
}

// FromDB creates service interfaces from the regular database connection
func (sif *ServiceInterfaceFactory) FromDB() *ServiceInterfaces {
	factory := NewServiceFactory(sif.database)
	return factory.CreateAllServiceInterfaces()
}

// FromTx creates service interfaces from a transactional database connection
func (sif *ServiceInterfaceFactory) FromTx(txDB *db.TxDB) *ServiceInterfaces {
	factory := NewServiceFactory(txDB)
	return factory.CreateAllServiceInterfaces()
}

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
