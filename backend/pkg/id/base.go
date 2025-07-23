package id

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"
)

type ID[T Resource] string

func New[T Resource](v string) (ID[T], error) {
	var id ID[T]
	if err := id.initValues(v); err != nil {
		return id, err
	}
	return id, nil
}

func NewGeneration[T Resource]() ID[T] {
	var resource T

	val := fmt.Sprintf(
		"%s%s",
		resource.Prefix(),
		xid.New().String(),
	)

	return ID[T](val)
}

func (id *ID[T]) UnmarshalText(data []byte) error {
	return id.initValues(string(data))
}

// String will make a string of the ID
func (id ID[T]) String() string {
	return string(id)
}

func (b *ID[T]) MarshalJSON() ([]byte, error) { return json.Marshal(b.String()) }

// UnmarshalJSON will unmarshal the id into our id
func (b *ID[T]) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	return b.initValues(id)
}

// Value returns the ID + Prefix
func (b ID[T]) Value() (driver.Value, error) {
	return b.String(), nil
}

// Scan will scan in the db value
func (b *ID[T]) Scan(src interface{}) error {
	var id string
	switch t := src.(type) {
	case nil:
		return nil
	case string:
		id = t
	case []byte:
		if len(t) == 0 {
			return errors.New("no value for the ID")
		}
		id = string(t)
	default:
		return errors.New("incompatible type for ID")
	}

	return b.initValues(id)
}

func (b *ID[T]) initValues(id string) error {
	*b = ID[T](id)
	return b.validate()
}

// validate will ensure the ID is valid
func (b ID[T]) validate() error {
	var resource T
	if strings.HasPrefix(b.String(), resource.Prefix()) {
		return nil
	}
	return fmt.Errorf("Invalid ID prefix: %v, %v", b.String(), resource.Prefix())
}
