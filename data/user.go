package data

import (
	"time"

	"github.com/google/uuid"
	"github.com/upper/db/v4"
)

type User struct {
	ID         int64      `db:"id,omitempty,pk"       json:"id,string"`
	ExternalId uuid.UUID  `db:"external_id"`
	Email      string     `db:"email"`
	Username   string     `db:"username"`
	Firstname  string     `db:"firstname"`
	Lastname   string     `db:"lastname"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at, omitempty"`
}

func (user *User) Store(sess db.Session) db.Store {
	return sess.Collection("users")
}
