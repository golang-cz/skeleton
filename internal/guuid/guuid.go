package guuid

import (
	"github.com/gofrs/uuid/v5"
)

func NewV7() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
