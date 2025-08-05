package id

type Resource interface{ Prefix() string }

type User struct {
	ID string
}

func (u User) Prefix() string {
	return "usr_"
}

type Store struct {
	ID string
}

func (s Store) Prefix() string {
	return "str_"
}

type ShopifyStore struct {
	ID string
}

func (s ShopifyStore) Prefix() string {
	return "sps_"
}

type ShopifyUser struct {
	ID string
}

func (s ShopifyUser) Prefix() string {
	return "spu_"
}

// Core Domain Types
type Product struct {
	ID string
}

func (p Product) Prefix() string {
	return "prd_"
}

type ProductVariant struct {
	ID string
}

func (p ProductVariant) Prefix() string {
	return "var_"
}

type InventoryItem struct {
	ID string
}

func (i InventoryItem) Prefix() string {
	return "inv_"
}

type Location struct {
	ID string
}

func (l Location) Prefix() string {
	return "loc_"
}

type Order struct {
	ID string
}

func (o Order) Prefix() string {
	return "ord_"
}

type OrderLineItem struct {
	ID string
}

func (o OrderLineItem) Prefix() string {
	return "oli_"
}

type InventoryLevel struct {
	ID string
}

func (i InventoryLevel) Prefix() string {
	return "ivl_"
}

// Platform Integration Types
type PlatformIntegration struct {
	ID string
}

func (p PlatformIntegration) Prefix() string {
	return "int_"
}

type SyncState struct {
	ID string
}

func (s SyncState) Prefix() string {
	return "syc_"
}
