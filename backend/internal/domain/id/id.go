package id

import (
	"fmt"

	"github.com/MeViksry/quuid"
)

type ID = quuid.UUID

func New() (ID, error) {
	id, err := quuid.NewUUIDv7()
	if err != nil {
		return quuid.NilUUID, fmt.Errorf("id: generate uuidv7: %w", err)
	}
	return id, nil
}

func MustNew() ID {
	id, err := New()
	if err != nil {
		panic(err)
	}
	return id
}

func Parse(value string) (ID, error) {
	return quuid.ParseUUID(value)
}
