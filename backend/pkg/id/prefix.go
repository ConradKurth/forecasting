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

type Product struct {
	ID string
}

func (p Product) Prefix() string {
	return "prd_"
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
