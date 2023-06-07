package api

import (
	"fmt"

	"golang.org/x/exp/slog"

	"github.com/golang-cz/skeleton/config"
	data "github.com/golang-cz/skeleton/data/database"
)

type API struct {
	Config    *config.AppConfig
	DbSession *data.Database
}

func New(conf *config.AppConfig) (*API, error) {
	// Database
	database, err := data.NewDBSession(conf.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}
	app := &API{Config: conf, DbSession: database}

	return app, nil
}

func (app *API) Close() {
	slog.Info("API: closing NATS & DB connections..")

	app.DbSession.Session.Close()
}
