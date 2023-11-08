package proto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID        uuid.UUID  `db:"id,omitempty,pk"       json:"id"`
	Email     string     `db:"email"                 json:"email"`
	Firstname string     `db:"firstname"             json:"firstname"`
	Lastname  string     `db:"lastname"              json:"lastname"`
	CreatedAt *time.Time `db:"created_at"            json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at, omitempty" json:"updated_at"`
}
