package data

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/upper/db/v4"

	"github.com/golang-cz/skeleton/pkg/utc"
	"github.com/golang-cz/skeleton/proto"
)

type User struct {
	*proto.User

	CreatedAt time.Time  `json:"createdAt"           db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt"           db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

type UserStore struct {
	db.Collection
}

// Interface checks
var _ = interface {
	db.Record
	db.BeforeCreateHook
	db.BeforeUpdateHook
}(&User{})

var _ = interface {
	db.Store
}(&UserStore{})

func Users(sess db.Session) *UserStore {
	return &UserStore{sess.Collection("users")}
}

func (u *User) Store(sess db.Session) db.Store {
	return Users(sess)
}

func (u *User) BeforeCreate(sess db.Session) error {
	if err := u.Validate(); err != nil {
		return fmt.Errorf("user is not valid: %w", err)
	}

	u.CreatedAt = utc.Now()
	u.UpdatedAt = u.CreatedAt

	return nil
}

func (u *User) BeforeUpdate(sess db.Session) error {
	if err := u.Validate(); err != nil {
		return fmt.Errorf("user is not valid: %w", err)
	}

	u.UpdatedAt = utc.Now()

	return nil
}

func (u *User) Validate() error {
	return nil
}

func (s UserStore) Find(conds ...interface{}) db.Result {
	return s.Collection.Find(conds...)
}

func (s UserStore) FindActive(conds ...interface{}) db.Result {
	return s.Find(append([]interface{}{db.Cond{"deleted_at": db.IsNull()}}, conds...)...)
}

func (s UserStore) FindOne(conds ...interface{}) (user *User, err error) {
	if err = s.Find(conds...).One(&user); err != nil {
		return nil, fmt.Errorf("get first record: %w", err)
	}

	return user, nil
}

func (s UserStore) FindActiveOne(conds ...interface{}) (user *User, err error) {
	if err = s.FindActive(conds...).One(&user); err != nil {
		return nil, fmt.Errorf("get first record: %w", err)
	}

	return user, nil
}

func (s UserStore) FindById(id uuid.UUID, conds ...interface{}) (user *User, err error) {
	return s.FindOne(append([]interface{}{db.Cond{"id": id}}, conds...)...)
}

func (s UserStore) FindActiveById(id uuid.UUID, conds ...interface{}) (user *User, err error) {
	return s.FindActiveOne(append([]interface{}{db.Cond{"id": id}}, conds...)...)
}
