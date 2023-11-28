package data

import (
	"github.com/gofrs/uuid/v5"
	"github.com/upper/db/v4"

	"github.com/golang-cz/skeleton/proto"
)

type User struct {
	*proto.User
}

type UserStore struct {
	db.Collection
}

func Users(sess db.Session) *UserStore {
	return &UserStore{sess.Collection("users")}
}

func (us *UserStore) FindById(id uuid.UUID) (*User, error) {
	var user *User
	err := us.Find(db.Cond{"id": id}).One(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
