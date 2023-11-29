package proto

import (
	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID        uuid.UUID `db:"id,omitempty,pk" json:"id"`
	Email     string    `db:"email"           json:"email"`
	Firstname string    `db:"firstname"       json:"firstname"`
	Lastname  string    `db:"lastname"        json:"lastname"`
}
