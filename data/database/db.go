package data

import (
	"errors"
	"fmt"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"

	"github.com/golang-cz/skeleton/config"
)

var DB *Database

type Database struct {
	Session db.Session

	User User
}

func NewDBSession(conf config.DBConfig) (db.Session, error) {
	if conf.Host == "" {
		return nil, errors.New("failed to connect to DB: no host")
	}

	connURL := postgresql.ConnectionURL{
		User:     conf.Username,
		Password: conf.Password,
		Host:     conf.Host,
		Database: conf.Database,
		Options: map[string]string{
			"application_name": conf.AppName,
		},
	}

	if conf.SSLMode != "" {
		connURL.Options["sslmode"] = conf.SSLMode
	}

	if conf.ConnectionTimeout > 0 {
		connURL.Options["connect_timeout"] = fmt.Sprintf("%d", conf.ConnectionTimeout)
	}

	DB, err := postgresql.Open(connURL)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to %v@%v/%v: %w",
			conf.Username,
			conf.Host,
			conf.Database,
			err,
		)
	}

	// var user User
	//
	// err = DB.Get(&user, db.Cond{"title": "The Shining"})
	// if err != nil {
	// 	log.Printf("User from DB")
	// }

	return DB, nil
}
