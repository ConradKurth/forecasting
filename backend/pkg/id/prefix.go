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
