package data

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-cz/skeleton/config"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
)

type Database struct {
	Session db.Session

	User UserStore
}

func NewDBSession(conf config.DBConfig) (*Database, error) {
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

	dbSession, err := postgresql.Open(connURL)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to %v@%v/%v: %w",
			conf.Username,
			conf.Host,
			conf.Database,
			err,
		)
	}

	db.LC().SetLogger(log.Default())

	return &Database{Session: dbSession}, nil
}
