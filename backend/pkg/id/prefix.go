package id

type Resource interface{ Prefix() string }

type Store struct {
	ID string
}

func (s Store) Prefix() string {
	return "store"
}

type Product struct {
	ID string
}

func (p Product) Prefix() string {
	return "product"
}
